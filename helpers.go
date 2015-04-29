package worker

import (
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

func StructType(v interface{}) (string, error) {
	typ := reflect.TypeOf(v)

	// If job is a pointer get the type
	// of the dereferenced object.
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return "", NewError("not a struct")
	}

	items := strings.Split(typ.String(), ".")

	if len(items) > 0 {
		return items[len(items)-1], nil
	}

	return "", NewError("bad struct name")
}
