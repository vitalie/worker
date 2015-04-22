package worker

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
