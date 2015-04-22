package worker

import (
	"golang.org/x/net/context"
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
	Put(context.Context, Job) error
	Get(context.Context) (*Message, error)
	Done(context.Context, *Message) error
	Fail(context.Context, *Message) error
	Size(context.Context) (uint64, error)
}
