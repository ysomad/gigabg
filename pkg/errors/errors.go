package errors

// Error is the constant error type.
//
// See https://dave.cheney.net/2016/04/07/constant-errors.
type Error string

//nolint:errcheck // compile-time interface check
var _ error = Error("")

func (err Error) Error() string {
	return string(err)
}
