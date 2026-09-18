package storage

import "errors"

var (
	ErrNotFound = errors.New("storage: not found")
	ErrReadOnly = errors.New("storage: read only")
)
