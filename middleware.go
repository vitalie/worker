package worker

type FactoryRunner func(StatusWriter, string, *Args)

type Handler interface {
	Exec(StatusWriter, string, *Args, FactoryRunner)
}

type HandlerFunc func(StatusWriter, string, *Args, FactoryRunner)

func (h HandlerFunc) Exec(sw StatusWriter, fact string, args *Args, next FactoryRunner) {
	h(sw, fact, args, next)
}

type middleware struct {
	handler Handler
	next    *middleware
}

func (m middleware) Exec(sw StatusWriter, fact string, args *Args) {
	m.handler.Exec(sw, fact, args, m.next.Exec)
}
