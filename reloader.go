package httpreloader

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"
	"sync"
)

// Server custom http server with Reloader functions
type Server struct {
	*http.Server
	Reloader *reloader
}

var (
	ErrCertificateNotLoaded = errors.New("certificates is not loaded")
	ErrCertificateNotFound  = errors.New("certificate for domain is not found")
)

// NewServer return new server with Reloader
func NewServer(addr, certFile, keyFile string, handler http.Handler) (*Server, error) {
	var err error
	srv := new(Server)
	srv.Reloader, err = NewReloader(certFile, keyFile)
	srv.Server = &http.Server{
		Addr:    addr,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: srv.Reloader.GetCertificateFunc(),
		},
	}
	return srv, err
}

// ListenAndServeTLS replace function http.ListenAndServeTLS
func (srv *Server) ListenAndServeTLS() error {
	return srv.Server.ListenAndServeTLS("", "")
}

type reloader struct {
	mu                 sync.RWMutex
	defaultCertificate *tls.Certificate
	certs              map[string]*tls.Certificate
}

// NewReloader return new reloader
func NewReloader(certFile, keyFile string) (*reloader, error) {
	r := reloader{
		certs: map[string]*tls.Certificate{},
	}
	return &r, r.UpdateCertificate(certFile, keyFile)
}

// UpdateCertificate update or add certificate for domains.
// If domains is not specified- update or add default certificate.
// For wildcard domains, you need to use the prefix "*" (Ex *.domain.com)
func (r *reloader) UpdateCertificate(certFile, keyFile string, domain ...string) error {
	newCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(domain) == 0 {
		r.defaultCertificate = &newCert
		return nil
	}

	for i := range domain {
		r.certs[normalizeDomainName(domain[i])] = &newCert
	}
	return nil
}

// RemoveCertificate removes the certificate for the domain from use
func (r *reloader) RemoveCertificate(domain string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.certs[domain]; !ok {
		return ErrCertificateNotFound
	}
	delete(r.certs, domain)
	return nil
}

// GetCertificateFunc return func for use in http.Server-TLSConfig-GetCertificate
func (r *reloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if cert := r.findCertificate(normalizeDomainName(clientHello.ServerName)); cert != nil {
			return cert, nil
		}
		if r.defaultCertificate != nil {
			return r.defaultCertificate, nil
		}
		return &tls.Certificate{}, ErrCertificateNotLoaded
	}
}

func (r *reloader) findCertificate(domain string) *tls.Certificate {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for k, v := range r.certs {
		if strings.Compare(k, domain) == 0 {
			return v
		}
		if strings.HasPrefix(k, "*.") && strings.HasSuffix(domain, k[1:]) {
			return v

		}
	}
	return nil
}

func normalizeDomainName(d string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(d), "."))
}
