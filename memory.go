package worker

import (
	"encoding/json"
	"strconv"
	"sync"
)

// MemoryQueue represents an ordered queue,
// this queue is used in unit tests only.
type MemoryQueue struct {
	sync.Mutex
	count uint64
	l     []Message
}

func NewMemoryQueue() Queue {
	return &MemoryQueue{
		l: []Message{},
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
	id := strconv.FormatUint(q.count, 10)
	msg, err := newEnvelope(id, payload)
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

	var lst []Message
	for _, m := range q.l {
		if m.ID() != msg.ID() {
			lst = append(lst, m)
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
