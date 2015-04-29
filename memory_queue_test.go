package worker_test

import (
	"testing"

	"github.com/vitalie/worker"
)

func TestMemoryQueue(t *testing.T) {
	j := &addJob{X: 1, Y: 2}
	q := worker.NewMemoryQueue()

	err := q.Put(j)
	if err != nil {
		t.Error(err)
	}

	size, err := q.Size()
	if err != nil {
		t.Error(err)
	}

	if size != 1 {
		t.Errorf("expecting size to be %v, got %v", 1, size)
	}

	msg, err := q.Get()
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

	err = q.Delete(msg)
	if err != nil {
		t.Error(err)
	}

	size, err = q.Size()
	if err != nil {
		t.Error(err)
	}

	if size != 0 {
		t.Errorf("expecting size to be %v, got %v", 0, size)
	}
}
