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

func (r *Recovery) Exec(sw StatusWriter, fact string, args *Args, next FactoryRunner) {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, r.StackSize)
			stack = stack[:runtime.Stack(stack, r.StackAll)]

			f := "PANIC: %s\n%s"
			r.Logger.Printf(f, err, stack)
		}
	}()

	next(sw, fact, args)
}
