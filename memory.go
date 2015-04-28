package worker

import (
	"encoding/json"
	"sync"
)

// MemoryQueue represents an ordered queue,
// this queue is used in unit tests only.
type MemoryQueue struct {
	sync.Mutex
	count uint64
	l     []*simpleEnvelope
}

func NewMemoryQueue() Queue {
	return &MemoryQueue{
		l: []*simpleEnvelope{},
	}
}

func (q *MemoryQueue) Put(j Job) error {
	q.Lock()
	defer q.Unlock()

	typ, err := StructType(j)
	if err != nil {
		return err
	}

	job := &Payload{
		Type: typ,
		Args: j,
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}

	q.count++
	msg, err := newSimpleEnvelope(q.count, payload)
	if err != nil {
		return err
	}
	q.l = append(q.l, msg)

	return nil
}

func (q *MemoryQueue) Get() (Message, error) {
	q.Lock()
	defer q.Unlock()

	if len(q.l) == 0 {
		return nil, &Error{Err: "timeout", IsTimeout: true}
	}

	var m Message
	m, q.l = q.l[len(q.l)-1], q.l[:len(q.l)-1]
	return m, nil
}

func (q *MemoryQueue) Delete(msg Message) error {
	q.Lock()
	defer q.Unlock()

	env, ok := msg.(*simpleEnvelope)
	if !ok {
		return NewErrorFmt("bad envelope: %v", msg)
	}

	var lst []Message
	for _, i := range q.l {
		if i.ID != env.ID {
			lst = append(lst, i)
		}
	}

	return nil
}

func (q *MemoryQueue) Reject(msg Message) error {
	return q.Delete(msg)
}

func (q *MemoryQueue) Size() (uint64, error) {
	q.Lock()
	defer q.Unlock()

	size := len(q.l)
	return uint64(size), nil
}
