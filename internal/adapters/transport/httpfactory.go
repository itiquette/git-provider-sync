// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package transport

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"itiquette/git-provider-sync/internal/domain"
)

const (
	// ProviderTypeGitHub represents GitHub provider type.
	ProviderTypeGitHub = "github"
	// ProviderTypeGitLab represents GitLab provider type.
	ProviderTypeGitLab = "gitlab"
	// ProviderTypeGitea represents Gitea provider type.
	ProviderTypeGitea = "gitea"
)

// HTTPFactory creates and configures HTTP clients for various providers and operations.
type HTTPFactory struct {
	defaultTimeout  time.Duration
	maxRetries      int
	retryDelay      time.Duration
	skipTLSVerify   bool
	proxyURL        *url.URL
	customHeaders   map[string]string
	rateLimitConfig RateLimitConfig
}

// HTTPConfig contains configuration for HTTP clients.
type HTTPConfig struct {
	Timeout       time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
	SkipTLSVerify bool
	ProxyURL      string
	Headers       map[string]string
	RateLimit     RateLimitConfig
	UserAgent     string
}

// RateLimitConfig contains rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	Enable            bool
}

// HTTPClientOptions contains options for creating HTTP clients.
type HTTPClientOptions struct {
	Provider      string
	Authenticated bool
	Token         string
	Username      string
	Password      string
	Custom        map[string]any
}

// NewHTTPFactory creates a new HTTP factory with default configuration.
func NewHTTPFactory(config HTTPConfig) (*HTTPFactory, error) {
	var proxyURL *url.URL

	if config.ProxyURL != "" {
		parsed, err := url.Parse(config.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}

		proxyURL = parsed
	}

	// Security: TLS verification bypass is not allowed - always verify TLS certificates
	if config.SkipTLSVerify {
		return nil, errors.New("TLS verification bypass is not permitted for security reasons")
	}

	return &HTTPFactory{
		defaultTimeout:  config.Timeout,
		maxRetries:      config.MaxRetries,
		retryDelay:      config.RetryDelay,
		skipTLSVerify:   false, // Always enforce TLS verification
		proxyURL:        proxyURL,
		customHeaders:   config.Headers,
		rateLimitConfig: config.RateLimit,
	}, nil
}

// CreateClient creates an HTTP client with the specified options.
func (f *HTTPFactory) CreateClient(options HTTPClientOptions) (*http.Client, error) {
	// Create retryable HTTP client
	retryClient := retryablehttp.NewClient()

	// Configure retry policy
	retryClient.RetryMax = f.maxRetries
	retryClient.RetryWaitMin = f.retryDelay
	retryClient.RetryWaitMax = f.retryDelay * 5

	// Configure backoff strategy
	retryClient.Backoff = f.exponentialBackoff

	// Configure retry conditions
	retryClient.CheckRetry = f.retryPolicy

	// Disable default logging if needed
	retryClient.Logger = nil

	// Create base HTTP client
	httpClient := &http.Client{
		Timeout: f.defaultTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Always verify TLS certificates for security
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12, // Secure minimum TLS version
			},
			Proxy: f.getProxy,
		},
	}

	// Set the HTTP client for retryable client
	retryClient.HTTPClient = httpClient

	// Wrap with provider-specific middleware
	client := f.wrapWithMiddleware(retryClient.StandardClient(), options)

	return client, nil
}

// CreateProviderClient creates a client specifically configured for a git provider.
func (f *HTTPFactory) CreateProviderClient(provider string, token string) (*http.Client, error) {
	options := HTTPClientOptions{
		Provider:      provider,
		Authenticated: token != "",
		Token:         token,
	}

	// Provider-specific configurations
	switch provider {
	case ProviderTypeGitHub:
		options.Custom = map[string]any{
			"api_url":    "https://api.github.com",
			"accept":     "application/vnd.github.v3+json",
			"user_agent": "git-provider-sync/1.0",
		}
	case ProviderTypeGitLab:
		options.Custom = map[string]any{
			"api_url":    "https://gitlab.com/api/v4",
			"user_agent": "git-provider-sync/1.0",
		}
	case ProviderTypeGitea:
		options.Custom = map[string]any{
			"api_url":    "https://gitea.com/api/v1",
			"user_agent": "git-provider-sync/1.0",
		}
	default:
		return nil, fmt.Errorf("%w: %s", domain.ErrUnsupportedProvider, provider)
	}

	return f.CreateClient(options)
}

// CreateGitHTTPClient creates an HTTP client optimized for git operations.
func (f *HTTPFactory) CreateGitHTTPClient() (*http.Client, error) {
	// Git operations typically need longer timeouts
	httpClient := &http.Client{
		Timeout: f.defaultTimeout * 3, // 3x default timeout for git ops
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Always verify TLS certificates for security
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12, // Secure minimum TLS version
			},
			Proxy:              f.getProxy,
			DisableCompression: true, // Git doesn't need compression
		},
	}

	return httpClient, nil
}

