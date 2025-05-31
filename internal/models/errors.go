package models

import (
	"errors"
	"syscall"
)

var (
	ErrHTTPBadRequest          = errors.New("bad request")           // error for 400
	ErrHTTPNotFound            = errors.New("not found")             // error for 404
	ErrHTTPInternalServerError = errors.New("internal server error") // error for 500
	ErrNotFloat                = errors.New("not float")             // error if metric not float
	ErrNotInteger              = errors.New("not integer")           // error if metric not integer
	ErrNotSupported            = errors.New("not supported")         // error if metric not float nor integer
	ErrUnmarshalling           = errors.New("error unmarshalling")   // error for unmarshalling error
	ErrHTTPForbidden           = errors.New("forbidden")             // error for 403
)

const (
	MessageNotSupported string = "not supported" // message if metric not float nor integer
	MessageNotFloat     string = "not a float"   // message if metric not float
	MessageNotInteger   string = "not a integer" // message if metric not integer
	MessageNotFound     string = "not found"     // message for not founded metric
	MessageBadRequest   string = "bad request"   // message for bad request
)

// ErrRetryable slice of errors for syscall means if error is retryable.
var ErrRetryable = []error{
	syscall.ECONNRESET,
	syscall.ECONNABORTED,
	syscall.ECONNREFUSED,
}

// ErrIsRetryable returns true if error is retryable.
func ErrIsRetryable(err error) bool {
	for _, e := range ErrRetryable {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
