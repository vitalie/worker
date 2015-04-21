package worker

import (
	"github.com/bitly/go-simplejson"
)

func toJson(data []byte) (*simplejson.Json, error) {
	if json, err := simplejson.NewJson(data); err != nil {
		return nil, err
	} else {
		return json, nil
	}
}
