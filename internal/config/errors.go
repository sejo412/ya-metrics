package config

import "errors"

var (
	ErrHttpBadRequest = errors.New("bad request")
	ErrHttpNotFound   = errors.New("not found")
)

const (
	MessageNotSupported string = "not supported"
	MessageNotFloat     string = "not float"
	MessageNotInteger   string = "not integer"
)
