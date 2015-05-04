# Worker [![Travis CI](https://travis-ci.org/vitalie/worker.svg?branch=master)](https://travis-ci.org/vitalie/worker) [![GoDoc](https://godoc.org/github.com/vitalie/worker?status.svg)](http://godoc.org/github.com/vitalie/worker)

An experimental background jobs processor. The following queues are supported:

* [AWS SQS](http://godoc.org/github.com/vitalie/worker#SQSQueue "Amazon Simple Queue Service (SQS)")
* [Beanstalk](http://godoc.org/github.com/vitalie/worker#BeanstalkQueue)
* [Memory](http://godoc.org/github.com/vitalie/worker#MemoryQueue)

## Installation

``` bash
$ go get -u github.com/vitalie/worker
```

## Usage

``` go
package main

import (
	"log"

	"github.com/vitalie/worker"
	"golang.org/x/net/context"
)

// addJob represents a simple background job.
// When the job is queued, visible (public) fields
// of the struct are serialized in JSON format along
// with struct's name.
type addJob struct {
	X, Y int
}

// Make implements Factory interface, it initialize
// the struct with data received from the job queue.
func (j *addJob) Make(args *worker.Args) (worker.Job, error) {
	job := &addJob{
		X: args.Get("X").MustInt(-1),
		Y: args.Get("Y").MustInt(-1),
	}
	return job, nil
}

// Run implements Runner interface, this is the
// function which is executed by background processor
// after initialization.
func (j *addJob) Run() error {
	sum := j.X + j.Y
	log.Printf("sum(%d, %d) = %d\n", j.X, j.Y, sum)
	return nil
}

func main() {
	q := worker.NewMemoryQueue()
	q.Put(&addJob{2, 3})

	// Create a worker pool with default settings,
	// common middlewares (Recovery, Logger)
	// using the `q` queue.
	pool := worker.NewPool(
		worker.SetQueue(q),
	)

	// Register the job.
	pool.Add(&addJob{})

	// Starts the workers and processes the jobs
	// from the queue until process exists.
	pool.Run(context.Background())
}
```

Example output:

``` bash
vitalie@black:~/tmp$ go run t.go
[worker] addJob {"X":2,"Y":3} ... started
2015/04/24 14:36:21 sum(2, 3) = 5
[worker] addJob {"X":2,"Y":3} ... in 98.945µs ... OK
^C[worker] Quit signal received ...
[worker] Stopping workers ...
[worker] Shutdown completed!
```

## TODO

- Job scheduler

## Credits

- [codegangsta](https://github.com/codegangsta)
- [jrallison](https://github.com/jrallison)

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
