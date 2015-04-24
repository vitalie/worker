package worker_test

import (
	"log"
	"testing"

	"github.com/vitalie/worker"
	"golang.org/x/net/context"
)

var c chan int = make(chan int)

// addJob represets a test job which computes the sum of the
// X and Y and sends the result through out channel.
type addJob struct {
	X, Y int
	out  chan<- int
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

// badJob represets an background job which panics.
type badJob struct{}

func (j *badJob) Make(args *worker.Args) (worker.Job, error) { return &badJob{}, nil }

func (j *badJob) Run() error { panic("Boom!") }

func TestPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := worker.NewMemoryQueue()
	var sumtests = []struct {
		x, y int
		want int
	}{
		{0, 1, 1},
		{1, 0, 1},
		{2, 3, 5},
	}

	pool := worker.NewPool(
		worker.SetQueue(q),
	)
	pool.Add(&addJob{})
	pool.Add(&badJob{})

	go pool.Run(ctx)

	if err := q.Put(ctx, &badJob{}); err != nil {
		t.Fatal(err)
	}

	for _, tt := range sumtests {
		if err := q.Put(ctx, &addJob{X: tt.x, Y: tt.y}); err != nil {
			t.Fatal(err)
		}

		if got := <-c; got != tt.want {
			t.Errorf("sum(%d, %d) = %d; got %d", tt.x, tt.y, tt.want, got)
		} else {
			log.Printf("sum(%d, %d) = %d\n", tt.x, tt.y, got)
		}
	}
}
