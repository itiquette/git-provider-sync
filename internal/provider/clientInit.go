// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package provider

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

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/archive"
	"itiquette/git-provider-sync/internal/provider/directory"
	"itiquette/git-provider-sync/internal/provider/gitea"
	"itiquette/git-provider-sync/internal/provider/github"
	"itiquette/git-provider-sync/internal/provider/gitlab"

	"github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

var (
	ErrNonSupportedProvider = errors.New("unsupported provider")
	ErrInvalidProxy         = errors.New("invalid proxy configuration")
	ErrCertificateLoad      = errors.New("failed to load certificates")
)

// ProxyFunc defines the type for proxy configuration functions.
type ProxyFunc func(req *http.Request) (*url.URL, error)

// NewGitProviderClient creates a new git provider client with improved error handling.
func NewGitProviderClient(ctx context.Context, opt model.GitProviderClientOption) (interfaces.GitProvider, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering NewGitProviderClient")

	httpClient, err := newHTTPClient(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Install git protocol handler
	client.InstallProtocol("https", githttp.NewClient(httpClient))

	return createProvider(ctx, opt, httpClient)
}

// createProvider handles provider creation with proper error handling.
func createProvider(ctx context.Context, opt model.GitProviderClientOption, httpClient *http.Client) (interfaces.GitProvider, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering createProvider")
	opt.DebugLog(logger).Msg("createProvider")

	var provider interfaces.GitProvider

	var err error

	switch opt.ProviderType {
	case config.GITEA:
		provider, err = gitea.NewGiteaAPIClient(ctx, httpClient, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gitea client: %w", err)
		}

	case config.GITHUB:
		provider, err = github.NewGitHubAPIClient(ctx, httpClient, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub client: %w", err)
		}

	case config.GITLAB:
		provider, err = gitlab.NewGitLabAPIClient(ctx, httpClient, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitLab client: %w", err)
		}

	case config.ARCHIVE:
		provider = archive.Client{}
	case config.DIRECTORY:
		provider = directory.Client{}
	default:
		return nil, fmt.Errorf("%w: %s", ErrNonSupportedProvider, opt.ProviderType)
	}

	logger.Debug().Str("name", provider.Name()).Msg("created provider client")

	return provider, nil
}

// newHTTPClient creates a new HTTP client with proper error handling.
func newHTTPClient(ctx context.Context, opt model.GitProviderClientOption) (*http.Client, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newHTTPClient")

	certPool, err := loadCertificates(ctx, opt.AuthCfg.CertDirPath)
	if err != nil {
		return nil, fmt.Errorf("certificate loading error: %w", err)
	}

	proxyFunc, err := setupProxy(ctx, opt.AuthCfg.ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("proxy setup error: %w", err)
	}

	tlsConfig := newTLSConfig(ctx, certPool)
	transport := newHTTPTransport(ctx, proxyFunc, tlsConfig)

	return &http.Client{
		Transport: transport,
		// Total timeout for entire request/response cycle
		Timeout: 30 * time.Second,
		// Limit redirect chains to prevent infinite loops
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}

			return nil
		},
	}, nil
}

// newHTTPTransport returns an http.Transport with production-ready default settings.
func newHTTPTransport(ctx context.Context, proxyFunc ProxyFunc, tlsConfig *tls.Config) *http.Transport {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newHTTPTransport")

	return &http.Transport{
		// Use proxy url setting or system proxy settings (HTTP_PROXY, HTTPS_PROXY)
		Proxy: proxyFunc,

		// Maximum time for TLS handshake - prevents hanging on SSL/TLS
		TLSHandshakeTimeout: 10 * time.Second,

		// Total number of idle connections across all hosts
		MaxIdleConns: 100,

		// Maximum idle connections per host
		MaxIdleConnsPerHost: 10,

		// Maximum total connections per host (idle + in-use)
		MaxConnsPerHost: 100,

		// How long to keep idle connections in pool before closing
		IdleConnTimeout: 90 * time.Second,

		// Time to wait for server's "100 Continue" response for large requests
		ExpectContinueTimeout: 1 * time.Second,

		TLSClientConfig: tlsConfig,

		// Connection settings including timeouts and keep-alive
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // Time limit for establishing TCP connection
			KeepAlive: 30 * time.Second, // Interval for TCP keepalive packets
			DualStack: true,             // Enable both IPv4 and IPv6
		}).DialContext,

		// Enable HTTP/2 support when available
		ForceAttemptHTTP2: true,

		// Buffer sizes for reading/writing - 4KB is good balance
		WriteBufferSize: 4 * 1024,
		ReadBufferSize:  4 * 1024,
	}
}

func newTLSConfig(ctx context.Context, caCertPool *x509.CertPool) *tls.Config {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newTLSConfig")

	return &tls.Config{
		RootCAs:                caCertPool,           // Custom CA cert pool for verification
		MinVersion:             tls.VersionTLS12,     // Minimum TLS version (good security practice)
		MaxVersion:             tls.VersionTLS13,     // Maximum TLS version
		ClientAuth:             tls.NoClientCert,     // No client certificate required
		Renegotiation:          tls.RenegotiateNever, // Disable renegotiation (security best practice)
		SessionTicketsDisabled: false,                // Disable session tickets for performance
		InsecureSkipVerify:     false,                // Ensure certificate verification
	}
}

// setupProxy configures the proxy.
func setupProxy(ctx context.Context, proxyURL string) (ProxyFunc, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering setupProxy")

	if proxyURL == "" {
		return http.ProxyFromEnvironment, nil
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidProxy, err)
	}

	return http.ProxyURL(parsedURL), nil
}

// loadCertificates loads certificates.
func loadCertificates(ctx context.Context, dirPath string) (*x509.CertPool, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering loadCertificates")

	if dirPath == "" {
		return nil, nil //nolint
	}

	caCertPool := x509.NewCertPool()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read directory: %w", ErrCertificateLoad, err)
	}

	for _, entry := range entries {
		if err := processCertificateFile(ctx, entry, dirPath, caCertPool); err != nil {
			return nil, err
		}
	}

	return caCertPool, nil
}

// processCertificateFile handles loading a single certificate file.
func processCertificateFile(ctx context.Context, entry os.DirEntry, dirPath string, pool *x509.CertPool) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering processCertificateFile")

	if !isCertFile(entry.Name()) {
		return nil
	}

	certPath := filepath.Join(dirPath, entry.Name())
	cert, err := os.ReadFile(certPath)

	if err != nil {
		return fmt.Errorf("%w: failed to read certificate %s: %w", ErrCertificateLoad, certPath, err)
	}

	if !pool.AppendCertsFromPEM(cert) {
		return fmt.Errorf("%w: failed to parse certificate %s", ErrCertificateLoad, certPath)
	}

	return nil
}

// isCertFile checks if the filename has a certificate extension.
func isCertFile(filename string) bool {
	ext := filepath.Ext(filename)

	return ext == ".crt" || ext == ".pem"
}
