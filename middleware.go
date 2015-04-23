package worker

type JobRunner func(sw StatusWriter, fact string, args *Args)

type Handler interface {
	Exec(sw StatusWriter, fact string, args *Args, next JobRunner)
}

type HandlerFunc func(sw StatusWriter, fact string, args *Args, next JobRunner)

func (h HandlerFunc) Exec(sw StatusWriter, fact string, args *Args, next JobRunner) {
	h(sw, fact, args, next)
}

type middleware struct {
	handler Handler
	next    *middleware
}

func (m middleware) Exec(sw StatusWriter, fact string, args *Args) {
	m.handler.Exec(sw, fact, args, m.next.Exec)
}
