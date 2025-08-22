package observability

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new structured logger with observability context
func NewLogger(serviceName string, level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	
	// Add service name to all log entries
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// LoggerWithTraceContext adds trace context to logger
func LoggerWithTraceContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return logger
	}

	spanContext := span.SpanContext()
	return logger.With(
		zap.String("trace_id", spanContext.TraceID().String()),
		zap.String("span_id", spanContext.SpanID().String()),
	)
}

// LoggerWithRequestID adds request ID to logger
func LoggerWithRequestID(logger *zap.Logger, requestID string) *zap.Logger {
	return logger.With(zap.String("request_id", requestID))
}

// LoggerWithOrderID adds order ID to logger for business context
func LoggerWithOrderID(logger *zap.Logger, orderID string) *zap.Logger {
	return logger.With(zap.String("order_id", orderID))
}

// LoggerWithCustomerID adds customer ID to logger for business context
func LoggerWithCustomerID(logger *zap.Logger, customerID string) *zap.Logger {
	return logger.With(zap.String("customer_id", customerID))
}

