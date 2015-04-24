package worker

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kr/beanstalk"
)

const (
	defaultHost = "localhost"
	defaultPort = "11300"
	defaultTube = "default"
	defaultPrio = 100

	sizeKey = "current-jobs-ready"
)

var (
	defaultTimeout time.Duration = 1 * time.Second
)

type BeanstalkQueue struct {
	host string
	port string
	name string
	conn *beanstalk.Conn
	tube *beanstalk.Tube
	tset *beanstalk.TubeSet
}

func NewBeanstalkQueue(opts ...func(*BeanstalkQueue)) (Queue, error) {
	q := &BeanstalkQueue{
		host: defaultHost,
		port: defaultPort,
		name: defaultTube,
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

	job := &Envelope{
		Type: typ,
		Args: j,
	}

	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = q.tube.Put(body, defaultPrio, 0, DefaultTTR)
	if err != nil {
		return err
	}

	return nil
}

func (q *BeanstalkQueue) Get() (*Message, error) {
	id, payload, err := q.tset.Reserve(defaultTimeout)
	if err != nil {
		if cerr, ok := err.(beanstalk.ConnError); ok && cerr.Err == beanstalk.ErrTimeout {
			return nil, &Error{Err: "timeout", IsTimeout: true}
		}
		return nil, err
	}

	msg, err := NewMessage(id, payload)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (q *BeanstalkQueue) Delete(m *Message) error {
	return q.conn.Delete(m.ID)
}

func (q *BeanstalkQueue) Reject(m *Message) error {
	return q.conn.Bury(m.ID, defaultPrio)
}

func (q *BeanstalkQueue) Size() (uint64, error) {
	var size uint64

	dict, err := q.tube.Stats()
	if err != nil {
		return 0, err
	}

	v, ok := dict[sizeKey]
	if !ok {
		return 0, fmt.Errorf("worker: bad size %v", v)
	}

	size, err = strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}

	return size, nil
}
