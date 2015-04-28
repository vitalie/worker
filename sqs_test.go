package worker_test

import (
	"os"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/vitalie/worker"
)

func TestSQSQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		t.Skip("skipping test, AWS credentials not found.")
	}

	j := &addJob{X: 1, Y: 2}
	q, err := worker.NewSQSQueue("worker_test")
	if err != nil {
		t.Error(err)
	}

	err = q.Put(j)
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
