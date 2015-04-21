package worker

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
)

const (
	DefaultWorkersCount = 10
)

type Pool struct {
	count int
	queue Queue

	mux map[string]Job
}

func SetPoolQueue(q Queue) func(*Pool) {
	return func(p *Pool) {
		p.queue = q
	}
}

func NewPool(opts ...func(*Pool)) *Pool {
	pool := &Pool{
		count: DefaultWorkersCount,
		mux:   map[string]Job{},
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
	msg, err := p.queue.Get()
	if err != nil {
		return err
	}

	if f, ok := p.mux[msg.Type()]; ok {
		j, err := f.Make(msg.Args())
		if err != nil {
			log.Println("job make:", err)
		}
		if err := j.Run(); err != nil {
			log.Println("job run:", err)
		}
	} else {
		log.Println("bad type:", msg.Type())
	}

	return nil
}
