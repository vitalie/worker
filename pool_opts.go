package worker

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
