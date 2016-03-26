package worker

import (
	"log"
	"os"
	"runtime"

	"gopkg.in/airbrake/gobrake.v2"
)

type Airbrake struct {
	Logger    *log.Logger
	Airbrake  *gobrake.Notifier
	StackAll  bool
	StackSize int
}

func NewAirbrake(id int64, key, env string) *Airbrake {
	airbrake := gobrake.NewNotifier(id, key)
	airbrake.AddFilter(func(notice *gobrake.Notice) *gobrake.Notice {
		notice.Context["environment"] = env
		return notice
	})

	return &Airbrake{
		Logger:    log.New(os.Stdout, "[worker] ", 0),
		Airbrake:  airbrake,
		StackAll:  false,
		StackSize: 8 * 1024,
	}
}

func (r *Airbrake) Exec(sw StatusWriter, fact string, args *Args, next JobRunner) {
	// Deferred function calls are executed in Last In First Out order
	// after the surrounding function returns.
	defer r.Airbrake.Flush()
	defer func() {
		if err := recover(); err != nil {
			jinfo := fact + " " + args.String()
			stack := make([]byte, r.StackSize)
			stack = stack[:runtime.Stack(stack, r.StackAll)]

			f := "%s: PANIC: %s\n%s"
			r.Logger.Printf(f, jinfo, err, stack)
			sw.Set(err)

			go func() {
				r.Airbrake.Notify(f, nil)
			}()
		}
	}()

	next(sw, fact, args)
}
