package autotls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

type AutoTLSConfig struct {
	Domains  []string `json:"domains"`
	Email    string   `json:"email"`
	CacheDir string   `json:"cache_dir"`
	EnableH3 bool     `json:"enable_h3"`
	IsDev    bool     `json:"is_dev"`
}

type AutoTLSManager struct {
	config  AutoTLSConfig
	manager *autocert.Manager
}

func NewAutoTLSManager(cfg AutoTLSConfig) *AutoTLSManager {
	if cfg.CacheDir == "" {
		cfg.CacheDir = ".certs_cache"
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Domains...),
		Cache:      autocert.DirCache(cfg.CacheDir),
		Email:      cfg.Email,
	}

	return &AutoTLSManager{
		config:  cfg,
		manager: m,
	}
}

func (a *AutoTLSManager) GetTLSConfig() (*tls.Config, error) {
	if a.config.IsDev {
		return a.GenerateSelfSignedCert()
	}

	tlsCfg := a.manager.TLSConfig()
	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}
	tlsCfg.NextProtos = []string{"h2", "http/1.1"}
	if a.config.EnableH3 {
		tlsCfg.NextProtos = append([]string{"h3"}, tlsCfg.NextProtos...)
	}
	return tlsCfg, nil
}

func (a *AutoTLSManager) GenerateSelfSignedCert() (*tls.Config, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ServGateway Dev AutoTLS"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost", "servgateway.local"},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load dev X509 key pair: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2", "http/1.1"},
	}, nil
}

func ListenAndServeTLS(server *http.Server, tlsMgr *AutoTLSManager) error {
	tlsCfg, err := tlsMgr.GetTLSConfig()
	if err != nil {
		return err
	}

	server.TLSConfig = tlsCfg
	return server.ListenAndServeTLS("", "")
}
