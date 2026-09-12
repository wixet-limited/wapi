package errorreporting

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	DSN         string
	Environment string
	Release     string
}

type Reporter struct {
	enabled bool
}

func Setup(cfg Config) (*Reporter, error) {
	if cfg.DSN == "" {
		return &Reporter{}, nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		AttachStacktrace: true,
		EnableTracing:    false,
		SendDefaultPII:   false,
	}); err != nil {
		return nil, fmt.Errorf("initialize Sentry: %w", err)
	}

	return &Reporter{
		enabled: true,
	}, nil
}

func (r *Reporter) Enabled() bool {
	return r.enabled
}

func (r *Reporter) CaptureException(
	ctx context.Context,
	err error,
) {
	if !r.enabled || err == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.WithScope(func(scope *sentry.Scope) {
		spanContext := trace.SpanContextFromContext(ctx)

		if spanContext.IsValid() {
			scope.SetTag(
				"otel.trace_id",
				spanContext.TraceID().String(),
			)

			scope.SetTag(
				"otel.span_id",
				spanContext.SpanID().String(),
			)
		}

		hub.CaptureException(err)
	})
}

func (r *Reporter) WrapHTTP(
	handler http.Handler,
) http.Handler {
	if !r.enabled {
		return handler
	}

	middleware := sentryhttp.New(
		sentryhttp.Options{
			Repanic:         true,
			WaitForDelivery: false,
		},
	)

	return middleware.Handle(handler)
}

func (r *Reporter) Shutdown(
	ctx context.Context,
) error {
	if !r.enabled {
		return nil
	}

	flushed := sentry.FlushWithContext(ctx)

	if client := sentry.CurrentHub().Client(); client != nil {
		client.Close()
	}

	if !flushed {
		return fmt.Errorf(
			"Sentry flush did not complete before timeout",
		)
	}

	return nil
}
