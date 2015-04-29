package worker

type Envelope struct {
	*data
}

func NewEnvelope(body []byte) (*Envelope, error) {
	json, err := toJson(body)
	if err != nil {
		return nil, err
	}
	return &Envelope{data: &data{json}}, nil
}

func (m *Envelope) Type() string {
	return m.Get("type").MustString("")
}

func (m *Envelope) Args() *Args {
	if args, ok := m.CheckGet("args"); ok {
		return &Args{&data{args}}
	} else {
		json, _ := toJson([]byte("[]"))
		return &Args{&data{json}}
	}
}
