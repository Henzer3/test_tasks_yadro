package entity

import "errors"

var (
	ErrBadArguments = errors.New("bad arguments")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrParseFailed  = errors.New("parse failed")
)
