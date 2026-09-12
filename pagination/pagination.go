package pagination

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// Encode serializes a feature-specific cursor as JSON
// and returns it encoded as Base64URL.
//
// The cursor remains opaque to API clients.
func Encode[T any](
	value T,
) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf(
			"marshal cursor: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(
		data,
	), nil
}

// Decode decodes a Base64URL cursor and deserializes
// its feature-specific JSON representation.
func Decode[T any](
	value string,
) (T, error) {
	var result T

	value = strings.TrimSpace(value)

	if value == "" {
		return result, fmt.Errorf(
			"cursor is empty",
		)
	}

	data, err :=
		base64.RawURLEncoding.DecodeString(
			value,
		)
	if err != nil {
		return result, fmt.Errorf(
			"decode cursor: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		data,
		&result,
	); err != nil {
		return result, fmt.Errorf(
			"unmarshal cursor: %w",
			err,
		)
	}

	return result, nil
}

// QueryHash creates a deterministic fingerprint for
// the query represented by the provided parts.
//
// Each feature decides which values participate in
// the fingerprint and how they are normalized.
//
// The null separator avoids ambiguous concatenations:
//
// ["ab", "c"] != ["a", "bc"]
func QueryHash(
	parts ...string,
) string {
	canonical := strings.Join(
		parts,
		"\x00",
	)

	sum := sha256.Sum256(
		[]byte(canonical),
	)

	return base64.RawURLEncoding.EncodeToString(
		sum[:],
	)
}
