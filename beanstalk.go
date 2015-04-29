package worker

import (
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/kr/beanstalk"
)

const (
	beanstalkHost  = "localhost"          // Beanstalk default host.
	beanstalkPort  = "11300"              // Beanstalk default port.
	beanstalkTube  = "default"            // Beanstalk default queue.
	beanstalkPrio  = 100                  // Beanstalk default job priority.
	beanstalkReady = "current-jobs-ready" // Beanstalk key for queue size.
)

var (
	beanstalkTimeout time.Duration = 1 * time.Second // Beanstalk reserve timeout.
)

// beanstalkMessage represents data returned by Reserve.
type beanstalkMessage struct {
	ID        uint64 // Message ID.
	*Envelope        // Holds parsed json.
}

// newBeanstalkMessage returns an instance of beanstalkMessage.
func newBeanstalkMessage(id uint64, payload []byte) (*beanstalkMessage, error) {
	base, err := NewEnvelope(payload)
	if err != nil {
		return nil, err
	}

	env := &beanstalkMessage{
		ID:       id,
		Envelope: base,
	}

	return env, nil
}

// BeanstalkQueue represents a Beanstalk queue.
type BeanstalkQueue struct {
	host string // Beanstalk host.
	port string // Beanstalk port.
	name string // Beanstalk tube name.
	prio uint32 // Beanstalk priority.

	conn *beanstalk.Conn
	tube *beanstalk.Tube
	tset *beanstalk.TubeSet
}

// NewBeanstalkQueue returns a queue instance using custom options.
func NewBeanstalkQueue(opts ...func(*BeanstalkQueue)) (Queue, error) {
	q := &BeanstalkQueue{
		host: beanstalkHost,
		port: beanstalkPort,
		name: beanstalkTube,
		prio: beanstalkPrio,
	}

	addr := net.JoinHostPort(q.host, q.port)

	conn, err := beanstalk.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	q.conn = conn

	tube := &beanstalk.Tube{
		Conn: q.conn,
		Name: q.name,
	}
	q.tube = tube
	q.tset = beanstalk.NewTubeSet(conn, q.name)

	return q, nil
}

// Put puts the job in the queue.
func (q *BeanstalkQueue) Put(j Job) error {
	typ, err := StructType(j)
	if err != nil {
		return err
	}

	job := &Payload{
		Type: typ,
		Args: j,
	}

	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = q.tube.Put(body, q.prio, 0, DefaultTTR)
	if err != nil {
		return err
	}

	return nil
}

// Get peeks a job from the queue.
func (q *BeanstalkQueue) Get() (Message, error) {
	id, payload, err := q.tset.Reserve(beanstalkTimeout)
	if err != nil {
		if cerr, ok := err.(beanstalk.ConnError); ok && cerr.Err == beanstalk.ErrTimeout {
			return nil, &Error{Err: "timeout", IsTimeout: true}
		}
		return nil, err
	}

	msg, err := newBeanstalkMessage(id, payload)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Delete deletes a job from the queue.
func (q *BeanstalkQueue) Delete(m Message) error {
	if env, ok := m.(*beanstalkMessage); ok {
		return q.conn.Delete(env.ID)
	}

	return NewErrorFmt("bad envelope: %v", m)
}

// Reject rejects the job marking it as failed.
func (q *BeanstalkQueue) Reject(m Message) error {
	if env, ok := m.(*beanstalkMessage); ok {
		return q.conn.Bury(env.ID, q.prio)
	}

	return NewErrorFmt("bad envelope: %v", m)
}

// Size returns the queue size, only ready jobs are returned.
func (q *BeanstalkQueue) Size() (uint64, error) {
	var size uint64

	dict, err := q.tube.Stats()
	if err != nil {
		return 0, err
	}

	if v, ok := dict[beanstalkReady]; !ok {
		return 0, NewErrorFmt("bad dict %v", v)
	} else {
		size, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, err
		}
	}

	return size, nil
}
