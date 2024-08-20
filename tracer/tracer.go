package tracer

import (
	"context"

	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Custom function to add trace/span IDs to the logger
func WithTrace(ctx context.Context, logger *zap.Logger) *zap.Logger {
	span, found := tracer.SpanFromContext(ctx) // Assuming tracer is from some tracing library
	if !found {
		return logger
	}

	return logger.With(
		zap.Uint64(ext.LogKeyTraceID, span.Context().TraceID()),
		zap.Uint64(ext.LogKeySpanID, span.Context().SpanID()),
	)
}
