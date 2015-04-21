package worker

import (
	"errors"
)

var (
	ErrTimeout = errors.New("timeout")
)

type Runner interface {
	Type() string
	Run() error
}

type Factory interface {
	Make(*Args) (Job, error)
}

type Job interface {
	Runner
	Factory
}

type Queue interface {
	Put(Job) error
	Get() (*Message, error)
	Del(*Message) error
}
