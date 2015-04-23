package worker

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{log.New(os.Stdout, "[worker] ", 0)}
}

func (l *Logger) Exec(sw StatusWriter, fact string, args *Args, next FactoryRunner) {
	jinfo := fact + " " + args.String()
	start := time.Now()

	l.Println(jinfo, "... started")

	next(sw, fact, args)

	status := "OK"
	if !sw.OK() {
		status = fmt.Sprintf("FAIL (%v)", sw.Get())
	}

	l.Printf("%s ... in %v ... %v", jinfo, time.Since(start), status)
}
