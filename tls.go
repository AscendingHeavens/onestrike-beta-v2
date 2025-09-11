package onestrike

import (
	"crypto/tls"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

// These are for dependency injection in tests
var listenAndServeTLS = http.ListenAndServeTLS
var logFatal = log.Fatal

func (s *Server) StartTLS(addr, certFile, keyFile string) {
	log.Printf("Starting server with TLS on %s", addr)
	if err := listenAndServeTLS(addr, certFile, keyFile, s); err != nil {
		logFatal(err)
	}
}

func (s *Server) StartAutoTLS(domain string) {
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
	logFatal(server.ListenAndServeTLS("", ""))
}
