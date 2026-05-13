package bootstrap

import (
	"context"
	"errors"
	"goph-profile/internal/config"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

type Telemetry struct {
	Logger *log.LoggerProvider
	Tracer *trace.TracerProvider
	Metric *metric.MeterProvider
}

func NewTelemetryStack(ctx context.Context, cfg config.App) (Telemetry, error) {
	res, err := newResource(ctx, cfg.Name)
	if err != nil {
		return Telemetry{}, err
	}

	meterExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return Telemetry{}, err
	}

	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return Telemetry{}, err
	}

	logExporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return Telemetry{}, err
	}

	meterProvider := newMeterProvider(res, meterExporter)
	tracerProvider := newTracerProvider(res, traceExporter)
	loggerProvider := newLoggerProvider(res, logExporter)

	otel.SetMeterProvider(meterProvider)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
		),
	)

	global.SetLoggerProvider(loggerProvider)

	return Telemetry{
		Logger: loggerProvider,
		Tracer: tracerProvider,
		Metric: meterProvider,
	}, nil
}

func (t Telemetry) Shutdown(ctx context.Context) error {
	var errs []error

	if err := t.Logger.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	if err := t.Tracer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	if err := t.Metric.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func newResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
}

func newTracerProvider(res *resource.Resource, exporter trace.SpanExporter) *trace.TracerProvider {
	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
}

func newMeterProvider(res *resource.Resource, exporter metric.Exporter) *metric.MeterProvider {
	return metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(3*time.Second))),
	)
}

func newLoggerProvider(res *resource.Resource, exporter log.Exporter) *log.LoggerProvider {
	return log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)
}
