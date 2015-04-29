package worker

import (
	"github.com/bitly/go-simplejson"
)

type Message interface {
	Type() string
	Args() *Args
}

type Queue interface {
	Put(Job) error
	Get() (Message, error)
	Delete(Message) error
	Reject(Message) error
	Size() (uint64, error)
}

// Payload represents a queue message payload.
type Payload struct {
	Type string      `json:"type"`
	Args interface{} `json:"args"`
}

type data struct {
	*simplejson.Json
}

func (d *data) String() string {
	json, err := d.MarshalJSON()
	if err != nil {
		return "MarshalJSON: " + err.Error()
	}
	return string(json)
}

type Args struct {
	*data
}
