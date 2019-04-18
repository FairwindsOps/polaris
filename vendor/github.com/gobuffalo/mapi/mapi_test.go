package mapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func M() Mapi {
	return Mapi{
		"string": "Hello",
		"int":    42,
		"bool":   true,
		"smap": map[string]string{
			"a": "A",
			"b": "B",
		},
		"imap": map[string]interface{}{
			"c":     "C",
			"three": 3,
			"funk":  func() {},
			"mapix": Mapi{
				"one": 1,
				"two": "too",
				"fnc": func() {},
			},
			"iimap": map[string]interface{}{
				"d": "D",
				"ismap": map[string]string{
					"e": "E",
					"f": "F",
				},
			},
		},
	}
}

func Test_Mapi_UnmarshalJSON(t *testing.T) {
	r := require.New(t)

	m := M()
	b, err := json.Marshal(m)
	r.NoError(err)

	mm := Mapi{}
	r.NoError(json.Unmarshal(b, &mm))
	r.Len(mm, 5)

	imap, ok := mm["imap"].(Mapi)
	r.True(ok)
	r.Equal("C", imap["c"])

	iimap, ok := imap["iimap"].(Mapi)
	r.True(ok)
	r.Equal("D", iimap["d"])
}

func Test_Mapi_Pluck(t *testing.T) {
	table := []struct {
		key  string
		out  interface{}
		boom bool
	}{
		{"string", "Hello", false},
		{"imap:mapix:two", "too", false},
		{"imap:mapix:func", nil, true},
		{"imap:iimap:ismap:f", "F", false},
		{"imap:i:dont:exist", nil, true},
	}

	m := M()
	for _, tt := range table {
		t.Run(tt.key, func(st *testing.T) {
			r := require.New(st)
			i, err := m.Pluck(tt.key)
			if tt.boom {
				r.Error(err)
				return
			}
			r.NoError(err)
			r.Equal(tt.out, i)
		})
	}
}

func Test_Mapi_MarshalJSON(t *testing.T) {
	r := require.New(t)

	m := M()

	act, err := json.Marshal(m)
	r.NoError(err)

	am := map[string]interface{}{
		"string": "Hello",
		"int":    42,
		"bool":   true,
		"smap": map[string]string{
			"a": "A",
			"b": "B",
		},
		"imap": map[string]interface{}{
			"c":     "C",
			"three": 3,
			"mapix": map[string]interface{}{
				"one": 1,
				"two": "too",
			},
			"iimap": map[string]interface{}{
				"d": "D",
				"ismap": map[string]string{
					"e": "E",
					"f": "F",
				},
			},
		},
	}

	exp, err := json.Marshal(am)
	r.NoError(err)
	r.Equal(string(exp), string(act))
}
