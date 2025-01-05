package models

import (
	"errors"
	"syscall"
)

var (
	ErrHTTPBadRequest          = errors.New("bad request")
	ErrHTTPNotFound            = errors.New("not found")
	ErrHTTPInternalServerError = errors.New("internal server error")
	ErrNotFloat                = errors.New("not float")
	ErrNotInteger              = errors.New("not integer")
	ErrNotSupported            = errors.New("not supported")
	ErrUnmarshalling           = errors.New("error unmarshalling")
)

const (
	MessageNotSupported string = "not supported"
	MessageNotFloat     string = "not a float"
	MessageNotInteger   string = "not a integer"
	MessageNotFound     string = "not found"
	MessageBadRequest   string = "bad request"
)

var ErrRetryable = []error{
	syscall.ECONNRESET,
	syscall.ECONNABORTED,
	syscall.ECONNREFUSED,
}

func ErrIsRetryable(err error) bool {
	for _, e := range ErrRetryable {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
