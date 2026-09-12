package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
}

func Setup(
	ctx context.Context,
	cfg Config,
) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error

	shutdown := func(ctx context.Context) error {
		var err error

		for i := len(shutdownFuncs) - 1; i >= 0; i-- {
			err = errors.Join(
				err,
				shutdownFuncs[i](ctx),
			)
		}

		shutdownFuncs = nil

		return err
	}

	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithAttributes(
			attribute.String(
				"service.name",
				cfg.ServiceName,
			),
			attribute.String(
				"service.version",
				cfg.ServiceVersion,
			),
		),
	)
	if err != nil {
		return shutdown, fmt.Errorf(
			"create OpenTelemetry resource: %w",
			err,
		)
	}

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	otel.SetErrorHandler(
		otel.ErrorHandlerFunc(func(err error) {
			slog.Default().Error(
				"OpenTelemetry error",
				"error",
				err,
			)
		}),
	)

	spanExporter, err :=
		autoexport.NewSpanExporter(ctx)

	if err != nil {
		return shutdown, fmt.Errorf(
			"create OpenTelemetry trace exporter: %w",
			err,
		)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(spanExporter),
		trace.WithResource(res),
	)

	shutdownFuncs = append(
		shutdownFuncs,
		tracerProvider.Shutdown,
	)

	otel.SetTracerProvider(tracerProvider)

	metricReader, err :=
		autoexport.NewMetricReader(ctx)

	if err != nil {
		shutdownErr := shutdown(ctx)

		return shutdown, errors.Join(
			fmt.Errorf(
				"create OpenTelemetry metric reader: %w",
				err,
			),
			shutdownErr,
		)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metricReader),
		metric.WithResource(res),
	)

	shutdownFuncs = append(
		shutdownFuncs,
		meterProvider.Shutdown,
	)

	otel.SetMeterProvider(meterProvider)

	return shutdown, nil
}
