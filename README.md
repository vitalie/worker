# Worker [![Travis CI](https://travis-ci.org/vitalie/worker.svg?branch=master)](https://travis-ci.org/vitalie/worker) [![GoDoc](https://godoc.org/github.com/vitalie/worker?status.svg)](http://godoc.org/github.com/vitalie/worker)

An experimental background jobs processor using Beanstalk queue.

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
type addJob struct {
	X, Y int
}

// Make implements Factory interface, it's job is to
// initialize the struct with data received from the job queue.
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
	ctx := context.Background()

	q := worker.NewMemoryQueue()

	// Put job in the queue, visible/public fields
	// of the struct are serialized in JSON format
	// along with struct name.
	q.Put(ctx, &addJob{2, 3})

	// Create a worker pool with default settings,
	// common middlewares (Recovery, Logger)
	// using the `q` queue.
	pool := worker.NewPool(
		worker.SetQueue(q),
	)

	// Register tasks.
	pool.Add(&addJob{})

	// Starts the workers and processes the jobs
	// from the queue until process exists.
	pool.Run(ctx)
}
```

Example output:

``` bash
vitalie@black:~/tmp$ go run t.go
[worker] addJob {"X":2,"Y":3} ... started
2015/04/24 11:35:03 sum(2, 3) = 5
[worker] addJob {"X":2,"Y":3} ... in 89.866µs ... OK
```

## TODO

- Amazon SQS Adapter

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
