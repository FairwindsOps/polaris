package genny

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DryRunner_Request(t *testing.T) {
	table := []struct {
		code   int
		method string
		path   string
		boom   bool
	}{
		{200, "GET", "/a", false},
		{200, "POST", "/b", false},
		{399, "PATCH", "/c", false},
		{401, "GET", "/d", true},
		{500, "POST", "/e", true},
	}

	for _, tt := range table {
		t.Run(tt.method+tt.path, func(st *testing.T) {
			r := require.New(st)

			req, err := http.NewRequest(tt.method, tt.path, nil)
			r.NoError(err)
			run := DryRunner(context.Background())

			run.RequestFn = func(req *http.Request, c *http.Client) (*http.Response, error) {
				if tt.boom {
					return nil, fmt.Errorf("error %d", tt.code)
				}
				return &http.Response{StatusCode: tt.code}, nil
			}

			res, err := run.Request(req)
			if tt.boom {
				r.Error(err)
				return
			}
			r.NoError(err)
			r.Equal(tt.code, res.StatusCode)
		})
	}
}
