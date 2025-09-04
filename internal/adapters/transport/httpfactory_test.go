// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package transport

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
)

func TestNewHTTPFactory_InvalidProxyURL(t *testing.T) {
	t.Parallel()

	config := HTTPConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
		ProxyURL:   "http://[invalid-ipv6::",
	}

	factory, err := NewHTTPFactory(config)

	require.Error(t, err)
	assert.Nil(t, factory) // Should be nil for invalid proxy URL
	assert.Contains(t, err.Error(), "invalid proxy URL")
}

func TestHTTPFactory_CreateClient(t *testing.T) {
	t.Parallel()

	config := HTTPConfig{
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryDelay:    time.Second,
		SkipTLSVerify: false,
		Headers: map[string]string{
			"User-Agent": "test-agent",
		},
	}

	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	options := HTTPClientOptions{
		Provider:      "github",
		Authenticated: true,
		Token:         "test-token",
	}

	client, err := factory.CreateClient(options)

	require.NoError(t, err)
	require.NotNil(t, client)
	// CreateClient uses retryablehttp which wraps the client, so timeout may be different
	assert.NotNil(t, client.Transport)
}

func TestHTTPFactory_CreateProviderClient_GitHub(t *testing.T) {
	t.Parallel()

	config := GetDefaultHTTPConfig()
	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	client, err := factory.CreateProviderClient(ProviderTypeGitHub, "test-token")

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestHTTPFactory_CreateProviderClient_GitLab(t *testing.T) {
	t.Parallel()

	config := GetDefaultHTTPConfig()
	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	client, err := factory.CreateProviderClient(ProviderTypeGitLab, "test-token")

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestHTTPFactory_CreateProviderClient_Gitea(t *testing.T) {
	t.Parallel()

	config := GetDefaultHTTPConfig()
	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	client, err := factory.CreateProviderClient(ProviderTypeGitea, "test-token")

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestHTTPFactory_CreateProviderClient_UnsupportedProvider(t *testing.T) {
	t.Parallel()

	config := GetDefaultHTTPConfig()
	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	client, err := factory.CreateProviderClient("unsupported", "test-token")

	require.Error(t, err)
	require.Nil(t, client)
	require.ErrorIs(t, err, domain.ErrUnsupportedProvider)
}

func TestHTTPFactory_CreateGitHTTPClient(t *testing.T) {
	t.Parallel()

	config := HTTPConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	client, err := factory.CreateGitHTTPClient()

	require.NoError(t, err)
	require.NotNil(t, client)
	// Git client should have 3x timeout
	assert.Equal(t, config.Timeout*3, client.Timeout)

	// Check transport configuration
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.True(t, transport.DisableCompression)
}

func TestHTTPFactory_CreateWebhookClient(t *testing.T) {
	t.Parallel()

	config := GetDefaultHTTPConfig()
	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	client, err := factory.CreateWebhookClient()

	require.NoError(t, err)
	require.NotNil(t, client)
	// CreateWebhookClient also uses retryablehttp, so timeout behavior may be different
	assert.NotNil(t, client.Transport)
}

func TestGetProviderHTTPConfig_GitHub(t *testing.T) {
	t.Parallel()

	config := GetProviderHTTPConfig("github")

	assert.Equal(t, "application/vnd.github.v3+json", config.Headers["Accept"])
	assert.InDelta(t, float64(5), config.RateLimit.RequestsPerSecond, 0.01)
	assert.Equal(t, 45*time.Second, config.Timeout)
}

func TestGetProviderHTTPConfig_GitLab(t *testing.T) {
	t.Parallel()

	config := GetProviderHTTPConfig("gitlab")

	assert.InDelta(t, float64(8), config.RateLimit.RequestsPerSecond, 0.01)
	assert.Equal(t, 60*time.Second, config.Timeout)
}

func TestGetProviderHTTPConfig_Gitea(t *testing.T) {
	t.Parallel()

	config := GetProviderHTTPConfig("gitea")

	assert.InDelta(t, float64(15), config.RateLimit.RequestsPerSecond, 0.01)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

func TestGetProviderHTTPConfig_Unknown(t *testing.T) {
	t.Parallel()

	config := GetProviderHTTPConfig("unknown")

	// Should return default config
	defaultConfig := GetDefaultHTTPConfig()
	assert.Equal(t, defaultConfig.Timeout, config.Timeout)
	assert.InDelta(t, defaultConfig.RateLimit.RequestsPerSecond, config.RateLimit.RequestsPerSecond, 0.01)
}

func TestHTTPFactory_getProxy_WithProxyURL(t *testing.T) {
	t.Parallel()

	config := HTTPConfig{
		ProxyURL: "http://proxy.example.com:8080",
	}

	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	proxyURL, err := factory.getProxy(req)

	require.NoError(t, err)
	require.NotNil(t, proxyURL)
	assert.Equal(t, "proxy.example.com:8080", proxyURL.Host)
}

func TestHTTPFactory_getProxy_NoProxyURL(t *testing.T) {
	t.Parallel()

	config := HTTPConfig{}
	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	proxyURL, err := factory.getProxy(req)

	require.NoError(t, err)
	assert.Nil(t, proxyURL)
}

func TestHTTPFactory_exponentialBackoff(t *testing.T) {
	t.Parallel()

	factory := &HTTPFactory{}
	minDelay := time.Second
	maxDelay := 30 * time.Second

	tests := []struct {
		name       string
		attemptNum int
		expected   time.Duration
	}{
		{"first attempt", 0, time.Second},
		{"second attempt", 1, 2 * time.Second},
		{"third attempt", 2, 4 * time.Second},
		{"fourth attempt", 3, 8 * time.Second},
		{"large attempt", 20, maxDelay},       // Should cap at maxDelay
		{"negative attempt", -1, time.Second}, // Should default to 0
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			delay := factory.exponentialBackoff(minDelay, maxDelay, test.attemptNum, nil)
			assert.Equal(t, test.expected, delay)
		})
	}
}

func TestHTTPFactory_retryPolicy(t *testing.T) {
	t.Parallel()

	factory := &HTTPFactory{}
	ctx := context.Background()

	tests := []struct {
		name           string
		response       *http.Response
		err            error
		expectedRetry  bool
		expectedErrNil bool
	}{
		{
			name:           "connection error",
			response:       nil,
			err:            assert.AnError,
			expectedRetry:  true,
			expectedErrNil: false,
		},
		{
			name: "success response",
			response: &http.Response{
				StatusCode: http.StatusOK,
			},
			err:            nil,
			expectedRetry:  false,
			expectedErrNil: true,
		},
		{
			name: "server error",
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			err:            nil,
			expectedRetry:  true,
			expectedErrNil: true,
		},
		{
			name: "rate limit error",
			response: &http.Response{
				StatusCode: http.StatusTooManyRequests,
			},
			err:            nil,
			expectedRetry:  true,
			expectedErrNil: true,
		},
		{
			name: "client error",
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
			},
			err:            nil,
			expectedRetry:  false,
			expectedErrNil: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			retry, err := factory.retryPolicy(ctx, test.response, test.err)

			assert.Equal(t, test.expectedRetry, retry)

			if test.expectedErrNil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestHTTPFactory_retryPolicy_ContextCancelled(t *testing.T) {
	t.Parallel()

	factory := &HTTPFactory{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context

	retry, err := factory.retryPolicy(ctx, nil, nil)

	assert.False(t, retry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestMiddlewareTransport_RoundTrip(t *testing.T) {
	t.Parallel()

	// Create a mock base transport
	baseTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       http.NoBody,
		},
	}

	factory := &HTTPFactory{
		customHeaders: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}

	options := HTTPClientOptions{
		Provider:      "github",
		Authenticated: true,
		Token:         "test-token",
		Custom: map[string]any{
			"accept":     "application/json",
			"user_agent": "test-agent",
		},
	}

	transport := &middlewareTransport{
		base:    baseTransport,
		factory: factory,
		options: options,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check that custom headers were added
	assert.Equal(t, "custom-value", req.Header.Get("X-Custom-Header"))
	// Check that provider headers were added
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
	// Check that authentication was added
	assert.Equal(t, "token test-token", req.Header.Get("Authorization"))
}

func TestMiddlewareTransport_addProviderHeaders(t *testing.T) {
	t.Parallel()

	options := HTTPClientOptions{
		Custom: map[string]any{
			"accept":     "application/json",
			"user_agent": "test-agent",
		},
	}

	transport := &middlewareTransport{
		options: options,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	transport.addProviderHeaders(req)

	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
}

func TestMiddlewareTransport_addAuthentication_GitHub(t *testing.T) {
	t.Parallel()

	options := HTTPClientOptions{
		Provider:      "github",
		Authenticated: true,
		Token:         "github-token",
	}

	transport := &middlewareTransport{
		options: options,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	transport.addAuthentication(req)

	assert.Equal(t, "token github-token", req.Header.Get("Authorization"))
}

func TestMiddlewareTransport_addAuthentication_GitLab(t *testing.T) {
	t.Parallel()

	options := HTTPClientOptions{
		Provider:      "gitlab",
		Authenticated: true,
		Token:         "gitlab-token",
	}

	transport := &middlewareTransport{
		options: options,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	transport.addAuthentication(req)

	assert.Equal(t, "Bearer gitlab-token", req.Header.Get("Authorization"))
}

func TestMiddlewareTransport_addAuthentication_BasicAuth(t *testing.T) {
	t.Parallel()

	options := HTTPClientOptions{
		Provider:      "unknown",
		Authenticated: true,
		Username:      "testuser",
		Password:      "testpass",
	}

	transport := &middlewareTransport{
		options: options,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	transport.addAuthentication(req)

	username, password, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "testpass", password)
}

func TestMiddlewareTransport_addAuthentication_NotAuthenticated(t *testing.T) {
	t.Parallel()

	options := HTTPClientOptions{
		Provider:      "github",
		Authenticated: false,
	}

	transport := &middlewareTransport{
		options: options,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	transport.addAuthentication(req)

	assert.Empty(t, req.Header.Get("Authorization"))
}

// Mock transport for testing.
type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.response, m.err
}

// Test TLS bypass security enhancement

func TestNewHTTPFactory_TLSBypassNotAllowed(t *testing.T) {
	t.Parallel()

	config := HTTPConfig{
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryDelay:    time.Second,
		SkipTLSVerify: true, // Attempt to bypass TLS verification
	}

	factory, err := NewHTTPFactory(config)

	require.Error(t, err)
	assert.Nil(t, factory)
	assert.Contains(t, err.Error(), "TLS verification bypass is not permitted for security reasons")
}

func TestNewHTTPFactory_TLSVerificationAlwaysEnabled(t *testing.T) {
	t.Parallel()

	config := HTTPConfig{
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryDelay:    time.Second,
		SkipTLSVerify: false, // Normal secure TLS verification
	}

	factory, err := NewHTTPFactory(config)
	require.NoError(t, err)

	require.NotNil(t, factory)
	assert.False(t, factory.skipTLSVerify) // Should always be false for secure operation

	// Verify that all created clients also have TLS verification enabled
	client, err := factory.CreateGitHTTPClient()
	require.NoError(t, err)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
}
