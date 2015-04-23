package worker_test

import (
	"log"
	"testing"

	"github.com/vitalie/worker"
	"golang.org/x/net/context"
)

var c chan int = make(chan int)

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

func TestBeanstalkQueue(t *testing.T) {
	j := &addJob{X: 1, Y: 2}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q, err := worker.NewBeanstalkQueue()
	if err != nil {
		t.Error(err)
	}

	s1, err := q.Size(ctx)
	if err != nil {
		t.Error(err)
	}

	err = q.Put(ctx, j)
	if err != nil {
		t.Error(err)
	}

	s2, err := q.Size(ctx)
	if err != nil {
		t.Error(err)
	}

	if s2 < s1+1 {
		t.Errorf("expecting size to be %v, got %v", s1+1, s2)
	}

	msg, err := q.Get(ctx)
	if err != nil {
		t.Error(err)
	}

	typ := "addJob"
	if msg.Type() != typ {
		t.Errorf("expecting %q, got %q", typ, msg.Type())
	}

	x := msg.Args().Get("X").MustInt(-1)
	y := msg.Args().Get("Y").MustInt(-1)
	if x != 1 || y != 2 {
		t.Errorf("expecting (1, 2), got (%v, %v)", x, y)
		t.Error(msg.Args().Get("X"))
	}

	err = q.Delete(ctx, msg)
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
		{2, 3, 5},
	}

	pool := worker.NewPool(
		worker.SetQueue(q),
	)
	pool.Add(&addJob{})

	go pool.Start(ctx)

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
