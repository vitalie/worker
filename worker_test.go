package worker_test

import (
	"log"
	"testing"

	"github.com/vitalie/worker"
	"golang.org/x/net/context"
)

type AddJob struct {
	X, Y int
}

func (j *AddJob) Type() string {
	return "add"
}

func (j *AddJob) Make(args *worker.Args) (worker.Job, error) {
	job := &AddJob{
		X: args.Get("X").MustInt(-1),
		Y: args.Get("Y").MustInt(-1),
	}
	return job, nil
}

func (j *AddJob) Run() error {
	log.Printf("sum(%v, %v) = %v", j.X, j.Y, j.X+j.Y)
	return nil
}

func TestBeanstalkQueue(t *testing.T) {
	j := &AddJob{X: 1, Y: 2}

	q, err := worker.NewBeanstalkQueue()
	if err != nil {
		t.Error(err)
	}

	err = q.Put(j)
	if err != nil {
		t.Error(err)
	}

	msg, err := q.Get()
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
}

func TestPool(t *testing.T) {
	q, err := worker.NewBeanstalkQueue()
	if err != nil {
		t.Error(err)
	}

	var sumtests = []struct {
		x, y int
		want int
	}{
		{1, 2, 3},
	}

	ctx := context.Background()

	pool := worker.NewPool(worker.SetPoolQueue(q))
	pool.Add(&AddJob{})
	pool.Run(ctx)

	for _, tt := range sumtests {
		if err := q.Put(&AddJob{X: tt.x, Y: tt.y}); err != nil {
			t.Fatal(err)
		}

		// if got := <-c; got != tt.want {
		// 	t.Errorf("sum(%d, %d) = %d; got %d", tt.x, tt.y, tt.want, got)
		// }
	}
}
