package worker

import (
	"encoding/json"
	"sync"
)

type memoryMessage struct {
	ID uint64
	*Envelope
}

func newMemoryMessage(id uint64, payload []byte) (*memoryMessage, error) {
	base, err := NewEnvelope(payload)
	if err != nil {
		return nil, err
	}

	env := &memoryMessage{
		ID:       id,
		Envelope: base,
	}

	return env, nil
}

// MemoryQueue represents an ordered queue,
// this queue is used in unit tests only.
type MemoryQueue struct {
	sync.Mutex
	size uint64
	l    []*memoryMessage
}

func NewMemoryQueue() Queue {
	return &MemoryQueue{
		l: []*memoryMessage{},
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

	q.size++
	msg, err := newMemoryMessage(q.size, payload)
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

	env, ok := msg.(*memoryMessage)
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
