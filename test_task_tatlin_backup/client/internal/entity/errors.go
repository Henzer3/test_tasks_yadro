package entity

import "errors"

var (
	ErrAlreadyexist      = errors.New("already exist")
	ErrNotFoundDNS       = errors.New("not found")
	ErrUnAvailableServer = errors.New("unavailable server")
	ErrInternalError     = errors.New("internal error")
	ErrInvalidArgument   = errors.New("invalid argument")
)
