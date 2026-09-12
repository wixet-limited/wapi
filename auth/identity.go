package auth

import (
	"context"
)

type Identity struct {
	Email string
}

type contextKey struct{}

func WithIdentity(
	ctx context.Context,
	identity Identity,
) context.Context {
	return context.WithValue(
		ctx,
		contextKey{},
		identity,
	)
}

func MustIdentity(
	ctx context.Context,
) Identity {
	identity, ok := ctx.Value(
		contextKey{},
	).(Identity)

	if !ok || identity.Email == "" {
		panic(
			"authenticated identity missing from context",
		)
	}

	return identity
}
