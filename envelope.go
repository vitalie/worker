package worker

import (
	"strconv"
)

type envelope struct {
	MsgID uint64
	*data
}

func newEnvelope(id uint64, body []byte) (*envelope, error) {
	json, err := toJson(body)
	if err != nil {
		return nil, err
	}
	return &envelope{MsgID: id, data: &data{json}}, nil
}

func (m *envelope) ID() string {
	return strconv.FormatUint(m.MsgID, 10)
}

func (m *envelope) Type() string {
	return m.Get("type").MustString("<unknown>")
}

func (m *envelope) Args() *Args {
	if args, ok := m.CheckGet("args"); ok {
		return &Args{&data{args}}
	} else {
		json, _ := toJson([]byte("[]"))
		return &Args{&data{json}}
	}
}

func (m *envelope) String() string {
	if m == nil {
		return "<nil>"
	}

	return "&{" + m.ID() + ", " + m.data.String() + "}"
}
