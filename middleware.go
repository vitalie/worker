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

// CommonStack is used to configure default middleware
// that's common for most applications: Recovery, Logger.
func CommonStack() []Handler {
	return []Handler{NewRecovery(), NewLogger()}
}

// AirbrakeStack is used to configure default middleware
// using Airbrake middleware instead of Recovery,
// configured middlewares: Airbreak, Logger.
func AirbreakStack(id int64, key, env string) []Handler {
	return []Handler{NewAirbrake(id, key, env), NewLogger()}
}
