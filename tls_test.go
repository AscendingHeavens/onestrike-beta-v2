package onestrike

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartTLS_CallsListenAndServeTLS(t *testing.T) {
	called := false
	var addr, cert, key string

	// mock ListenAndServeTLS
	listenAndServeTLS = func(a, c, k string, h http.Handler) error {
		called = true
		addr = a
		cert = c
		key = k
		return nil
	}

	// logFatal should not be called
	logFatal = func(v ...interface{}) {}

	srv := &Server{}
	srv.StartTLS("127.0.0.1:8443", "cert.pem", "key.pem")

	assert.True(t, called)
	assert.Equal(t, "127.0.0.1:8443", addr)
	assert.Equal(t, "cert.pem", cert)
	assert.Equal(t, "key.pem", key)
}

func TestStartTLS_ListenAndServeTLSError(t *testing.T) {
	// simulate error
	listenAndServeTLS = func(a, c, k string, h http.Handler) error {
		return errors.New("tls error")
	}

	called := false
	logFatal = func(v ...interface{}) {
		called = true
		assert.Contains(t, v[0].(error).Error(), "tls error")
	}

	srv := &Server{}
	srv.StartTLS("127.0.0.1:8443", "cert.pem", "key.pem")
	assert.True(t, called)
}

// Test with mock starter
func TestStartAutoTLS_CallsListenAndServeTLS(t *testing.T) {
	called := false
	var capturedServer *http.Server

	// Create a mock starter
	mockStarter := &mockTLSStarter{
		startFunc: func(server *http.Server) {
			called = true
			capturedServer = server
		},
	}

	srv := &Server{}

	// Run StartAutoTLS with the mock starter
	srv.StartAutoTLSWithStarter("example.com", mockStarter)

	// Verify the server was configured correctly
	assert.True(t, called, "startTLSServer should have been called")
	assert.NotNil(t, capturedServer, "server should not be nil")
	assert.Equal(t, ":443", capturedServer.Addr, "server should listen on port 443")
	assert.NotNil(t, capturedServer.TLSConfig, "TLS config should be set")
	assert.NotNil(t, capturedServer.TLSConfig.GetCertificate, "GetCertificate should be set")
}

// Mock implementation for testing
type mockTLSStarter struct {
	startFunc func(*http.Server)
}

func (m *mockTLSStarter) startTLSServer(server *http.Server) {
	if m.startFunc != nil {
		m.startFunc(server)
	}
}
