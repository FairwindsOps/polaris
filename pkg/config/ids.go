package config

import (
	"reflect"
)

// GetIDFromField returns the JSON key associated with a particular field, which serves as the check ID.
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
