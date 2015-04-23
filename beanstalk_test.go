package worker_test

import (
	"testing"

	"github.com/vitalie/worker"
	"golang.org/x/net/context"
)

func TestBeanstalkQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

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
