package timedtask

import "errors"

var (
	SerializeErr error = errors.New("serialize failed")
	NoTaskErr    error = errors.New("no callback function")
)
