package worker

import (
	"github.com/bitly/go-simplejson"
)

type envelope struct {
	Class string      `json:"class"`
	Args  interface{} `json:"args"`
}

type data struct {
	*simplejson.Json
}

type Args struct {
	*simplejson.Json
}

type Message struct {
	ID uint64
	*data
}

func NewMessage(d []byte) (*Message, error) {
	json, err := toJson(d)
	if err != nil {
		return nil, err
	}
	return &Message{data: &data{json}}, nil
}

func (m *Message) Type() string {
	return m.Get("class").MustString()
}

func (m *Message) Args() *Args {
	if args, ok := m.CheckGet("args"); ok {
		return &Args{args}
	} else {
		data, _ := toJson([]byte("[]"))
		return &Args{data}
	}
}
