package context

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type (
	contextKey   int
	TraceContext struct {
		TraceId string
	}
)

const TraceContextKey contextKey = 1

func GetTraceId(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	sctx := span.SpanContext()
	if sctx.IsValid() {
		return span.SpanContext().TraceID().String()
	}
	if traceSpan := ctx.Value(TraceContextKey); traceSpan != nil {
		t, ok := traceSpan.(TraceContext)
		if !ok {
			return ""
		}
		return t.TraceId
	}

	return ""
}
