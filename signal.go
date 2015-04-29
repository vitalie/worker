package worker

import (
	"os"
	"os/signal"
	"syscall"
)

// trap traps OS signals to allow clean exit.
func trap() <-chan struct{} {
	out := make(chan struct{}, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer close(out)
		for s := range signals {
			switch s {
			case syscall.SIGINT, syscall.SIGUSR1, syscall.SIGTERM:
				out <- struct{}{}
				return
			}
		}
	}()

	return out
}
