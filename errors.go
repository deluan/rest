package rest

import (
	"errors"
	"fmt"
)

/*
Possible errors returned by a Repository implementation. Any error other than these will make the REST controller
return a 500 http status code.
*/
var (
	// ErrNotFound will make the controller return a 404 error
	ErrNotFound = errors.New("data not found")

	// ErrPermissionDenied will make the controller return a 403 error
	ErrPermissionDenied = errors.New("permission denied")
)

// ValidationError will make the controller return a 400 error, with the listed errors in the body
type ValidationError struct {
	Errors map[string]string `json:"errors"`
}

func (m ValidationError) Error() string {
	return fmt.Sprintf("Errors: %v", m.Errors)
}
