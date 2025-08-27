// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package transport provides HTTP client factory and transport utilities for Git Provider Sync.
//
//nolint:funcorder // HTTP client factory with many helper methods
package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

var (
	// ErrInvalidProxy indicates invalid proxy configuration.
	ErrInvalidProxy = errors.New("invalid proxy configuration")
	// ErrCertificateLoad indicates failure to load certificates.
	ErrCertificateLoad = errors.New("failed to load certificates")
)

// ProxyFunc defines the type for proxy configuration functions.
type ProxyFunc func(req *http.Request) (*url.URL, error)

// HTTPClientFactory provides sophisticated HTTP client creation with advanced features.
type HTTPClientFactory struct {
	logger ports.Logger
}

// NewHTTPClientFactory creates a new HTTP client factory.
func NewHTTPClientFactory(logger ports.Logger) *HTTPClientFactory {
	return &HTTPClientFactory{
		logger: logger,
	}
}

// HTTPClientConfig contains configuration for HTTP client creation.
type HTTPClientConfig struct {
	// Authentication and security
	CertDirPath string
	ProxyURL    string

	// Timeout settings
	RequestTimeout int // seconds

	// TLS settings
	InsecureSkipVerify bool
	MinTLSVersion      uint16
	MaxTLSVersion      uint16

	// Connection settings
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration

	// Dial settings
	DialTimeout time.Duration
	KeepAlive   time.Duration
	DualStack   bool

	// HTTP/2 settings
	ForceAttemptHTTP2 bool

	// Buffer settings
	WriteBufferSize int
	ReadBufferSize  int
}

// DefaultHTTPClientConfig returns a configuration with production-ready defaults.
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		RequestTimeout:      30,
		MinTLSVersion:       tls.VersionTLS12,
		MaxTLSVersion:       tls.VersionTLS13,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		DialTimeout:         30 * time.Second,
		KeepAlive:           30 * time.Second,
		DualStack:           true,
		ForceAttemptHTTP2:   true,
		WriteBufferSize:     4 * 1024,
		ReadBufferSize:      4 * 1024,
	}
}

// CreateHTTPClient creates a new HTTP client with sophisticated configuration.
func (f *HTTPClientFactory) CreateHTTPClient(ctx context.Context, config HTTPClientConfig) (*http.Client, error) {
	f.logger.Debug(ctx, "Creating HTTP client with advanced configuration", map[string]interface{}{
		"cert_dir_path":   config.CertDirPath,
		"has_proxy":       config.ProxyURL != "",
		"request_timeout": config.RequestTimeout,
		"min_tls_version": config.MinTLSVersion,
		"max_tls_version": config.MaxTLSVersion,
	})

	certPool, err := f.loadCertificates(ctx, config.CertDirPath)
	if err != nil {
		return nil, fmt.Errorf("certificate loading error: %w", err)
	}

	proxyFunc, err := f.setupProxy(ctx, config.ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("proxy setup error: %w", err)
	}

	tlsConfig := f.createTLSConfig(ctx, certPool, config)
	transport := f.createHTTPTransport(ctx, proxyFunc, tlsConfig, config)

	client := &http.Client{
		Transport: transport,
		// Total timeout for entire request/response cycle
		Timeout: time.Duration(config.RequestTimeout) * time.Second,
		// Limit redirect chains to prevent infinite loops
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}

			return nil
		},
	}

	f.logger.Info(ctx, "HTTP client created successfully", map[string]interface{}{
		"cert_pool_configured": certPool != nil,
		"proxy_configured":     config.ProxyURL != "",
		"timeout_seconds":      config.RequestTimeout,
	})

	return client, nil
}

