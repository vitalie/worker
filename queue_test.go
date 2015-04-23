package worker_test

import (
	"encoding/json"

	"github.com/vitalie/worker"
	"golang.org/x/net/context"
)

type MemoryQueue struct {
	count uint64
	l     []*worker.Message
}

func NewMemoryQueue() worker.Queue {
	return &MemoryQueue{
		l: []*worker.Message{},
	}
}

func (q *MemoryQueue) Put(ctx context.Context, j worker.Job) error {
	typ, err := worker.StructType(j)
	if err != nil {
		return err
	}

	job := &worker.Envelope{
		Type: typ,
		Args: j,
	}

	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	q.count++
	msg, err := worker.NewMessage(q.count, body)
	if err != nil {
		return err
	}
	q.l = append(q.l, msg)

	return nil
}

func (q *MemoryQueue) Get(ctx context.Context) (*worker.Message, error) {
	if len(q.l) == 0 {
		return nil, &worker.Error{Err: "timeout", IsTimeout: true}
	}

	var m *worker.Message
	m, q.l = q.l[len(q.l)-1], q.l[:len(q.l)-1]
	return m, nil
}

func (q *MemoryQueue) Delete(ctx context.Context, msg *worker.Message) error {
	var lst []*worker.Message
	for _, m := range q.l {
		if m.ID != msg.ID {
			lst = append(lst, m)
		}
	}
	return nil
}

func (q *MemoryQueue) Reject(ctx context.Context, msg *worker.Message) error {
	return q.Delete(ctx, msg)
}

func (q *MemoryQueue) Size(ctx context.Context) (uint64, error) {
	size := len(q.l)
	return uint64(size), nil
}
