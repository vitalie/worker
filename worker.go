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
	Ack(*Message) error
	Size() (uint64, error)
}
