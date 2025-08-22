package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor creates a gRPC unary server interceptor for observability
func UnaryServerInterceptor(serviceName string, logger *zap.Logger) grpc.UnaryServerInterceptor {
	tracer := otel.Tracer(serviceName)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Start tracing span
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.service", serviceName),
				attribute.String("rpc.method", info.FullMethod),
			),
		)
		defer span.End()

		// Add trace context to logger
		contextLogger := LoggerWithTraceContext(ctx, logger)
		contextLogger.Info("gRPC request started", zap.String("method", info.FullMethod))

		// Call the handler
		resp, err := handler(ctx, req)

		// Record metrics and tracing
		duration := time.Since(start)
		statusCode := "success"
		if err != nil {
			statusCode = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			contextLogger.Error("gRPC request failed", 
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.Duration("duration", duration))
		} else {
			span.SetStatus(codes.Ok, "")
			contextLogger.Info("gRPC request completed", 
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration))
		}

		// Record Prometheus metrics
		RequestsTotal.WithLabelValues(serviceName, info.FullMethod, statusCode).Inc()
		RequestDuration.WithLabelValues(serviceName, info.FullMethod).Observe(duration.Seconds())

		return resp, err
	}
}

// UnaryClientInterceptor creates a gRPC unary client interceptor for observability
func UnaryClientInterceptor(serviceName string, logger *zap.Logger) grpc.UnaryClientInterceptor {
	tracer := otel.Tracer(serviceName)

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		// Start tracing span
		ctx, span := tracer.Start(ctx, method,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.service", serviceName),
				attribute.String("rpc.method", method),
				attribute.String("rpc.target", cc.Target()),
			),
		)
		defer span.End()

		// Add trace context to logger
		contextLogger := LoggerWithTraceContext(ctx, logger)
		contextLogger.Debug("gRPC client request started", zap.String("method", method))

		// Call the invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Record metrics and tracing
		duration := time.Since(start)
		statusCode := "success"
		if err != nil {
			statusCode = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			contextLogger.Error("gRPC client request failed", 
				zap.String("method", method),
				zap.Error(err),
				zap.Duration("duration", duration))
		} else {
			span.SetStatus(codes.Ok, "")
			contextLogger.Debug("gRPC client request completed", 
				zap.String("method", method),
				zap.Duration("duration", duration))
		}

		// Record Prometheus metrics
		RequestsTotal.WithLabelValues(serviceName+"-client", method, statusCode).Inc()
		RequestDuration.WithLabelValues(serviceName+"-client", method).Observe(duration.Seconds())

		return err
	}
}

// StreamServerInterceptor creates a gRPC stream server interceptor for observability
func StreamServerInterceptor(serviceName string, logger *zap.Logger) grpc.StreamServerInterceptor {
	tracer := otel.Tracer(serviceName)

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Start tracing span
		ctx, span := tracer.Start(ss.Context(), info.FullMethod,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.service", serviceName),
				attribute.String("rpc.method", info.FullMethod),
				attribute.Bool("rpc.streaming", true),
			),
		)
		defer span.End()

		// Add trace context to logger
		contextLogger := LoggerWithTraceContext(ctx, logger)
		contextLogger.Info("gRPC stream started", zap.String("method", info.FullMethod))

		// Call the handler
		err := handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})

		// Record metrics and tracing
		duration := time.Since(start)
		statusCode := "success"
		if err != nil {
			statusCode = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			contextLogger.Error("gRPC stream failed", 
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.Duration("duration", duration))
		} else {
			span.SetStatus(codes.Ok, "")
			contextLogger.Info("gRPC stream completed", 
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration))
		}

		// Record Prometheus metrics
		RequestsTotal.WithLabelValues(serviceName, info.FullMethod, statusCode).Inc()
		RequestDuration.WithLabelValues(serviceName, info.FullMethod).Observe(duration.Seconds())

		return err
	}
}

// wrappedServerStream wraps grpc.ServerStream to inject context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

