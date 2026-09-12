package httpkit

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			defer func() {
				recovered := recover()

				if recovered == nil {
					return
				}

				logger.ErrorContext(
					r.Context(),
					"panic recovered",
					"method",
					r.Method,
					"path",
					r.URL.Path,
					"panic",
					recovered,
					"stack",
					string(debug.Stack()),
				)

				WriteProblem(
					w,
					http.StatusInternalServerError,
					"Internal server error",
					"An unexpected error occurred.",
				)
			}()

			next.ServeHTTP(w, r)
		},
	)
}
