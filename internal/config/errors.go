package config

import "errors"

var (
	ErrNotGoModule = errors.New("not a Go module — cd to module root")
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}
