package worker

type StatusWriter interface {
	Set(error)
	Get() error
	OK() bool
}

type statusWriter struct {
	Err error
}

func NewStatusWriter() StatusWriter {
	return &statusWriter{}
}

func (sw *statusWriter) Set(err error) {
	sw.Err = err
}

func (sw *statusWriter) Get() error {
	return sw.Err
}

func (sw *statusWriter) OK() bool {
	return sw.Err == nil
}
