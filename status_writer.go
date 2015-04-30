package worker

type StatusWriter interface {
	Set(interface{})
	Get() error
	OK() bool
}

type statusWriter struct {
	Err error
}

func NewStatusWriter() StatusWriter {
	return &statusWriter{}
}

func (sw *statusWriter) Set(err interface{}) {
	if cerr, ok := err.(error); ok {
		sw.Err = cerr
	} else {
		sw.Err = NewErrorFmt("%v", err)
	}
}

func (sw *statusWriter) Get() error {
	return sw.Err
}

func (sw *statusWriter) OK() bool {
	return sw.Err == nil
}
