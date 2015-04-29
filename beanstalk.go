package worker

import (
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/kr/beanstalk"
)

const (
	beanstalkHost  = "localhost"
	beanstalkPort  = "11300"
	beanstalkTube  = "default"
	beanstalkPrio  = 100
	beanstalkReady = "current-jobs-ready"
)

var (
	beanstalkTimeout time.Duration = 1 * time.Second
)

type beanstalkMessage struct {
	ID uint64
	*Envelope
}

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

type BeanstalkQueue struct {
	host string
	port string
	name string
	prio uint32
	conn *beanstalk.Conn
	tube *beanstalk.Tube
	tset *beanstalk.TubeSet
}

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

func (q *BeanstalkQueue) Delete(m Message) error {
	if env, ok := m.(*beanstalkMessage); ok {
		return q.conn.Delete(env.ID)
	}

	return NewErrorFmt("bad envelope: %v", m)
}

func (q *BeanstalkQueue) Reject(m Message) error {
	if env, ok := m.(*beanstalkMessage); ok {
		return q.conn.Bury(env.ID, q.prio)
	}

	return NewErrorFmt("bad envelope: %v", m)
}

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
