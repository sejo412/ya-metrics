package config

import "errors"

var (
	ErrHttpBadRequest = errors.New("bad request")
	ErrHttpNotFound   = errors.New("not found")
)

const (
	MessageNotSupported string = "not supported"
	MessageNotFloat     string = "not a float"
	MessageNotInteger   string = "not a integer"
	MessageNotFound     string = "not found"
)
