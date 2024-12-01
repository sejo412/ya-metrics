package domain

import "errors"

var (
	ErrHTTPBadRequest          = errors.New("bad request")
	ErrHTTPNotFound            = errors.New("not found")
	ErrHTTPInternalServerError = errors.New("internal server error")
	ErrNotFloat                = errors.New("not float")
	ErrNotInteger              = errors.New("not integer")
	ErrNotSupported            = errors.New("not supported")
)

const (
	MessageNotSupported string = "not supported"
	MessageNotFloat     string = "not a float"
	MessageNotInteger   string = "not a integer"
	MessageNotFound     string = "not found"
)
