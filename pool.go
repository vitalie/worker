package worker

import (
	"log"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context"
)

const (
	DefaultWorkersCount = 10
)

var (
	DefaultTTR time.Duration = 5 * time.Minute
)

// Pool represents a pool of workers connected to a queue.
type Pool struct {
	queue Queue         // input queue
	count int           // workers count
	ttr   time.Duration // Time to run.

	middleware middleware
	handlers   []Handler
	mux        map[string]Factory
	logger     *log.Logger
}

// NewPool returns a new Pool instance.
func NewPool(opts ...func(*Pool)) *Pool {
	pool := &Pool{
		queue:    NewMemoryQueue(),
		count:    DefaultWorkersCount,
		ttr:      DefaultTTR,
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
			if err := p.execute(fact, args); err != nil {
				sw.Set(err)
			}
		}),
		&middleware{},
	}
}

// Run starts processing jobs from the queue.
func (p *Pool) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	// Handle unix signals.
	sig := trap()

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
		p.logger.Println("Quit signal received ...")
		cancel()
	}

	p.logger.Println("Stopping workers ...")
	close(c)
	wg.Wait()
	p.logger.Println("Shutdown completed!")
	return err
}

// master polls the input queue sending jobs to workers through a blocking channel.
func (p *Pool) master(ctx context.Context, c chan<- Message) {
	qs := newQueueService(p.queue)
	var r *response
	for {
		select {
		case <-ctx.Done():
			return
		case r = <-qs.get():
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

		ctx, cancel := context.WithTimeout(ctx, p.ttr)
		defer cancel()

		// Wait job completion.
		select {
		case <-ctx.Done():
			if err := p.queue.Reject(msg); err != nil {
				p.logger.Println("Reject failure:", msg, err)
			}
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
