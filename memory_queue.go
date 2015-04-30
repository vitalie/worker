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
// this queue is used mainly for unit tests.
type MemoryQueue struct {
	sync.Mutex
	counter uint64
	ready   []*memoryMessage
	failed  []*memoryMessage
}

func NewMemoryQueue() Queue {
	return &MemoryQueue{
		ready:  []*memoryMessage{},
		failed: []*memoryMessage{},
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

	q.counter++
	msg, err := newMemoryMessage(q.counter, payload)
	if err != nil {
		return err
	}
	q.ready = append(q.ready, msg)

	return nil
}

func (q *MemoryQueue) Get() (Message, error) {
	q.Lock()
	defer q.Unlock()

	if len(q.ready) == 0 {
		return nil, &Error{Err: "timeout", IsTimeout: true}
	}

	var m Message
	m, q.ready = q.ready[len(q.ready)-1], q.ready[:len(q.ready)-1]
	return m, nil
}

func (q *MemoryQueue) Delete(msg Message) error {
	q.Lock()
	defer q.Unlock()

	env, ok := msg.(*memoryMessage)
	if !ok {
		return NewErrorFmt("bad envelope: %v", msg)
	}

	_, q.ready = q.remove(env.ID, q.ready)

	return nil
}

func (q *MemoryQueue) Reject(msg Message) error {
	env, ok := msg.(*memoryMessage)
	if !ok {
		return NewErrorFmt("bad envelope: %v", msg)
	}

	m, l := q.remove(env.ID, q.ready)
	q.ready = l
	q.failed = append(q.failed, m)

	return nil
}

func (q *MemoryQueue) Size() (uint64, uint64, error) {
	q.Lock()
	defer q.Unlock()

	ready := len(q.ready)
	failed := len(q.failed)

	return uint64(ready), uint64(failed), nil
}

func (q *MemoryQueue) remove(id uint64, list []*memoryMessage) (*memoryMessage, []*memoryMessage) {
	var m *memoryMessage
	var l []*memoryMessage

	for _, i := range list {
		if i.ID != id {
			l = append(l, i)
		} else {
			m = i
		}
	}

	return m, l
}
