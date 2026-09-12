// Package envconf reads configuration values from the process environment.
//
// It deliberately stops at parsing single values: which variables an
// application reads, their defaults and how they relate to each other is a
// decision of that application, not of this package.
package envconf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Value returns the trimmed value of name, or the empty string when it is
// unset or blank.
func Value(
	name string,
) string {
	return strings.TrimSpace(
		os.Getenv(name),
	)
}

// String returns the trimmed value of name, or fallback when it is unset or
// blank.
func String(
	name string,
	fallback string,
) string {
	value := Value(name)

	if value == "" {
		return fallback
	}

	return value
}

// Duration parses name as a Go duration, or returns fallback when it is unset
// or blank.
func Duration(
	name string,
	fallback time.Duration,
) (time.Duration, error) {
	value := Value(name)

	if value == "" {
		return fallback, nil
	}

	result, err :=
		time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %w",
			name,
			err,
		)
	}

	return result, nil
}

// Int32 parses name as a 32 bit integer, or returns fallback when it is unset
// or blank.
func Int32(
	name string,
	fallback int32,
) (int32, error) {
	value := Value(name)

	if value == "" {
		return fallback, nil
	}

	result, err :=
		strconv.ParseInt(
			value,
			10,
			32,
		)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %w",
			name,
			err,
		)
	}

	return int32(result), nil
}

// CSV splits a comma separated value, trimming each entry and dropping the
// empty ones.
func CSV(
	value string,
) []string {
	parts :=
		strings.Split(
			value,
			",",
		)

	result :=
		make(
			[]string,
			0,
			len(parts),
		)

	for _, part := range parts {
		part =
			strings.TrimSpace(part)

		if part != "" {
			result =
				append(
					result,
					part,
				)
		}
	}

	return result
}
