package onestrike

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AscendingHeavens/onestrike/server"
	"github.com/stretchr/testify/assert"
)

func TestServer_HandleAndServeHTTP(t *testing.T) {
	s := New()

	// global middleware that adds header
	globalCalled := false
	s.Use(func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			globalCalled = true
			return next(c)
		}
	})

	handlerCalled := false
	handler := func(c *server.Context) *server.Response {
		handlerCalled = true
		return &server.Response{Success: true, Message: "ok", Code: 200}
	}

	s.Handle(http.MethodGet, "/test", handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.True(t, globalCalled)
	assert.True(t, handlerCalled)
	assert.Equal(t, 200, rec.Code)

	var resp server.Response
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp.Message)
}

func TestServer_404(t *testing.T) {
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)
	assert.Equal(t, 404, rec.Code)
}

func TestServer_ConditionalMiddleware(t *testing.T) {
	s := New()

	called := false
	s.UseIf("/api/*", func(next server.HandlerFunc) server.HandlerFunc {
		return func(c *server.Context) *server.Response {
			called = true
			return next(c)
		}
	})

	handlerCalled := false
	handler := func(c *server.Context) *server.Response {
		handlerCalled = true
		return &server.Response{Success: true, Message: "ok", Code: 200}
	}

	s.Handle(http.MethodGet, "/api/test", handler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.True(t, handlerCalled)
	assert.Equal(t, 200, rec.Code)
}

func TestServer_HTTPMethods(t *testing.T) {
	methods := []struct {
		name     string
		register func(s *Server, path string, h server.HandlerFunc)
		method   string
		path     string
	}{
		{"GET", func(s *Server, p string, h server.HandlerFunc) { s.GET(p, h) }, http.MethodGet, "/get"},
		{"POST", func(s *Server, p string, h server.HandlerFunc) { s.POST(p, h) }, http.MethodPost, "/post"},
		{"PUT", func(s *Server, p string, h server.HandlerFunc) { s.PUT(p, h) }, http.MethodPut, "/put"},
		{"PATCH", func(s *Server, p string, h server.HandlerFunc) { s.PATCH(p, h) }, http.MethodPatch, "/patch"},
		{"DELETE", func(s *Server, p string, h server.HandlerFunc) { s.DELETE(p, h) }, http.MethodDelete, "/delete"},
	}

	for _, tt := range methods {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			handlerCalled := false
			handler := func(c *server.Context) *server.Response {
				handlerCalled = true
				return &server.Response{Success: true, Message: "ok", Code: 200}
			}

			// Register route
			tt.register(s, tt.path, handler)

			// Make request
			req := httptest.NewRequest(tt.method, tt.path, nil)
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