// createHTTPTransport creates an HTTP transport with production-ready settings.
func (f *HTTPClientFactory) createHTTPTransport(
	ctx context.Context,
	proxyFunc ProxyFunc,
	tlsConfig *tls.Config,
	config HTTPClientConfig,
) *http.Transport {
	f.logger.Debug(ctx, "Creating HTTP transport with advanced settings", map[string]interface{}{
		"max_idle_conns":          config.MaxIdleConns,
		"max_idle_conns_per_host": config.MaxIdleConnsPerHost,
		"max_conns_per_host":      config.MaxConnsPerHost,
		"idle_conn_timeout":       config.IdleConnTimeout.String(),
		"dial_timeout":            config.DialTimeout.String(),
		"keep_alive":              config.KeepAlive.String(),
		"dual_stack":              config.DualStack,
		"force_http2":             config.ForceAttemptHTTP2,
	})

	return &http.Transport{
		// Use proxy url setting or system proxy settings (HTTP_PROXY, HTTPS_PROXY)
		Proxy: proxyFunc,

		// Maximum time for TLS handshake - prevents hanging on SSL/TLS
		TLSHandshakeTimeout: 10 * time.Second,

		// Total number of idle connections across all hosts
		MaxIdleConns: config.MaxIdleConns,

		// Maximum idle connections per host
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,

		// Maximum total connections per host (idle + in-use)
		MaxConnsPerHost: config.MaxConnsPerHost,

		// How long to keep idle connections in pool before closing
		IdleConnTimeout: config.IdleConnTimeout,

		// Time to wait for server's "100 Continue" response for large requests
		ExpectContinueTimeout: 1 * time.Second,

		TLSClientConfig: tlsConfig,

		// Connection settings including timeouts and keep-alive
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout, // Time limit for establishing TCP connection
			KeepAlive: config.KeepAlive,   // Interval for TCP keepalive packets
			DualStack: config.DualStack,   // Enable both IPv4 and IPv6
		}).DialContext,

		// Enable HTTP/2 support when available
		ForceAttemptHTTP2: config.ForceAttemptHTTP2,

		// Buffer sizes for reading/writing - 4KB is good balance
		WriteBufferSize: config.WriteBufferSize,
		ReadBufferSize:  config.ReadBufferSize,
	}
}

// createTLSConfig creates a TLS configuration with security best practices.
func (f *HTTPClientFactory) createTLSConfig(
	ctx context.Context,
	caCertPool *x509.CertPool,
	config HTTPClientConfig,
) *tls.Config {
	f.logger.Debug(ctx, "Creating TLS configuration", map[string]interface{}{
		"has_custom_ca":        caCertPool != nil,
		"min_tls_version":      config.MinTLSVersion,
		"max_tls_version":      config.MaxTLSVersion,
		"insecure_skip_verify": config.InsecureSkipVerify,
	})

	return &tls.Config{
		RootCAs:                caCertPool,                // Custom CA cert pool for verification
		MinVersion:             config.MinTLSVersion,      // Minimum TLS version (good security practice)
		MaxVersion:             config.MaxTLSVersion,      // Maximum TLS version
		ClientAuth:             tls.NoClientCert,          // No client certificate required
		Renegotiation:          tls.RenegotiateNever,      // Disable renegotiation (security best practice)
		SessionTicketsDisabled: false,                     // Enable session tickets for performance
		InsecureSkipVerify:     config.InsecureSkipVerify, // #nosec G402 - Configurable for corporate environments
	}
}

// setupProxy configures the proxy function.
func (f *HTTPClientFactory) setupProxy(ctx context.Context, proxyURL string) (ProxyFunc, error) {
	f.logger.Debug(ctx, "Setting up proxy configuration", map[string]interface{}{
		"proxy_url": proxyURL,
	})

	if proxyURL == "" {
		f.logger.Debug(ctx, "No proxy URL provided, using environment proxy settings", nil)

		return http.ProxyFromEnvironment, nil
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidProxy, err)
	}

	f.logger.Info(ctx, "Proxy configured successfully", map[string]interface{}{
		"proxy_scheme": parsedURL.Scheme,
		"proxy_host":   parsedURL.Host,
	})

	return http.ProxyURL(parsedURL), nil
}

// loadCertificates loads custom CA certificates from a directory.
func (f *HTTPClientFactory) loadCertificates(ctx context.Context, dirPath string) (*x509.CertPool, error) {
	f.logger.Debug(ctx, "Loading custom certificates", map[string]interface{}{
		"cert_dir_path": dirPath,
	})

	if dirPath == "" {
		f.logger.Debug(ctx, "No certificate directory provided, using system certificates", nil)
		// Return system cert pool when no custom directory is provided
		systemPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("%w: failed to load system certificates: %w", ErrCertificateLoad, err)
		}

		return systemPool, nil
	}

	caCertPool := x509.NewCertPool()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read directory: %w", ErrCertificateLoad, err)
	}

	certCount := 0

	for _, entry := range entries {
		if err := f.processCertificateFile(ctx, entry, dirPath, caCertPool); err != nil {
			return nil, err
		}

		if f.isCertFile(entry.Name()) {
			certCount++
		}
	}

	f.logger.Info(ctx, "Custom certificates loaded successfully", map[string]interface{}{
		"cert_dir_path": dirPath,
		"cert_count":    certCount,
	})

	return caCertPool, nil
}

