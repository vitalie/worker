package worker

import (
	"errors"
)

var (
	ErrTimeout = errors.New("timeout")
)

type Job interface {
	Type() string
	Make(*Args) (Job, error)
	Run() error
}

type Queue interface {
	Put(Job) error
	Get() (*Message, error)
	Del(*Message) error
}
