package onestrike

import (
	"crypto/tls"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

// listenAndServeTLS is a package-level variable that wraps http.ListenAndServeTLS
// for dependency injection during testing. This allows tests to mock the TLS
// server startup without actually starting a real server.
var listenAndServeTLS = func(addr, certFile, keyFile string, h http.Handler) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, h)
}

// logFatal is a package-level variable that wraps log.Fatal for dependency
// injection during testing. This allows tests to capture fatal errors without
// actually terminating the test process.
var logFatal = func(v ...interface{}) {
	log.Fatal(v...)
}

// StartTLS starts the server with TLS using the provided certificate and key files.
// The server will bind to the specified address and serve HTTPS traffic.
//
// Parameters:
//   - addr: The address to bind to (e.g., ":443", "localhost:8443")
//   - certFile: Path to the TLS certificate file
//   - keyFile: Path to the TLS private key file
//
// This method will call log.Fatal if the server fails to start, terminating
// the program. Use this for production deployments where server startup
// failure should halt the application.
//
// Example:
//
//	server := &Server{}
//	server.StartTLS(":443", "/path/to/cert.pem", "/path/to/key.pem")
func (s *Server) StartTLS(addr, certFile, keyFile string) {
	log.Printf("Starting server with TLS on %s", addr)
	if err := listenAndServeTLS(addr, certFile, keyFile, s); err != nil {
		logFatal(err)
	}
}

// startTLSServer starts an HTTP server with TLS using the provided server configuration.
// This is an internal method that implements the TLSStarter interface, allowing
// the Server to start itself when used with StartAutoTLSWithStarter.
//
// This method assumes the server is already configured with TLS settings and
// will call log.Fatal if the server fails to start.
func (s *Server) startTLSServer(server *http.Server) {
	logFatal(server.ListenAndServeTLS("", ""))
}

// StartAutoTLS starts the server with automatic TLS certificate management using Let's Encrypt.
// This method automatically obtains and renews TLS certificates for the specified domain
// using the ACME protocol. The server will bind to port 443.
//
// Parameters:
//   - domain: The domain name for which to obtain certificates (e.g., "example.com")
//
// The method sets up:
//   - Automatic certificate cache in a local "certs" directory
//   - Automatic acceptance of Let's Encrypt Terms of Service
//   - Host whitelist policy for the specified domain
//   - TLS configuration with automatic certificate retrieval
//
// Requirements:
//   - The server must be accessible from the internet on port 443
//   - The domain must point to the server's IP address
//   - Port 80 should also be available for ACME challenges (handled automatically by autocert)
//
// This method will call log.Fatal if the server fails to start, terminating
// the program. Use this for production deployments where server startup
// failure should halt the application.
//
// Example:
//
//	server := &Server{}
//	server.StartAutoTLS("example.com") // Will serve HTTPS on port 443
func (s *Server) StartAutoTLS(domain string) {
	s.StartAutoTLSWithStarter(domain, s)
}

// StartAutoTLSWithStarter starts the server with automatic TLS certificate management
// using a custom TLS starter. This method provides the same automatic certificate
// functionality as StartAutoTLS but allows dependency injection of the server
// startup mechanism, making it testable.
//
// Parameters:
//   - domain: The domain name for which to obtain certificates (e.g., "example.com")
//   - starter: An implementation of TLSStarter interface that handles server startup
//
// This method is primarily intended for testing purposes where you need to mock
// the server startup behavior. For production use, prefer StartAutoTLS which
// uses the server's own startup mechanism.
//
// The method configures:
//   - autocert.Manager with local certificate caching
//   - Automatic TOS acceptance for Let's Encrypt
//   - Host policy restricting certificates to the specified domain
//   - HTTP server bound to port 443 with TLS configuration
//
// Example:
//
//	server := &Server{}
//	mockStarter := &MockTLSStarter{...}
//	server.StartAutoTLSWithStarter("example.com", mockStarter)
func (s *Server) StartAutoTLSWithStarter(domain string, starter TLSStarter) {
	manager := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
	}

	server := &http.Server{
		Addr:      ":443",
		Handler:   s,
		TLSConfig: &tls.Config{GetCertificate: manager.GetCertificate},
	}

	log.Printf("Starting server with AutoTLS on %s", domain)
	starter.startTLSServer(server)
}
