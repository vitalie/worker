package worker

import (
	"strconv"

	"github.com/bitly/go-simplejson"
)

type Envelope struct {
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

type Message struct {
	ID uint64
	*data
}

func NewMessage(id uint64, body []byte) (*Message, error) {
	json, err := toJson(body)
	if err != nil {
		return nil, err
	}
	return &Message{ID: id, data: &data{json}}, nil
}

func (m *Message) Type() string {
	return m.Get("type").MustString("<unknown>")
}

func (m *Message) Args() *Args {
	if args, ok := m.CheckGet("args"); ok {
		return &Args{&data{args}}
	} else {
		json, _ := toJson([]byte("[]"))
		return &Args{&data{json}}
	}
}

func (m *Message) String() string {
	if m == nil {
		return "<nil>"
	}

	return "&{" + strconv.FormatUint(m.ID, 10) + ", " + m.data.String() + "}"
}
