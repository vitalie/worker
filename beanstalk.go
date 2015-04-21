package worker

import (
	"encoding/json"
	"net"
	"time"

	"github.com/kr/beanstalk"
)

const (
	DefaultHost = "localhost"
	DefaultPort = "11300"
	DefaultTube = "default"
	DefaultPrio = 100
)

var (
	DefaultTTR     time.Duration = 60 * time.Second
	DefaultTimeout time.Duration = 24 * 365 * time.Hour
)

type BeanstalkQueue struct {
	host string
	port string
	name string

	conn *beanstalk.Conn
	tube *beanstalk.Tube
	tset *beanstalk.TubeSet
}

func NewBeanstalkQueue(opts ...func(*BeanstalkQueue)) (*BeanstalkQueue, error) {
	q := &BeanstalkQueue{
		host: DefaultHost,
		port: DefaultPort,
		name: DefaultTube,
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
	job := &envelope{
		Class: j.Type(),
		Args:  j,
	}

	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = q.tube.Put(body, DefaultPrio, 0, DefaultTTR)
	if err != nil {
		return err
	}

	return nil
}

func (q *BeanstalkQueue) Get() (*Message, error) {
	id, payload, err := q.tset.Reserve(DefaultTimeout)
	if err != nil {
		if cerr, ok := err.(beanstalk.ConnError); ok && cerr.Err == beanstalk.ErrTimeout {
			return nil, ErrTimeout
		}
		return nil, err
	}

	msg, err := NewMessage(payload)
	if err != nil {
		return nil, err
	}
	msg.ID = id

	return msg, nil
}

func (q *BeanstalkQueue) Del(*Message) error {
	return nil
}
