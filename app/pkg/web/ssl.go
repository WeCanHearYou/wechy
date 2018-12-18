package web

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"net/http"
	"strings"

	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/goenning/sqlcertcache"
	"golang.org/x/crypto/acme/autocert"
)

// ErrInvalidServerName is returned when SNI server name is invalid
var ErrInvalidServerName = errors.New("ssl: invalid server name")

func getDefaultTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		MaxVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		},
	}
}

//CertificateManager is used to manage SSL certificates
type CertificateManager struct {
	cert    tls.Certificate
	leaf    *x509.Certificate
	autossl autocert.Manager
}

//NewCertificateManager creates a new CertificateManager
func NewCertificateManager(certFile, keyFile string, conn *sql.DB) (*CertificateManager, error) {
	cache, err := sqlcertcache.New(conn, "autocert_cache")
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize new sqlcertcache")
	}

	manager := &CertificateManager{
		autossl: autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  cache,
		},
	}

	if certFile != "" && keyFile != "" {
		var err error
		manager.cert, err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load X509KeyPair for %s and %s", certFile, keyFile)
		}

		manager.leaf, err = x509.ParseCertificate(manager.cert.Certificate[0])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse x509 certificate")
		}
	}

	return manager, nil
}

//GetCertificate decides which certificate to use
//It first tries to use loaded certificate for incoming request if it's compatible
//Otherwise fallsback to a automatically generated certificate by Let's Encrypt
func (m *CertificateManager) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {

	// Fail is ServerName is empty or does't contain a dot
	if hello.ServerName == "" || !strings.Contains(strings.Trim(hello.ServerName, "."), ".") {
		return nil, ErrInvalidServerName
	}

	if m.leaf != nil {
		if !env.IsSingleHostMode() {
			subdomain := strings.TrimSuffix(hello.ServerName, env.MultiTenantDomain())
			if strings.Count(subdomain, ".") > 0 {
				return nil, ErrInvalidServerName
			}
		}

		if m.leaf.VerifyHostname(hello.ServerName) == nil {
			return &m.cert, nil
		}
	}

	return m.autossl.GetCertificate(hello)
}

//StartHTTPServer creates a new HTTP server on port 80 that is used for the ACME HTTP Challenge
func (m *CertificateManager) StartHTTPServer() {
	http.ListenAndServe(":80", m.autossl.HTTPHandler(nil))
}
