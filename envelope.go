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

func (e *Envelope) Type() string {
	return e.Get("type").MustString("")
}

func (e *Envelope) Args() *Args {
	if args, ok := e.CheckGet("args"); ok {
		return &Args{&data{args}}
	} else {
		json, _ := toJson([]byte("[]"))
		return &Args{&data{json}}
	}
}
