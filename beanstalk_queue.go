package worker

import (
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/kr/beanstalk"
)

const (
	BeanstalkHost = "localhost" // Beanstalk default host.
	BeanstalkPort = "11300"     // Beanstalk default port.
	BeanstalkTube = "default"   // Beanstalk default queue.
	BeanstalkPrio = 100         // Beanstalk default job priority.

	beanstalkReadyKey  = "current-jobs-ready"
	beanstalkFailedKey = "current-jobs-buried"
)

var (
	BeanstalkTimeout time.Duration = 1 * time.Second // Beanstalk reserve timeout.
	BeanstalkTTR     time.Duration = 2 * DefaultTTR  // Beanstalk default TTR (time to run).
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
	Host string        // Beanstalk host.
	Port string        // Beanstalk port.
	Name string        // Beanstalk tube name.
	Prio uint32        // Beanstalk priority.
	TTR  time.Duration // Beanstalk time to run.

	conn *beanstalk.Conn
	tube *beanstalk.Tube
	tset *beanstalk.TubeSet
}

// NewBeanstalkQueue returns a queue instance using custom options.
func NewBeanstalkQueue(opts ...func(*BeanstalkQueue)) (Queue, error) {
	q := &BeanstalkQueue{
		Host: BeanstalkHost,
		Port: BeanstalkPort,
		Name: BeanstalkTube,
		Prio: BeanstalkPrio,
		TTR:  BeanstalkTTR,
	}

	addr := net.JoinHostPort(q.Host, q.Port)

	conn, err := beanstalk.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	q.conn = conn

	tube := &beanstalk.Tube{
		Conn: q.conn,
		Name: q.Name,
	}
	q.tube = tube
	q.tset = beanstalk.NewTubeSet(conn, q.Name)

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

	_, err = q.tube.Put(body, q.Prio, 0, q.TTR)
	if err != nil {
		return err
	}

	return nil
}

// Get peeks a job from the queue.
func (q *BeanstalkQueue) Get() (Message, error) {
	id, payload, err := q.tset.Reserve(BeanstalkTimeout)
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
		return q.conn.Bury(env.ID, q.Prio+1)
	}

	return NewErrorFmt("bad envelope: %v", m)
}

// Size returns the queue size, only ready jobs are returned.
func (q *BeanstalkQueue) Size() (uint64, uint64, error) {
	parse := func(key string, dict map[string]string) (uint64, error) {
		if v, ok := dict[key]; !ok {
			return 0, NewErrorFmt("bad dict %v", v)
		} else {
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return 0, err
			}
			return n, nil
		}
	}

	dict, err := q.tube.Stats()
	if err != nil {
		return 0, 0, err
	}

	ready, err := parse(beanstalkReadyKey, dict)
	if err != nil {
		return 0, 0, err
	}

	failed, err := parse(beanstalkFailedKey, dict)
	if err != nil {
		return 0, 0, err
	}

	return ready, failed, nil
}
