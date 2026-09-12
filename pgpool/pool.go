package pgpool

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	URL string

	MaxConns int32
	MinConns int32

	MaxConnIdleTime time.Duration
	MaxConnLifetime time.Duration
}

// New creates an instrumented connection pool and verifies it can reach
// the database before returning.
func New(
	ctx context.Context,
	cfg Config,
) (*pgxpool.Pool, error) {
	poolConfig, err :=
		pgxpool.ParseConfig(cfg.URL)

	if err != nil {
		return nil, fmt.Errorf(
			"parse PostgreSQL configuration: %w",
			err,
		)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	poolConfig.MaxConnIdleTime =
		cfg.MaxConnIdleTime

	poolConfig.MaxConnLifetime =
		cfg.MaxConnLifetime

	poolConfig.ConnConfig.Tracer =
		otelpgx.NewTracer()

	pool, err :=
		pgxpool.NewWithConfig(
			ctx,
			poolConfig,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"create PostgreSQL pool: %w",
			err,
		)
	}

	if err := otelpgx.RecordStats(pool); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"register PostgreSQL pool metrics: %w",
			err,
		)
	}

	return pool, nil
}
