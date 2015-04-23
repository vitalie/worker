package worker

import (
	"fmt"
)

type WorkerError struct {
	Err       string
	IsTimeout bool
}

func (e *WorkerError) Timeout() bool { return e.IsTimeout }

func (e *WorkerError) Error() string {
	if e == nil {
		return "<nil>"
	}

	return "worker: " + e.Err
}

func NewWorkerError(msg string) error {
	return &WorkerError{Err: msg}
}
func NewWorkerErrorFmt(format string, args ...interface{}) error {
	return NewWorkerError(fmt.Sprintf(format, args...))

}
