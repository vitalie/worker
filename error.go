package worker

import (
	"fmt"
)

type Error struct {
	Err       string
	IsTimeout bool
}

func (e *Error) Timeout() bool { return e.IsTimeout }

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	return "worker: " + e.Err
}

func NewError(msg string) error {
	return &Error{Err: msg}
}
func NewErrorFmt(format string, args ...interface{}) error {
	return NewError(fmt.Sprintf(format, args...))

}
