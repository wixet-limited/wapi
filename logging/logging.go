package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

func New(levelText string) (*slog.Logger, error) {
	var level slog.Level

	if err := level.UnmarshalText(
		[]byte(strings.ToUpper(levelText)),
	); err != nil {
		return nil, fmt.Errorf(
			"invalid log level %q: %w",
			levelText,
			err,
		)
	}

	base := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: level,
		},
	)

	return slog.New(
		&traceHandler{
			next: base,
		},
	), nil
}

type traceHandler struct {
	next slog.Handler
}

func (h *traceHandler) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return h.next.Enabled(ctx, level)
}

func (h *traceHandler) Handle(
	ctx context.Context,
	record slog.Record,
) error {
	spanContext := trace.SpanContextFromContext(ctx)

	if spanContext.IsValid() {
		record.AddAttrs(
			slog.String(
				"trace_id",
				spanContext.TraceID().String(),
			),
			slog.String(
				"span_id",
				spanContext.SpanID().String(),
			),
		)
	}

	return h.next.Handle(ctx, record)
}

func (h *traceHandler) WithAttrs(
	attrs []slog.Attr,
) slog.Handler {
	return &traceHandler{
		next: h.next.WithAttrs(attrs),
	}
}

func (h *traceHandler) WithGroup(
	name string,
) slog.Handler {
	return &traceHandler{
		next: h.next.WithGroup(name),
	}
}
