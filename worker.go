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
