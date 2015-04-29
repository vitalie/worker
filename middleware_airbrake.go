package worker

import (
	"log"
	"net/http"
	"os"
	"runtime"

	"gopkg.in/airbrake/gobrake.v1"
)

type Airbrake struct {
	Logger    *log.Logger
	Airbrake  *gobrake.Notifier
	StackAll  bool
	StackSize int
}

func NewAirbrake(id int64, key, env string) *Airbrake {
	airbrake := gobrake.NewNotifier(id, key)
	airbrake.SetContext("environment", env)

	return &Airbrake{
		Logger:    log.New(os.Stdout, "[worker] ", 0),
		Airbrake:  airbrake,
		StackAll:  false,
		StackSize: 8 * 1024,
	}
}

func (r *Airbrake) Exec(sw StatusWriter, fact string, args *Args, next JobRunner) {
	defer func() {
		if err := recover(); err != nil {
			jinfo := fact + " " + args.String()
			stack := make([]byte, r.StackSize)
			stack = stack[:runtime.Stack(stack, r.StackAll)]

			f := "%s: PANIC: %s\n%s"
			r.Logger.Printf(f, jinfo, err, stack)

			go func() {
				if err := r.Airbrake.Notify(f, &http.Request{}); err != nil {
					r.Logger.Println("notify error:", err)
				}
			}()
		}
	}()

	next(sw, fact, args)
}
