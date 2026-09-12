# wapi

[![Go Reference](https://pkg.go.dev/badge/github.com/wixet-limited/wapi.svg)](https://pkg.go.dev/github.com/wixet-limited/wapi)
[![CI](https://github.com/wixet-limited/wapi/actions/workflows/ci.yml/badge.svg)](https://github.com/wixet-limited/wapi/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wixet-limited/wapi)](https://goreportcard.com/report/github.com/wixet-limited/wapi)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**The plumbing every Go HTTP API rewrites, written once.**

What is left of a JSON API once you remove everything that knows about your
domain: problem responses, request validation, JWT verification, opaque
pagination cursors, structured logging, traces, graceful shutdown.

`wapi` is a set of small packages, not a framework. You import the two or
three you need and wire them yourself, in an order you can read.

```bash
go get github.com/wixet-limited/wapi
```

Requires Go 1.24+.

---

## What it looks like

No `wapi.New()`. No reflection. Your `main` stays the map of your service:

```go
ctx, stop := httpserve.Signals()
defer stop()

logger, _ := logging.New("info")
slog.SetDefault(logger)

pool, _ := pgpool.New(ctx, pgpool.Config{URL: os.Getenv("DATABASE_URL")})
defer pool.Close()

var handler http.Handler = myRouter()

// Read it top to bottom: that is the order a request travels.
handler = httpkit.OpenAPIValidation(spec, verifier)(handler)
handler = httpkit.Recovery(logger, handler)
handler = httpkit.CORS([]string{"https://app.example.com"})(handler)

server := &http.Server{Addr: ":8080", Handler: handler}

httpserve.Run(ctx, server, httpserve.Options{
    Logger:          logger,
    ShutdownTimeout: 10 * time.Second,
    StopSignals:     stop,
})
```

Every middleware is an `http.Handler` decorator. Nothing is registered behind
your back, so deleting a line deletes exactly one behaviour.

---

## Packages

| Package | What you get |
|---|---|
| [`httpkit`](httpkit) | RFC 7807 problem responses, panic recovery, CORS, health endpoints, OpenAPI request validation, bearer authentication |
| [`apperrors`](apperrors) | The error vocabulary a service returns — not found, conflict, validation, unauthorized, forbidden — mapped to status codes by `httpkit` |
| [`pagination`](pagination) | Opaque keyset cursors: base64url of a value you define, plus query fingerprints so a cursor cannot be replayed against a different filter |
| [`auth`](auth) | JWT verification against a JWKS endpoint, and the verified identity in the request context |
| [`httpserve`](httpserve) | Signal handling and graceful shutdown for an `*http.Server` |
| [`envconf`](envconf) | Reads single configuration values from the environment. Deliberately not a config framework |
| [`logging`](logging) | `slog` JSON handler that stamps every record with the active trace and span |
| [`observability`](observability) | OpenTelemetry traces and metrics setup, with one shutdown function |
| [`errorreporting`](errorreporting) | Sentry setup and middleware, correlated with OpenTelemetry trace IDs |
| [`pgpool`](pgpool) | An instrumented `pgxpool`, with sane connection limits |

Full API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/wixet-limited/wapi).

---

## Design

**A library calls nothing.** You call `httpkit.Recovery(logger, next)`. It
never calls you. There is no lifecycle to hook into, no interface you must
implement, no init order to learn. That is the whole design.

**No domain knowledge.** Nothing here knows what a user, an order or an
invoice is. The moment a package needs to, it belongs in your application
instead.

**The dependency list is public API.** `go.mod` is what every adopter
inherits, so it is short on purpose and each entry earns its place. Only the
packages you import cost you anything.

**Standard library shapes.** `http.Handler`, `context.Context`, `*slog.Logger`,
`error`. If you drop `wapi`, you are not rewriting your handlers.

---

## Versioning

Semantic versioning. Before `v1.0.0`, minor versions may break the API; the
changelog will say so plainly.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues before pull requests for
anything that adds a dependency.

## License

MIT — see [LICENSE](LICENSE).
