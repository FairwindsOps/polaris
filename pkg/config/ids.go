package config

import (
	"reflect"
)

func GetIDFromField(config interface{}, name string) string {
	t := reflect.TypeOf(config)
	field, ok := t.FieldByName(name)
	if !ok {
		panic("No JSON annotation for field " + name)
	}
	id, ok := field.Tag.Lookup("json")
	if !ok {
		panic("No JSON tag for field " + name)
	}
	return id
}
