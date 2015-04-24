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
    "fmt"

    "github.com/vitalie/worker"
)

type addJob struct {
    X, Y int
}

func (j *addJob) Make(args *worker.Args) (worker.Job, error) {
    job := &addJob{
        X:   args.Get("X").MustInt(-1),
        Y:   args.Get("Y").MustInt(-1),
        out: c,
    }
    return job, nil
}

func (j *addJob) Run() error {
    j.out <- j.X + j.Y
    return nil
}

func main() {
    // Create a worker pool with default settings.
    pool := worker.NewPool()

    // Register tasks.
    pool.Add(&addJob{})

    // Starts the workers and processes the jobs
    // from the queue until process exists.
    pool.Run(ctx)
}

```

## TODO

- Amazon SQS Adapter

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
