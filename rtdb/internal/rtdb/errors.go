package rtdb

import "errors"

var (
	ErrNotFound = errors.New("tag not found")
	ErrReadOnly = errors.New("tag is not writable")
)
