package models

import (
	"errors"
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
