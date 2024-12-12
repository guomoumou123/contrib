package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	otelmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	DefaultMeterName = "default-meter"
)

type (
	MeterProvider = metric.MeterProvider
	MeterOption   = metric.MeterOption
	Meter         = metric.Meter

	PrometheusMetric interface {
		Meter
		Register() prometheus.Registerer
		Shutdown(ctx context.Context) error
	}
	prometheusMetric struct {
		Meter
		exporter      *otelprometheus.Exporter
		meterProvider *otelmetric.MeterProvider
		register      prometheus.Registerer
	}
)

func NewPrometheusMetric(config Config) (PrometheusMetric, error) {
	register := config.GetRegisterer()
	gatherer := config.GetGatherer()
	exporter, err := otelprometheus.New(
		otelprometheus.WithRegisterer(register),
	)
	if err != nil {
		return nil, err
	}
	res := resource.Default()
	if svcName := config.ServiceName; svcName != "" {
		res, _ = resource.Merge(res, resource.NewWithAttributes(res.SchemaURL(), semconv.ServiceName(svcName)))
	}
	provider := otelmetric.NewMeterProvider(otelmetric.WithReader(exporter), otelmetric.WithResource(res))
	otel.SetMeterProvider(provider)
	if port := config.Port; port > 0 {
		listenMetricServer(port, register, gatherer)
	}

	return &prometheusMetric{
		Meter:         provider.Meter(DefaultMeterName),
		meterProvider: provider,
		register:      register,
		exporter:      exporter,
	}, nil
}

func (p *prometheusMetric) Register() prometheus.Registerer {
	return p.register
}

func (p *prometheusMetric) Shutdown(ctx context.Context) error {
	return p.meterProvider.Shutdown(ctx)
}

func listenMetricServer(port int, registerer prometheus.Registerer, gatherer prometheus.Gatherer) {
	addr := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.InstrumentMetricHandler(registerer, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("prometheus metric server panic: %v \n", r)
			}
		}()
		fmt.Println("prometheus metric server listen on: " + addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			fmt.Println("prometheus metric server start failed: " + err.Error())
		}
	}()
}
