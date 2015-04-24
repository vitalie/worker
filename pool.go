package worker

import (
	"log"
	"os"
	"sync"

	"golang.org/x/net/context"
)

const (
	DefaultWorkersCount = 10
)

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

	// Fan-out channel.
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
			if wkerr, ok := err.(*Error); ok && wkerr.Timeout() {
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

// worker executes jobs from the in channel in a separate goroutine.
func (p *Pool) worker(ctx context.Context, in chan *Message) {
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
				if err := p.queue.Delete(ctx, msg); err != nil {
					p.logger.Println("delete:", msg, err)
				}
			} else {
				if err := p.queue.Reject(ctx, msg); err != nil {
					p.logger.Println("reject:", msg, err)
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
