package onestrike

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rishi-Mishra0704/OneStrike/server"
	"github.com/stretchr/testify/assert"
)

func TestGroup_Handle(t *testing.T) {
	testServer := New()
	testServer.router = server.NewRouter() // ensure router is initialized

	// Middleware that adds a marker to the response
	mw1 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) *Response {
			c.Writer.Header().Set("X-MW1", "1")
			return next(c)
		}
	}
	mw2 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) *Response {
			c.Writer.Header().Set("X-MW2", "1")
			return next(c)
		}
	}

	// Apply server-level middleware
	testServer.Use(mw1)

	group := testServer.Group("/api/v1")
	group.Use(mw2)

	handlerCalled := false
	testHandler := func(c *Context) *Response {
		handlerCalled = true
		return &Response{Success: true, Message: "ok", Code: 200}
	}

	group.Handle(http.MethodGet, "/test", testHandler)

	// Verify route registered with prefixed path
	h, params := testServer.router.FindHandler(http.MethodGet, "/api/v1/test")
	assert.NotNil(t, h)
	assert.Empty(t, params)

	// Simulate a request to check middleware execution
	rec := httptest.NewRecorder()
	c := &Context{Writer: rec}
	resp := h(c)
	assert.True(t, handlerCalled)
	assert.Equal(t, "1", rec.Header().Get("X-MW1")) // server-level middleware
	assert.Equal(t, "1", rec.Header().Get("X-MW2")) // group-level middleware
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "ok", resp.Message)
}

func TestGroup_HTTPMethods(t *testing.T) {
	methods := []struct {
		name     string
		register func(g *Group, path string, h server.HandlerFunc)
		method   string
		path     string
	}{
		{"GET", func(g *Group, p string, h server.HandlerFunc) { g.GET(p, h) }, http.MethodGet, "/get"},
		{"POST", func(g *Group, p string, h server.HandlerFunc) { g.POST(p, h) }, http.MethodPost, "/post"},
		{"PUT", func(g *Group, p string, h server.HandlerFunc) { g.PUT(p, h) }, http.MethodPut, "/put"},
		{"PATCH", func(g *Group, p string, h server.HandlerFunc) { g.PATCH(p, h) }, http.MethodPatch, "/patch"},
		{"DELETE", func(g *Group, p string, h server.HandlerFunc) { g.DELETE(p, h) }, http.MethodDelete, "/delete"},
	}

	for _, tt := range methods {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			group := s.Group("/api/v1")
			handlerCalled := false
			handler := func(c *server.Context) *server.Response {
				handlerCalled = true
				return &server.Response{Success: true, Message: "ok", Code: 200}
			}

			// Register route using group helper
			tt.register(group, tt.path, handler)

			// Make request
			fullPath := "/api/v1" + tt.path
			req := httptest.NewRequest(tt.method, fullPath, nil)
			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)

			assert.True(t, handlerCalled, "handler should be called")
			assert.Equal(t, 200, rec.Code)

			var resp server.Response
			err := json.NewDecoder(rec.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, "ok", resp.Message)
		})
	}
}
