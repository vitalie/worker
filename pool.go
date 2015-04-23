package worker

import (
	"fmt"
	"log"
	"os"
	"sync"

	"golang.org/x/net/context"
)

const (
	DefaultWorkersCount = 10
)

type Pool struct {
	count int
	queue Queue

	mux    map[string]Job
	logger *log.Logger
}

func SetQueue(q Queue) func(*Pool) {
	return func(p *Pool) {
		p.queue = q
	}
}

func SetWorkers(count int) func(*Pool) {
	return func(p *Pool) {
		p.count = count
	}
}

func NewPool(opts ...func(*Pool)) *Pool {
	pool := &Pool{
		count:  DefaultWorkersCount,
		mux:    map[string]Job{},
		logger: log.New(os.Stdout, "[worker] ", 0),
	}

	for _, opt := range opts {
		opt(pool)
	}

	return pool
}

func (p *Pool) Add(j Job) error {
	if _, ok := p.mux[j.Type()]; ok {
		return fmt.Errorf("factory %q exists already", j.Type())
	}
	p.mux[j.Type()] = j
	return nil
}

func (p *Pool) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	c := make(chan *Message)
	defer close(c)

	// Start workers.
	wg.Add(p.count)
	for i := 0; i < p.count; i++ {
		go func() {
			defer wg.Done()
			p.worker(ctx, c)
		}()
	}

	// Start producer.
	for {
		msg, err := p.queue.Get(ctx)
		if err != nil {
			if wkerr, ok := err.(*WorkerError); ok && wkerr.Timeout() {
				continue
			}
			return err
		}

		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		case c <- msg:
		}
	}
}

func (p *Pool) worker(ctx context.Context, in chan *Message) {
	for msg := range in {
		if f, ok := p.mux[msg.Type()]; ok {
			// Start the job.
			c := p.process(f, msg.Args())

			// Wait completion.
			select {
			case <-ctx.Done():
				return
			case err := <-c:
				if err != nil {
					p.logger.Println(msg, "...", err)
					if err := p.queue.Fail(ctx, msg); err != nil {
						p.logger.Println(err)
					}
				} else {
					p.logger.Println(msg, "...", "OK")
					if err := p.queue.Done(ctx, msg); err != nil {
						p.logger.Println(err)
					}
				}
			}
		} else {
			if err := p.queue.Fail(ctx, msg); err != nil {
				p.logger.Println(err)
			}
		}
	}
}

func (p *Pool) process(f Factory, args *Args) <-chan error {
	out := make(chan error, 1)

	defer func() {
		if err := recover(); err != nil {
			out <- fmt.Errorf("worker:", err)
		}
		close(out)
	}()

	j, err := f.Make(args)
	if err != nil {
		out <- fmt.Errorf("worker: make failed: %v", err)
		return out
	}

	if err := j.Run(); err != nil {
		out <- fmt.Errorf("worker: run failed: %v", err)
		return out
	}

	out <- nil
	return out
}
