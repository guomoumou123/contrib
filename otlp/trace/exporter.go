package trace

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
)

func newHttpExporter(ctx context.Context, config *Config) (oteltrace.SpanExporter, error) {
	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpointURL(config.EndpointUrl),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newStdoutExporter(ctx context.Context) (oteltrace.SpanExporter, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
