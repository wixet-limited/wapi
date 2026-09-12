package httpkit

// Responder binds the response types of one generated operation, so handlers
// can feed a service result straight in and stay a single line.
//
// R is the operation's response interface (e.g. usersapi.GetUserResponseObject)
// and T the value the service returns. Each operation gets its own concrete
// 404 type from the generator, which is why NotFound is a field rather than
// something a plain function could return.
type Responder[R, T any] struct {
	// OK builds the success response from the service value.
	OK func(T) R

	// NotFound is the response for a service result that reports no match.
	// Leave it zero for operations whose spec declares no 404.
	NotFound R
}

// Value completes an operation whose service call cannot report "not found".
func (r Responder[R, T]) Value(value T, err error) (R, error) {
	if err != nil {
		var zero R
		return zero, err
	}

	return r.OK(value), nil
}

// Lookup completes an operation whose service call returns a value and reports
// whether it was found.
func (r Responder[R, T]) Lookup(value T, found bool, err error) (R, error) {
	if err != nil {
		var zero R
		return zero, err
	}

	if !found {
		return r.NotFound, nil
	}

	return r.OK(value), nil
}

// Exists completes an operation whose service call only reports whether the
// target existed, such as a delete.
func (r Responder[R, T]) Exists(found bool, err error) (R, error) {
	var value T

	return r.Lookup(value, found, err)
}
