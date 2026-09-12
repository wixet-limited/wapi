// Package httpserve runs an *http.Server until the process is asked to stop,
// then shuts it down gracefully.
//
// It owns no routing, no middleware and no configuration: the caller builds
// the server it wants and hands it over.
package httpserve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Signals returns a context that is cancelled on SIGINT or SIGTERM, together
// with the function that releases the handlers.
func Signals() (
	context.Context,
	context.CancelFunc,
) {
	return signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
}

type Options struct {
	// Logger records the server lifecycle. Required.
	Logger *slog.Logger

	// ShutdownTimeout bounds the graceful shutdown.
	ShutdownTimeout time.Duration

	// StopSignals, when set, is called as shutdown begins so that a second
	// interrupt terminates the process instead of being swallowed by the
	// shutdown that is already in flight.
	//
	// Pass the CancelFunc returned by Signals.
	StopSignals func()
}

// Run starts server and blocks until ctx is cancelled or ListenAndServe
// fails.
//
// A server closed through Shutdown is not an error.
func Run(
	ctx context.Context,
	server *http.Server,
	opts Options,
) error {
	serverErr :=
		make(
			chan error,
			1,
		)

	go func() {
		opts.Logger.Info(
			"HTTP server started",
			"addr",
			server.Addr,
		)

		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil &&
			!errors.Is(
				err,
				http.ErrServerClosed,
			) {
			return fmt.Errorf(
				"HTTP server: %w",
				err,
			)
		}

		return nil

	case <-ctx.Done():
		if opts.StopSignals != nil {
			opts.StopSignals()
		}
	}

	opts.Logger.Info(
		"shutting down HTTP server",
	)

	shutdownCtx, cancel :=
		context.WithTimeout(
			context.Background(),
			opts.ShutdownTimeout,
		)
	defer cancel()

	if err :=
		server.Shutdown(
			shutdownCtx,
		); err != nil {
		return fmt.Errorf(
			"shutdown HTTP server: %w",
			err,
		)
	}

	opts.Logger.Info(
		"HTTP server stopped",
	)

	return nil
}
