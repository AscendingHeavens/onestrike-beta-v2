package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	// Setup
	router := NewRouter()
	testHandler := func(c *Context) *Response {
		return &Response{Success: true, Message: "ok", Code: 200}
	}

	router.Handle("GET", "/users", testHandler)
	router.Handle("GET", "/users/:id", testHandler)
	router.Handle("POST", "/users/:id/update", testHandler)

	tests := []struct {
		name       string
		method     string
		path       string
		wantFound  bool
		wantParams map[string]string
	}{
		{
			"static route match",
			"GET",
			"/users",
			true,
			map[string]string{},
		},
		{
			"parameter route match",
			"GET",
			"/users/123",
			true,
			map[string]string{"id": "123"},
		},
		{
			"parameter route with post method",
			"POST",
			"/users/456/update",
			true,
			map[string]string{"id": "456"},
		},
		{
			"wrong method",
			"POST",
			"/users",
			false,
			nil,
		},
		{
			"no route found",
			"GET",
			"/unknown",
			false,
			nil,
		},
		{
			"path length mismatch",
			"GET",
			"/users/123/extra",
			false,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, params := router.FindHandler(tt.method, tt.path)
			if tt.wantFound {
				assert.NotNil(t, h)
				assert.Equal(t, tt.wantParams, params)
			} else {
				assert.Nil(t, h)
				assert.Nil(t, params)
			}
		})
	}
}
