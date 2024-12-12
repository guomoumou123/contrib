package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type (
	TracerProvider = trace.TracerProvider
	Tracer         = trace.Tracer
	OTLPTracer     interface {
		Tracer
		Shutdown(ctx context.Context) error
	}

	closeableTracer struct {
		Tracer
		provider *oteltrace.TracerProvider
	}
)

type Config struct {
	ServiceName string
	EndpointUrl string
	SamplerRate float64 //采样比例值
}

var DefaultTracerName = "default"

func newCloseableTracer(config *Config, exporter oteltrace.SpanExporter) closeableTracer {
	r, _ := resource.Merge(resource.Default(), resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(config.ServiceName)))

	tp := oteltrace.NewTracerProvider(
		oteltrace.WithBatcher(exporter),
		oteltrace.WithResource(r),
		oteltrace.WithSampler(oteltrace.ParentBased(oteltrace.TraceIDRatioBased(config.SamplerRate))),
	)
	otel.SetTracerProvider(tp)
	ct := closeableTracer{
		provider: tp,
		Tracer:   tp.Tracer(DefaultTracerName),
	}
	return ct
}

func (c closeableTracer) Shutdown(ctx context.Context) error {
	return c.provider.Shutdown(ctx)
}

func NewOTLPTracer(ctx context.Context, config *Config) (OTLPTracer, error) {
	exporter, err := newHttpExporter(ctx, config)
	if err != nil {
		return nil, err
	}
	tracer := newCloseableTracer(config, exporter)
	return tracer, nil
}

func NewStdOutOTLPTracer(ctx context.Context, config *Config) (OTLPTracer, error) {
	exporter, err := newStdoutExporter(ctx)
	if err != nil {
		return nil, err
	}
	tracer := newCloseableTracer(config, exporter)
	return tracer, nil
}
