package worker

type Runner interface {
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
	Delete(*Message) error
	Reject(*Message) error
	Size() (uint64, error)
}
