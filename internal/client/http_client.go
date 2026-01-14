package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/davidshato/terraform-provider-httpx/internal/config"
)

// HTTPClient wraps an http.Client with provider configuration
type HTTPClient struct {
	client  *http.Client
	config  *config.ProviderConfig
	timeout time.Duration
}

// NewHTTPClient creates a new HTTP client from provider configuration
func NewHTTPClient(cfg *config.ProviderConfig) (*HTTPClient, error) {
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond

	// Create TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec // User-configurable option for testing/development
	}

	// Configure TLS certificates if provided
	if cfg.CaCertPem != nil && *cfg.CaCertPem != "" {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(*cfg.CaCertPem)) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	if cfg.ClientCertPem != nil && cfg.ClientKeyPem != nil {
		cert, err := tls.X509KeyPair([]byte(*cfg.ClientCertPem), []byte(*cfg.ClientKeyPem))
		if err != nil {
			return nil, fmt.Errorf("failed to parse client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Create transport
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Configure proxy if provided
	if cfg.ProxyUrl != nil && *cfg.ProxyUrl != "" {
		proxyURL, err := url.Parse(*cfg.ProxyUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// Create HTTP client
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &HTTPClient{
		client:  httpClient,
		config:  cfg,
		timeout: timeout,
	}, nil
}

// Do executes an HTTP request
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// GetTimeout returns the configured timeout
func (c *HTTPClient) GetTimeout() time.Duration {
	return c.timeout
}

// LimitReader wraps an io.Reader to limit the number of bytes read
func LimitReader(r io.Reader, maxBytes int64) io.Reader {
	return io.LimitReader(r, maxBytes)
}
