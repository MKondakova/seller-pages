package api

import "errors"

var (
	ErrBadRequest     = errors.New("bad request")
	ErrInternalServer = errors.New("internal server error")
)
