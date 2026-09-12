package httpkit

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Pinger interface {
	Ping(context.Context) error
}

type Health struct {
	database Pinger
}

func NewHealth(
	database Pinger,
) *Health {
	return &Health{
		database: database,
	}
}

func (h *Health) Live(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeHealth(
		w,
		http.StatusOK,
		"ok",
	)
}

func (h *Health) Ready(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		time.Second,
	)
	defer cancel()

	if err := h.database.Ping(ctx); err != nil {
		writeHealth(
			w,
			http.StatusServiceUnavailable,
			"unavailable",
		)

		return
	}

	writeHealth(
		w,
		http.StatusOK,
		"ok",
	)
}

func writeHealth(
	w http.ResponseWriter,
	statusCode int,
	status string,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.Header().Set(
		"Cache-Control",
		"no-store",
	)

	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"status": status,
		},
	)
}
