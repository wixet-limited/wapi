package apperrors

import "fmt"

type NotFound struct {
	Resource string
	ID       string
}

func (e NotFound) Error() string {
	if e.ID == "" {
		return fmt.Sprintf("%s not found", e.Resource)
	}

	return fmt.Sprintf("%s %s not found", e.Resource, e.ID)
}

type Conflict struct {
	Resource string
	Detail   string
}

func (e Conflict) Error() string {
	if e.Detail != "" {
		return e.Detail
	}

	return fmt.Sprintf("%s conflicts with the current state", e.Resource)
}

type Validation struct {
	Detail string
}

func (e Validation) Error() string {
	return e.Detail
}

type Unauthorized struct{}

func (Unauthorized) Error() string {
	return "authentication required"
}

type Forbidden struct{}

func (Forbidden) Error() string {
	return "operation not permitted"
}
