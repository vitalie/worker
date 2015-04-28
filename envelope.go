package worker

type envelope struct {
	*data
}

func newEnvelope(body []byte) (*envelope, error) {
	json, err := toJson(body)
	if err != nil {
		return nil, err
	}
	return &envelope{data: &data{json}}, nil
}

func (m *envelope) Type() string {
	return m.Get("type").MustString("")
}

func (m *envelope) Args() *Args {
	if args, ok := m.CheckGet("args"); ok {
		return &Args{&data{args}}
	} else {
		json, _ := toJson([]byte("[]"))
		return &Args{&data{json}}
	}
}

type simpleEnvelope struct {
	ID uint64
	*envelope
}

func newSimpleEnvelope(id uint64, payload []byte) (*simpleEnvelope, error) {
	base, err := newEnvelope(payload)
	if err != nil {
		return nil, err
	}

	env := &simpleEnvelope{
		ID:       id,
		envelope: base,
	}

	return env, nil
}
