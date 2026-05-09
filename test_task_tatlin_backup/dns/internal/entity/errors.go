package entity

import "errors"

var (
	ErrNotFoundDNS  = errors.New("cant found dns")
	ErrAlreadyExist = errors.New("already exist")
)
