package errors

import "errors"

var (
	ErrSerialize error = errors.New("serialize failed")
	ErrNoTask    error = errors.New("no callback function")
	ErrNotFound  error = errors.New("key not found")
)
