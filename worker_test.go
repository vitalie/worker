package worker_test

import (
	"log"
	"testing"

	"github.com/vitalie/worker"
	"golang.org/x/net/context"
)

var c chan int = make(chan int)

type AddJob struct {
	X, Y int
	out  chan<- int
}

func (j *AddJob) Type() string {
	return "add"
}

func (j *AddJob) Make(args *worker.Args) (worker.Job, error) {
	job := &AddJob{
		X:   args.Get("X").MustInt(-1),
		Y:   args.Get("Y").MustInt(-1),
		out: c,
	}
	return job, nil
}

func (j *AddJob) Run() error {
	j.out <- j.X + j.Y
	return nil
}

func TestBeanstalk(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	j := &AddJob{X: 1, Y: 2}

	q, err := worker.NewBeanstalkQueue()
	if err != nil {
		t.Error(err)
	}

	err = q.Put(ctx, j)
	if err != nil {
		t.Error(err)
	}

	size, err := q.Size(ctx)
	if err != nil {
		t.Error(err)
	}

	if size != 1 {
		t.Errorf("expecting size to be 1, got %v", size)
	}

	msg, err := q.Get(ctx)
	if err != nil {
		t.Error(err)
	}

	if msg.Type() != j.Type() {
		t.Errorf("expecting %q, got %q", j.Type(), msg.Type())
	}

	x := msg.Args().Get("X").MustInt(-1)
	y := msg.Args().Get("Y").MustInt(-1)
	if x != 1 || y != 2 {
		t.Errorf("expecting (1, 2), got (%v, %v)", x, y)
		t.Error(msg.Args().Get("X"))
	}

	err = q.Done(ctx, msg)
	if err != nil {
		t.Error(err)
	}
}

func TestPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q, err := worker.NewBeanstalkQueue()
	if err != nil {
		t.Error(err)
	}

	var sumtests = []struct {
		x, y int
		want int
	}{
		{1, 0, 1},
		{0, 1, 1},
		{2, 3, 5},
	}

	pool := worker.NewPool(
		worker.SetPoolQueue(q),
	)
	pool.Add(&AddJob{})

	go pool.Run(ctx)

	for _, tt := range sumtests {
		if err := q.Put(ctx, &AddJob{X: tt.x, Y: tt.y}); err != nil {
			t.Fatal(err)
		}

		if got := <-c; got != tt.want {
			t.Errorf("sum(%d, %d) = %d; got %d", tt.x, tt.y, tt.want, got)
		} else {
			log.Printf("sum(%d, %d) = %d\n", tt.x, tt.y, got)
		}
	}
}
