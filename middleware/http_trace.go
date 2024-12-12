package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.seakoi.net/root/contrib/log"
	innermetrics "gitlab.seakoi.net/root/contrib/otlp/metrics"
	innertrace "gitlab.seakoi.net/root/contrib/otlp/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var DefaultTelemetryBucketBoundaries = []float64{
	100,
	200,
	300,
	500,
	700,
	float64(time.Second.Milliseconds() * 1),
	float64(time.Second.Milliseconds() * 2),
	float64(time.Second.Milliseconds() * 5),
}

var noPrintBodyHeader = map[string]string{
	"multipart/form-data":      "multipart/form-data",
	"application/octet-stream": "application/octet-stream",
}

type httpResponseWriter struct {
	gin.ResponseWriter
	serviceName string
	env         string
	context     context.Context
	logger      log.LogCore
	startTime   time.Time
	request     *http.Request
	span        trace.Span
}

func TelemetryTrace(serviceName, env string, logger log.LogCore) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := ctx.Request
		mctx := req.Context()
		requestBodySize := req.Header.Get("Content-Length")
		body, err := ctx.GetRawData()
		if err != nil {
			logger.ErrorWithCtx(mctx, "[Logger Middleware]", log.Any("错误信息", err.Error()))
			ctx.Abort()
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		optionshttp := []otelhttptrace.Option{
			otelhttptrace.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{},
				b3.New(b3.WithInjectEncoding(b3.B3SingleHeader|b3.B3MultipleHeader)))),
		}
		_, _, sctx := otelhttptrace.Extract(mctx, req, optionshttp...)

		spanCtx, span := otel.GetTracerProvider().Tracer(innertrace.DefaultTracerName).Start(trace.ContextWithRemoteSpanContext(mctx, sctx), requestMethodPath(ctx))
		defer span.End()
		span.SetAttributes(
			semconv.NetProtocolVersion(req.Proto),
			semconv.HTTPRoute(req.URL.Path),
			semconv.DeploymentEnvironment(env),
			semconv.HTTPRequestMethodKey.String(req.Method),
			semconv.ServiceName(serviceName),
			attribute.String(ATTR_PARAMS, req.URL.RawQuery),
			attribute.String(ATTR_REQUEST_BODY, string(body)),
			attribute.String(ATTR_REQUEST_BODY_SIZE, requestBodySize),
			attribute.String(ATTR_CLIENT_IP, ctx.ClientIP()),
			attribute.String("trace_id", span.SpanContext().TraceID().String()),
		)

		now := time.Now()
		ctx.Request = ctx.Request.Clone(spanCtx)
		ctx.Writer = httpResponseWriter{
			ResponseWriter: ctx.Writer,
			serviceName:    serviceName,
			env:            env,
			context:        spanCtx,
			startTime:      now,
			logger:         logger,
			span:           span,
			request:        ctx.Request,
		}
		requestLog(ctx, body, now, logger)
		ctx.Next()

	}
}

func requestMethodPath(ctx *gin.Context) string {
	path := ctx.Request.URL.Path
	if matchedTemplatePath := ctx.FullPath(); matchedTemplatePath != "" {
		path = matchedTemplatePath
	}
	return path
}

func requestLog(ctx *gin.Context, body []byte, start time.Time, logger log.LogCore) {
	req := ctx.Request
	mctx := req.Context()
	for _, v := range noPrintBodyHeader {
		if ctx.Request.Header.Get("Content-Type") == v {
			body = nil
			break
		}
	}

	logger.InfoWithCtx(
		mctx,
		"http.request",
		log.Any("client_ip", ctx.ClientIP()),
		log.Any("method", req.Method),
		log.Any("path", req.URL.Path),
		log.Any("params", req.URL.RawQuery),
		log.Any("header", req.Header),
		log.Any("body", string(body)),
		log.Any("body_size", len(body)),
		log.Any("time", start),
	)
}

func (w httpResponseWriter) Write(b []byte) (int, error) {
	resp := w.ResponseWriter
	n, err := w.ResponseWriter.Write(b)
	w.logger.InfoWithCtx(
		w.context,
		"http.response",
		log.Any("path", w.request.URL.Path),
		log.Any("body", string(b)),
		log.Any("status", resp.Status()),
		log.Any("response_content_length", len(b)),
		log.Any("latency", time.Since(w.startTime)),
	)
	request := w.request
	path := request.URL.Path
	meter := otel.Meter(innermetrics.DefaultMeterName, metric.WithInstrumentationVersion(sdk.Version()))
	attrs := []attribute.KeyValue{
		semconv.ServiceName(w.serviceName),
		semconv.HTTPRoute(path),
		semconv.HTTPRequestMethodKey.String(request.Method),
		semconv.DeploymentEnvironment(w.env),
		semconv.HTTPResponseStatusCode(w.Status()),
	}

	jagerAttr := append([]attribute.KeyValue{
		semconv.HTTPResponseBodySize(len(b)),
	}, attrs...)
	defer func(startTime time.Time) {
		metricName := "web"

		w.span.SetAttributes(jagerAttr...)

		if counter, err := meter.Int64Counter(metricName + ".request"); err == nil {
			counter.Add(w.context, 1, metric.WithAttributes(attrs...))
		}

		if histogram, err := meter.Int64Histogram(metricName+".histogram",
			metric.WithExplicitBucketBoundaries(DefaultTelemetryBucketBoundaries...)); err == nil {
			histogram.Record(w.context, time.Since(startTime).Milliseconds(),
				metric.WithAttributes(attrs...))
		}
	}(w.startTime)
	return n, err
}
