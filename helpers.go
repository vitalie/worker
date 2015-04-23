package worker

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bitly/go-simplejson"
)

func toJson(data []byte) (*simplejson.Json, error) {
	if json, err := simplejson.NewJson(data); err != nil {
		return nil, err
	} else {
		return json, nil
	}
}

func structType(v interface{}) (string, error) {
	typ := reflect.TypeOf(v)

	// If job is a pointer get the type
	// of the dereferenced object.
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return "", fmt.Errorf("worker: not a struct")
	}

	items := strings.Split(typ.String(), ".")

	if len(items) > 0 {
		return items[len(items)-1], nil
	}

	return "", fmt.Errorf("worker: bad struct name")
}
