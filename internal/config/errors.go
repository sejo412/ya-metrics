package config

import "errors"

var (
	ErrHTTPBadRequest = errors.New("bad request")
	ErrHTTPNotFound   = errors.New("not found")
)

const (
	MessageNotSupported string = "not supported"
	MessageNotFloat     string = "not a float"
	MessageNotInteger   string = "not a integer"
	MessageNotFound     string = "not found"
)