// CreateWebhookClient creates an HTTP client for webhook operations.
func (f *HTTPFactory) CreateWebhookClient() (*http.Client, error) {
	// Webhook clients need shorter timeouts and fewer retries
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 2
	retryClient.RetryWaitMin = time.Second
	retryClient.RetryWaitMax = time.Second * 3
	retryClient.Logger = nil

	httpClient := &http.Client{
		Timeout: time.Second * 10, // Short timeout for webhooks
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Always verify TLS certificates for security
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12, // Secure minimum TLS version
			},
			Proxy: f.getProxy,
		},
	}

	retryClient.HTTPClient = httpClient

	return retryClient.StandardClient(), nil
}

// GetDefaultHTTPConfig returns a default HTTP configuration.
func GetDefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryDelay:    time.Second,
		SkipTLSVerify: false,
		Headers: map[string]string{
			"User-Agent": "git-provider-sync/1.0",
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: 10,
			BurstSize:         20,
			Enable:            true,
		},
	}
}

// GetProviderHTTPConfig returns provider-specific HTTP configuration.
func GetProviderHTTPConfig(provider string) HTTPConfig {
	config := GetDefaultHTTPConfig()

	switch provider {
	case "github":
		config.Headers["Accept"] = "application/vnd.github.v3+json"
		config.RateLimit.RequestsPerSecond = 5 // GitHub has strict rate limits
		config.Timeout = 45 * time.Second
	case "gitlab":
		config.RateLimit.RequestsPerSecond = 8
		config.Timeout = 60 * time.Second
	case "gitea":
		config.RateLimit.RequestsPerSecond = 15
		config.Timeout = 30 * time.Second
	}

	return config
}

// Private helper methods

// WrapWithMiddleware wraps the HTTP client with provider-specific middleware.
func (f *HTTPFactory) wrapWithMiddleware(client *http.Client, options HTTPClientOptions) *http.Client {
	// Create a new transport that wraps the existing one
	transport := &middlewareTransport{
		base:    client.Transport,
		factory: f,
		options: options,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   client.Timeout,
	}
}

// GetProxy returns the proxy URL if configured.
func (f *HTTPFactory) getProxy(_ /* req */ *http.Request) (*url.URL, error) {
	return f.proxyURL, nil
}

// ExponentialBackoff implements exponential backoff strategy.
func (f *HTTPFactory) exponentialBackoff(minDelay, maxDelay time.Duration, attemptNum int, _ /* resp */ *http.Response) time.Duration {
	// Prevent integer overflow by capping attempt number
	const maxAttempts = 10 // Reasonable cap to prevent overflow
	if attemptNum > maxAttempts {
		return maxDelay
	}

	// Safe conversion with bounds checking
	if attemptNum < 0 {
		attemptNum = 0
	}

	// Calculate multiplier with overflow protection
	multiplier := int64(1)
	for i := 0; i < attemptNum && multiplier < int64(maxDelay/minDelay); i++ {
		multiplier *= 2
	}

	sleep := time.Duration(multiplier) * minDelay
	if sleep > maxDelay {
		sleep = maxDelay
	}

	return sleep
}

// RetryPolicy determines when to retry requests.
func (f *HTTPFactory) retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// Don't retry on context cancellation
	if ctx != nil && ctx.Err() != nil {
		return false, fmt.Errorf("context cancelled during retry policy check: %w", ctx.Err())
	}

	// Always retry on connection errors
	if err != nil {
		return true, err
	}

	// Don't retry on successful responses
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return false, nil
	}

	// Retry on server errors
	if resp.StatusCode >= 500 {
		return true, nil
	}

	// Retry on rate limit errors
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}

	// Don't retry on client errors (except rate limits)
	return false, nil
}

// MiddlewareTransport wraps HTTP transport with additional functionality.
type middlewareTransport struct {
	base    http.RoundTripper
	factory *HTTPFactory
	options HTTPClientOptions
}

// RoundTrip implements the http.RoundTripper interface.
func (t *middlewareTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add custom headers
	for key, value := range t.factory.customHeaders {
		req.Header.Set(key, value)
	}

	// Add provider-specific headers
	t.addProviderHeaders(req)

	// Add authentication
	t.addAuthentication(req)

	// Perform the request
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("transport round trip failed: %w", err)
	}

	return resp, nil
}

// AddProviderHeaders adds provider-specific headers.
func (t *middlewareTransport) addProviderHeaders(req *http.Request) {
	if custom, ok := t.options.Custom["accept"].(string); ok {
		req.Header.Set("Accept", custom)
	}

	if userAgent, ok := t.options.Custom["user_agent"].(string); ok {
		req.Header.Set("User-Agent", userAgent)
	}
}

// AddAuthentication adds authentication to the request.
func (t *middlewareTransport) addAuthentication(req *http.Request) {
	if !t.options.Authenticated {
		return
	}

	switch t.options.Provider {
	case "github", "gitea":
		if t.options.Token != "" {
			req.Header.Set("Authorization", "token "+t.options.Token)
		}
	case "gitlab":
		if t.options.Token != "" {
			req.Header.Set("Authorization", "Bearer "+t.options.Token)
		}
	default:
		// Basic auth fallback
		if t.options.Username != "" && t.options.Password != "" {
			req.SetBasicAuth(t.options.Username, t.options.Password)
		} else if t.options.Token != "" {
			req.Header.Set("Authorization", "Bearer "+t.options.Token)
		}
	}
}
