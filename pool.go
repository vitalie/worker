package worker

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/net/context"
)

const (
	DefaultWorkersCount = 10
)

type response struct {
	Msg Message
	Err error
}

// Pool represents a pool of workers connected to a queue.
type Pool struct {
	count int   // workers count
	queue Queue // input queue

	middleware middleware
	handlers   []Handler
	mux        map[string]Factory
	logger     *log.Logger
}

// SetQueue assigns a custom queue to worker pool.
func SetQueue(q Queue) func(*Pool) {
	return func(p *Pool) {
		p.queue = q
	}
}

// SetWorkers configures the pool concurrency.
func SetWorkers(n int) func(*Pool) {
	return func(p *Pool) {
		p.count = n
	}
}

// NewPool returns a new Pool instance.
func NewPool(opts ...func(*Pool)) *Pool {
	pool := &Pool{
		count:    DefaultWorkersCount,
		queue:    NewMemoryQueue(),
		mux:      map[string]Factory{},
		logger:   log.New(os.Stdout, "[worker] ", 0),
		handlers: CommonStack(),
	}

	// Apply options.
	for _, opt := range opts {
		opt(pool)
	}

	// Init middleware stack.
	pool.middleware = pool.build(pool.handlers)

	return pool
}

// Add registers a new job factory.
func (p *Pool) Add(f Factory) error {
	typ, err := StructType(f)
	if err != nil {
		return err
	}

	if _, ok := p.mux[typ]; ok {
		return NewErrorFmt("factory %q exists already", typ)
	}
	p.mux[typ] = f
	return nil
}

// Use appends a new middleware to current stack.
func (p *Pool) Use(h Handler) {
	p.handlers = append(p.handlers, h)
	p.middleware = p.build(p.handlers)
}

// build iterates over the handlers, it returns a
// list of middlewares with each item linked to
// the next one.
func (p *Pool) build(hs []Handler) middleware {
	if len(hs) == 0 {
		return p.last()
	}
	next := p.build(hs[1:])
	return middleware{hs[0], &next}
}

// last builds and returns the last middleware, which will
// execute the job without calling next middleware.
func (p *Pool) last() middleware {
	return middleware{
		HandlerFunc(func(sw StatusWriter, fact string, args *Args, next JobRunner) {
			err := p.execute(fact, args)
			sw.Set(err)
		}),
		&middleware{},
	}
}

// Run starts processing jobs from the queue.
func (p *Pool) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	// Handle unix signals.
	sig := p.trap()

	ctx, cancel := context.WithCancel(ctx)

	// Fan-out channel.
	c := make(chan Message)

	// Start workers.
	wg.Add(p.count + 1)
	for i := 0; i < p.count; i++ {
		go func() {
			defer wg.Done()
			p.worker(ctx, c)
		}()
	}

	// Start the master.
	go func() {
		defer wg.Done()
		p.master(ctx, c)
	}()

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-sig:
		cancel()
	}

	p.logger.Println("Stopping workers ...")
	close(c)
	wg.Wait()
	p.logger.Println("Shutdown completed!")
	return err
}

func (p *Pool) master(ctx context.Context, c chan<- Message) {
	var r *response
	for {
		select {
		case <-ctx.Done():
			return
		case r = <-p.get():
		}

		if r.Err != nil {
			if wkerr, ok := r.Err.(*Error); ok && wkerr.Temporary() {
				continue
			}
			return
		}

		select {
		case <-ctx.Done():
			return
		case c <- r.Msg:
		}
	}
}

// worker executes jobs from the in channel in a separate goroutine.
func (p *Pool) worker(ctx context.Context, in <-chan Message) {
	for msg := range in {
		status := NewStatusWriter()
		done := make(chan struct{}, 1)

		// Start the job in a separate goroutine.
		go func() {
			p.Exec(status, msg.Type(), msg.Args())
			done <- struct{}{}
		}()

		// Wait job completion.
		select {
		case <-ctx.Done():
			return
		case <-done:
			if status.OK() {
				if err := p.queue.Delete(msg); err != nil {
					p.logger.Println("Delete failure:", msg, err)
				}
			} else {
				if err := p.queue.Reject(msg); err != nil {
					p.logger.Println("Reject failure:", msg, err)
				}
			}
		}
	}
}

// Exec runs the job passing it through the middleware stack.
func (p *Pool) Exec(sw StatusWriter, fact string, args *Args) {
	p.middleware.Exec(sw, fact, args)
}

// execute runs the job without passing it through the
// middleware stack.
func (p *Pool) execute(fact string, args *Args) error {
	f, ok := p.mux[fact]
	if !ok {
		return NewErrorFmt("bad type: %v", fact)
	}

	j, err := f.Make(args)
	if err != nil {
		return NewErrorFmt("make: %v", err)
	}

	if err := j.Run(); err != nil {
		return NewErrorFmt("run: %v", err)
	}

	return nil
}

// trap traps OS signals to allow clean exit.
func (p *Pool) trap() <-chan struct{} {
	out := make(chan struct{}, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer close(out)
		for s := range signals {
			switch s {
			case syscall.SIGINT, syscall.SIGUSR1, syscall.SIGTERM:
				p.logger.Println("Quit signal received ...")
				out <- struct{}{}
				return
			}
		}
	}()

	return out
}

// get service peeks a message from queue and
// returns it through a channel without blocking
// the caller.
func (p *Pool) get() <-chan *response {
	// Buffered channel to avoid goroutine leaking.
	out := make(chan *response, 1)

	// Start a gorouting to avoid blocking.
	go func() {
		defer close(out)

		msg, err := p.queue.Get()
		if err != nil {
			out <- &response{Err: err}
			return
		}

		out <- &response{Msg: msg}
	}()

	return out
}
