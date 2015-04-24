package worker

import (
	"log"
	"os"
	"runtime"
)

type Recovery struct {
	Logger    *log.Logger
	StackAll  bool
	StackSize int
}

func NewRecovery() *Recovery {
	return &Recovery{
		Logger:    log.New(os.Stdout, "[worker] ", 0),
		StackAll:  false,
		StackSize: 8 * 1024,
	}
}

func (r *Recovery) Exec(sw StatusWriter, fact string, args *Args, next JobRunner) {
	defer func() {
		if err := recover(); err != nil {
			jinfo := fact + " " + args.String()
			stack := make([]byte, r.StackSize)
			stack = stack[:runtime.Stack(stack, r.StackAll)]

			f := "%s: PANIC: %s\n%s"
			r.Logger.Printf(f, jinfo, err, stack)
		}
	}()

	next(sw, fact, args)
}
