package httpkit

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/wixet-limited/wapi/apperrors"
)

type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type ErrorReporter interface {
	CaptureException(
		ctx context.Context,
		err error,
	)
}

func WriteProblem(
	w http.ResponseWriter,
	status int,
	title string,
	detail string,
) {
	w.Header().Set(
		"Content-Type",
		"application/problem+json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(
		Problem{
			Type:   "about:blank",
			Title:  title,
			Status: status,
			Detail: detail,
		},
	)
}

func RequestErrorHandler(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	WriteProblem(
		w,
		http.StatusBadRequest,
		"Invalid request",
		err.Error(),
	)
}

func NewResponseErrorHandler(
	reporter ErrorReporter,
) func(
	http.ResponseWriter,
	*http.Request,
	error,
) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {
		var notFound apperrors.NotFound

		if errors.As(err, &notFound) {
			WriteProblem(
				w,
				http.StatusNotFound,
				"Resource not found",
				notFound.Error(),
			)

			return
		}

		var validation apperrors.Validation

		if errors.As(err, &validation) {
			WriteProblem(
				w,
				http.StatusBadRequest,
				"Invalid request",
				validation.Error(),
			)

			return
		}

		var unauthorized apperrors.Unauthorized

		if errors.As(err, &unauthorized) {
			WriteProblem(
				w,
				http.StatusUnauthorized,
				"Unauthorized",
				unauthorized.Error(),
			)

			return
		}

		var forbidden apperrors.Forbidden

		if errors.As(err, &forbidden) {
			WriteProblem(
				w,
				http.StatusForbidden,
				"Forbidden",
				forbidden.Error(),
			)

			return
		}

		var conflict apperrors.Conflict

		if errors.As(err, &conflict) {
			WriteProblem(
				w,
				http.StatusConflict,
				"Conflict",
				conflict.Error(),
			)

			return
		}

		//
		// Unexpected error.
		//
		// This is the single place where normal errors
		// resulting in HTTP 500 are logged.
		//

		slog.ErrorContext(
			r.Context(),
			"unexpected request error",
			"error",
			err,
			"method",
			r.Method,
			"path",
			r.URL.Path,
			"stack",
			string(debug.Stack()),
		)

		//
		// Send only unexpected errors to Sentry.
		//

		reporter.CaptureException(
			r.Context(),
			err,
		)

		//
		// Never expose internal details to the client.
		//

		WriteProblem(
			w,
			http.StatusInternalServerError,
			"Internal server error",
			"An unexpected error occurred.",
		)
	}
}