// processCertificateFile handles loading a single certificate file.
func (f *HTTPClientFactory) processCertificateFile(
	ctx context.Context,
	entry os.DirEntry,
	dirPath string,
	pool *x509.CertPool,
) error {
	if !f.isCertFile(entry.Name()) {
		return nil
	}

	certPath := filepath.Join(dirPath, entry.Name())

	f.logger.Debug(ctx, "Loading certificate file", map[string]interface{}{
		"cert_path": certPath,
	})

	// #nosec G304 - Certificate path is from configuration
	cert, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("%w: failed to read certificate %s: %w", ErrCertificateLoad, certPath, err)
	}

	if !pool.AppendCertsFromPEM(cert) {
		return fmt.Errorf("%w: failed to parse certificate %s", ErrCertificateLoad, certPath)
	}

	f.logger.Debug(ctx, "Certificate loaded successfully", map[string]interface{}{
		"cert_path": certPath,
	})

	return nil
}

// isCertFile checks if the filename has a certificate extension.
func (f *HTTPClientFactory) isCertFile(filename string) bool {
	ext := filepath.Ext(filename)

	return ext == ".crt" || ext == ".pem"
}

// CreateGitTransportClient creates an HTTP client optimized for Git operations.
func (f *HTTPClientFactory) CreateGitTransportClient(ctx context.Context, config HTTPClientConfig) (*http.Client, error) {
	// Git operations need longer timeouts and more lenient settings
	gitConfig := config
	gitConfig.RequestTimeout = 300 // 5 minutes for large repos
	gitConfig.MaxIdleConns = 50
	gitConfig.MaxIdleConnsPerHost = 5
	gitConfig.IdleConnTimeout = 60 * time.Second

	httpClient, err := f.CreateHTTPClient(ctx, gitConfig)
	if err != nil {
		return nil, err
	}

	// Install git protocol handler
	client.InstallProtocol("https", githttp.NewClient(httpClient))

	f.logger.Info(ctx, "Git transport client created with protocol registration", map[string]interface{}{
		"timeout":  gitConfig.RequestTimeout,
		"protocol": "https",
	})

	return httpClient, nil
}

// CreateAPIClient creates an HTTP client optimized for API operations.
func (f *HTTPClientFactory) CreateAPIClient(ctx context.Context, config HTTPClientConfig) (*http.Client, error) {
	// API operations need faster timeouts and more connections
	apiConfig := config
	apiConfig.RequestTimeout = 30 // 30 seconds for API calls
	apiConfig.MaxIdleConns = 100
	apiConfig.MaxIdleConnsPerHost = 20
	apiConfig.IdleConnTimeout = 90 * time.Second

	return f.CreateHTTPClient(ctx, apiConfig)
}

// ValidateConfig validates the HTTP client configuration.
func (f *HTTPClientFactory) ValidateConfig(config HTTPClientConfig) error {
	if err := f.validateTimeouts(config); err != nil {
		return err
	}

	if err := f.validateConnectionLimits(config); err != nil {
		return err
	}

	if err := f.validatePaths(config); err != nil {
		return err
	}

	return nil
}

// validateTimeouts validates timeout-related configuration.
func (f *HTTPClientFactory) validateTimeouts(config HTTPClientConfig) error {
	if config.RequestTimeout <= 0 {
		return domain.ErrRequestTimeoutMustBePositive
	}

	if config.IdleConnTimeout <= 0 {
		return domain.ErrIdleConnectionTimeoutMustBePositive
	}

	if config.DialTimeout <= 0 {
		return domain.ErrDialTimeoutMustBePositive
	}

	if config.KeepAlive <= 0 {
		return domain.ErrKeepAliveMustBePositive
	}

	return nil
}

// validateConnectionLimits validates connection limit configuration.
func (f *HTTPClientFactory) validateConnectionLimits(config HTTPClientConfig) error {
	if config.MaxIdleConns <= 0 {
		return domain.ErrMaxIdleConnectionsMustBePositive
	}

	if config.MaxIdleConnsPerHost <= 0 {
		return domain.ErrMaxIdleConnectionsPerHostMustBePositive
	}

	if config.MaxConnsPerHost <= 0 {
		return domain.ErrMaxConnectionsPerHostMustBePositive
	}

	return nil
}

// validatePaths validates path and URL configuration.
func (f *HTTPClientFactory) validatePaths(config HTTPClientConfig) error {
	if config.CertDirPath != "" {
		if _, err := os.Stat(config.CertDirPath); os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", domain.ErrCertificateDirectoryNotExist, config.CertDirPath)
		}
	}

	if config.ProxyURL != "" {
		if _, err := url.Parse(config.ProxyURL); err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
	}

	return nil
}
