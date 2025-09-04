// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Mock logger for testing.
type mockLogger struct {
	mu            sync.Mutex
	debugMessages []string
	infoMessages  []string
}

func (m *mockLogger) Debug(_ context.Context, msg string, _ map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.debugMessages = append(m.debugMessages, msg)
}

func (m *mockLogger) Info(_ context.Context, msg string, _ map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.infoMessages = append(m.infoMessages, msg)
}

func (m *mockLogger) Error(_ context.Context, _ string, _ map[string]any) {}

func (m *mockLogger) Warn(_ context.Context, _ string, _ map[string]any) {}

func (m *mockLogger) Fatal(_ context.Context, _ string, _ map[string]any) {}

func (m *mockLogger) Trace(_ context.Context, _ string, _ map[string]any) {}

func (m *mockLogger) IsLevelEnabled(_ ports.LogLevel) bool { return true }

func TestHTTPClientFactory_Create_Success(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()

	client, err := factory.CreateHTTPClient(context.Background(), config)

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, time.Duration(config.RequestTimeout)*time.Second, client.Timeout)
	assert.NotNil(t, client.Transport)
	assert.Contains(t, logger.infoMessages, "HTTP client created successfully")
}

func TestHTTPClientFactory_Create_WithProxy(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()
	config.ProxyURL = "http://proxy.example.com:8080"

	client, err := factory.CreateHTTPClient(context.Background(), config)

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Contains(t, logger.infoMessages, "HTTP client created successfully")
}

func TestHTTPClientFactory_Create_InvalidProxy(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()
	config.ProxyURL = "http://[invalid-ipv6::"

	client, err := factory.CreateHTTPClient(context.Background(), config)

	require.Error(t, err)
	require.Nil(t, client)
	assert.Contains(t, err.Error(), "proxy setup error")
	assert.Contains(t, err.Error(), "invalid proxy configuration")
}

func TestHTTPClientFactory_Create_WithCertDirectory(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with a test certificate
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "test.crt")

	// Create a valid test certificate (generated with openssl for testing)
	certContent := `-----BEGIN CERTIFICATE-----
MIIDfzCCAmegAwIBAgIUCFzXvTnx/f2nqYAKALxEPD4gyD4wDQYJKoZIhvcNAQEL
BQAwTzELMAkGA1UEBhMCVVMxDTALBgNVBAgMBFRlc3QxDTALBgNVBAcMBFRlc3Qx
DTALBgNVBAoMBFRlc3QxEzARBgNVBAMMCnRlc3QubG9jYWwwHhcNMjUwNzI4MTIw
ODA4WhcNMjYwNzI4MTIwODA4WjBPMQswCQYDVQQGEwJVUzENMAsGA1UECAwEVGVz
dDENMAsGA1UEBwwEVGVzdDENMAsGA1UECgwEVGVzdDETMBEGA1UEAwwKdGVzdC5s
b2NhbDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMuz8/FBCNdAlPkL
8HboA4qkocvDp58OqNPb4OsemfI1hyb7PTIYuUR2U0T/fCNexui5irxVUTeKegW/
7mqIw3uA4hdBDvSBCMjcnWplZXJweMTABzy0cSWDXZ7pKHw5OaU11twBxIOxRJiR
uhCCqAs97Apbe7qMuM2M+GBFHZH76uF5/ZSPkoJysolU5iBW4bHpUlwUu6qnB3LL
EwVOTCbYIjG7bcjrIzKwaIDt61lgsoXEPBXU+j7FcG8hEEUkyOmkqdR/CZfRkxqQ
fcx3AJRVTDXLpMHPXrRgpmAK5PWaDztJAdrgnMtqo7cKrw5JO8ghtiAs8npD66u3
p1FWSEsCAwEAAaNTMFEwHQYDVR0OBBYEFKKgOevVn90Da+9HiKtZYZ0N5uDCMB8G
A1UdIwQYMBaAFKKgOevVn90Da+9HiKtZYZ0N5uDCMA8GA1UdEwEB/wQFMAMBAf8w
DQYJKoZIhvcNAQELBQADggEBALlpSAf9bMWvFN0RpS9Y4ymFcxt2GkAXfORuJZea
yEYEW5jg5HjUiJvqVSzRdVC9ZAl1+zMKcXj23kOB6NRXFMFWsz6ltzg/pSxyWnsU
9h3FfteIL2VMZVClZwt9+Z+ZFm7mUfrXDnPdNivu+ZIb/1ZijukmnUyQfGH23pUV
FZSIb8KGh5YKlXMMfllbYG1oQqynvSphGfHKfiLHKnTlm/8wlazMVbEXrbI6pSrs
NpAy5DpzMcbTq5Ms4gWmL4TdNekAqMuQr6utNtChnZK/1nxk9+6AZSmneJynRAcj
eZovO1wGwMVbfxCteUajepDTnclo0Y3iX+pN61aUwzsZ0HE=
-----END CERTIFICATE-----`

	err := os.WriteFile(certFile, []byte(certContent), 0600)
	require.NoError(t, err)

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()
	config.CertDirPath = tempDir

	client, err := factory.CreateHTTPClient(context.Background(), config)

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Contains(t, logger.infoMessages, "Custom certificates loaded successfully")
}

func TestHTTPClientFactory_Create_InvalidCertDirectory(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()
	config.CertDirPath = "/non/existent/directory"

	client, err := factory.CreateHTTPClient(context.Background(), config)

	require.Error(t, err)
	require.Nil(t, client)
	assert.Contains(t, err.Error(), "certificate loading error")
	assert.Contains(t, err.Error(), "failed to load certificates")
}

func TestHTTPClientFactory_CreateGitTransportClient(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()

	client, err := factory.CreateGitTransportClient(context.Background(), config)

	require.NoError(t, err)
	require.NotNil(t, client)
	// Git transport client should have a longer timeout
	assert.Equal(t, 300*time.Second, client.Timeout)
	assert.Contains(t, logger.infoMessages, "Git transport client created with protocol registration")
}

func TestHTTPClientFactory_CreateAPIClient(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()

	client, err := factory.CreateAPIClient(context.Background(), config)

	require.NoError(t, err)
	require.NotNil(t, client)
	// API client should have a 30-second timeout
	assert.Equal(t, 30*time.Second, client.Timeout)
}

func TestHTTPClientFactory_ValidateConfig_Success(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()

	err := factory.ValidateConfig(config)
	require.NoError(t, err)
}

func TestHTTPClientFactory_ValidateConfig_InvalidTimeouts(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)

	tests := []struct {
		name         string
		modifyConfig func(*HTTPClientConfig)
		expectedErr  error
	}{
		{
			name: "negative request timeout",
			modifyConfig: func(config *HTTPClientConfig) {
				config.RequestTimeout = -1
			},
			expectedErr: domain.ErrRequestTimeoutMustBePositive,
		},
		{
			name: "zero idle connection timeout",
			modifyConfig: func(config *HTTPClientConfig) {
				config.IdleConnTimeout = 0
			},
			expectedErr: domain.ErrIdleConnectionTimeoutMustBePositive,
		},
		{
			name: "negative dial timeout",
			modifyConfig: func(config *HTTPClientConfig) {
				config.DialTimeout = -1 * time.Second
			},
			expectedErr: domain.ErrDialTimeoutMustBePositive,
		},
		{
			name: "zero keep alive",
			modifyConfig: func(config *HTTPClientConfig) {
				config.KeepAlive = 0
			},
			expectedErr: domain.ErrKeepAliveMustBePositive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultHTTPClientConfig()
			test.modifyConfig(&config)

			err := factory.ValidateConfig(config)
			require.Error(t, err)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestHTTPClientFactory_ValidateConfig_InvalidConnectionLimits(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)

	tests := []struct {
		name         string
		modifyConfig func(*HTTPClientConfig)
		expectedErr  error
	}{
		{
			name: "zero max idle connections",
			modifyConfig: func(config *HTTPClientConfig) {
				config.MaxIdleConns = 0
			},
			expectedErr: domain.ErrMaxIdleConnectionsMustBePositive,
		},
		{
			name: "negative max idle connections per host",
			modifyConfig: func(config *HTTPClientConfig) {
				config.MaxIdleConnsPerHost = -1
			},
			expectedErr: domain.ErrMaxIdleConnectionsPerHostMustBePositive,
		},
		{
			name: "zero max connections per host",
			modifyConfig: func(config *HTTPClientConfig) {
				config.MaxConnsPerHost = 0
			},
			expectedErr: domain.ErrMaxConnectionsPerHostMustBePositive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultHTTPClientConfig()
			test.modifyConfig(&config)

			err := factory.ValidateConfig(config)
			require.Error(t, err)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestHTTPClientFactory_ValidateConfig_InvalidPaths(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)

	tests := []struct {
		name          string
		modifyConfig  func(*HTTPClientConfig)
		expectError   bool
		errorContains string
	}{
		{
			name: "non-existent cert directory",
			modifyConfig: func(config *HTTPClientConfig) {
				config.CertDirPath = "/non/existent/directory"
			},
			expectError:   true,
			errorContains: "certificate directory does not exist",
		},
		{
			name: "invalid proxy URL",
			modifyConfig: func(config *HTTPClientConfig) {
				config.ProxyURL = "http://[invalid-ipv6::"
			},
			expectError:   true,
			errorContains: "invalid proxy URL",
		},
		{
			name: "valid empty paths",
			modifyConfig: func(config *HTTPClientConfig) {
				config.CertDirPath = ""
				config.ProxyURL = ""
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultHTTPClientConfig()
			test.modifyConfig(&config)

			err := factory.ValidateConfig(config)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.errorContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHTTPClientFactory_setupProxy(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)

	tests := []struct {
		name        string
		proxyURL    string
		expectError bool
	}{
		{
			name:        "empty proxy URL",
			proxyURL:    "",
			expectError: false,
		},
		{
			name:        "valid proxy URL",
			proxyURL:    "http://proxy.example.com:8080",
			expectError: false,
		},
		{
			name:        "invalid proxy URL",
			proxyURL:    "http://[invalid-ipv6::",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			proxyFunc, err := factory.setupProxy(context.Background(), test.proxyURL)

			if test.expectError {
				require.Error(t, err)
				require.Nil(t, proxyFunc)
				require.ErrorIs(t, err, ErrInvalidProxy)
			} else {
				require.NoError(t, err)
				require.NotNil(t, proxyFunc)
			}
		})
	}
}

func TestHTTPClientFactory_loadCertificates_SystemPool(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)

	// Test with empty directory path - should use system pool
	certPool, err := factory.loadCertificates(context.Background(), "")

	require.NoError(t, err)
	require.NotNil(t, certPool)
	assert.Contains(t, logger.debugMessages, "No certificate directory provided, using system certificates")
}

func TestHTTPClientFactory_loadCertificates_CustomDirectory(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with certificate files
	tempDir := t.TempDir()

	// Valid certificate file
	validCertFile := filepath.Join(tempDir, "valid.crt")
	validCertContent := `-----BEGIN CERTIFICATE-----
MIIDfzCCAmegAwIBAgIUCFzXvTnx/f2nqYAKALxEPD4gyD4wDQYJKoZIhvcNAQEL
BQAwTzELMAkGA1UEBhMCVVMxDTALBgNVBAgMBFRlc3QxDTALBgNVBAcMBFRlc3Qx
DTALBgNVBAoMBFRlc3QxEzARBgNVBAMMCnRlc3QubG9jYWwwHhcNMjUwNzI4MTIw
ODA4WhcNMjYwNzI4MTIwODA4WjBPMQswCQYDVQQGEwJVUzENMAsGA1UECAwEVGVz
dDENMAsGA1UEBwwEVGVzdDENMAsGA1UECgwEVGVzdDETMBEGA1UEAwwKdGVzdC5s
b2NhbDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMuz8/FBCNdAlPkL
8HboA4qkocvDp58OqNPb4OsemfI1hyb7PTIYuUR2U0T/fCNexui5irxVUTeKegW/
7mqIw3uA4hdBDvSBCMjcnWplZXJweMTABzy0cSWDXZ7pKHw5OaU11twBxIOxRJiR
uhCCqAs97Apbe7qMuM2M+GBFHZH76uF5/ZSPkoJysolU5iBW4bHpUlwUu6qnB3LL
EwVOTCbYIjG7bcjrIzKwaIDt61lgsoXEPBXU+j7FcG8hEEUkyOmkqdR/CZfRkxqQ
fcx3AJRVTDXLpMHPXrRgpmAK5PWaDztJAdrgnMtqo7cKrw5JO8ghtiAs8npD66u3
p1FWSEsCAwEAAaNTMFEwHQYDVR0OBBYEFKKgOevVn90Da+9HiKtZYZ0N5uDCMB8G
A1UdIwQYMBaAFKKgOevVn90Da+9HiKtZYZ0N5uDCMA8GA1UdEwEB/wQFMAMBAf8w
DQYJKoZIhvcNAQELBQADggEBALlpSAf9bMWvFN0RpS9Y4ymFcxt2GkAXfORuJZea
yEYEW5jg5HjUiJvqVSzRdVC9ZAl1+zMKcXj23kOB6NRXFMFWsz6ltzg/pSxyWnsU
9h3FfteIL2VMZVClZwt9+Z+ZFm7mUfrXDnPdNivu+ZIb/1ZijukmnUyQfGH23pUV
FZSIb8KGh5YKlXMMfllbYG1oQqynvSphGfHKfiLHKnTlm/8wlazMVbEXrbI6pSrs
NpAy5DpzMcbTq5Ms4gWmL4TdNekAqMuQr6utNtChnZK/1nxk9+6AZSmneJynRAcj
eZovO1wGwMVbfxCteUajepDTnclo0Y3iX+pN61aUwzsZ0HE=
-----END CERTIFICATE-----`

	err := os.WriteFile(validCertFile, []byte(validCertContent), 0600)
	require.NoError(t, err)

	// Non-certificate file (should be ignored)
	nonCertFile := filepath.Join(tempDir, "readme.txt")
	err = os.WriteFile(nonCertFile, []byte("This is not a certificate"), 0600)
	require.NoError(t, err)

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)

	certPool, err := factory.loadCertificates(context.Background(), tempDir)

	require.NoError(t, err)
	require.NotNil(t, certPool)
	assert.Contains(t, logger.infoMessages, "Custom certificates loaded successfully")
}

func TestHTTPClientFactory_loadCertificates_InvalidCertificate(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with invalid certificate
	tempDir := t.TempDir()

	invalidCertFile := filepath.Join(tempDir, "invalid.crt")
	err := os.WriteFile(invalidCertFile, []byte("invalid certificate content"), 0600)
	require.NoError(t, err)

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)

	certPool, err := factory.loadCertificates(context.Background(), tempDir)

	require.Error(t, err)
	require.Nil(t, certPool)
	require.ErrorIs(t, err, ErrCertificateLoad)
	assert.Contains(t, err.Error(), "failed to parse certificate")
}

func TestHTTPClientFactory_createTLSConfig(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()
	config.InsecureSkipVerify = true

	tlsConfig := factory.createTLSConfig(context.Background(), nil, config)

	require.NotNil(t, tlsConfig)
	assert.Equal(t, config.MinTLSVersion, tlsConfig.MinVersion)
	assert.Equal(t, config.MaxTLSVersion, tlsConfig.MaxVersion)
	assert.True(t, tlsConfig.InsecureSkipVerify)
	assert.Equal(t, tls.NoClientCert, tlsConfig.ClientAuth)
	assert.Equal(t, tls.RenegotiateNever, tlsConfig.Renegotiation)
	assert.False(t, tlsConfig.SessionTicketsDisabled)
}

func TestHTTPClientFactory_createTLSConfig_WithCustomCerts(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()

	// Create a custom cert pool
	certPool := x509.NewCertPool()

	tlsConfig := factory.createTLSConfig(context.Background(), certPool, config)

	require.NotNil(t, tlsConfig)
	assert.Equal(t, certPool, tlsConfig.RootCAs)
}

func TestHTTPClientFactory_createHTTPTransport(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	factory := NewHTTPClientFactory(logger)
	config := DefaultHTTPClientConfig()

	// Mock proxy function
	proxyFunc := func(*http.Request) (*url.URL, error) {
		return nil, nil //nolint:nilnil // Valid for mock proxy function
	}

	// Mock TLS config
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	transport := factory.createHTTPTransport(context.Background(), proxyFunc, tlsConfig, config)

	require.NotNil(t, transport)
	assert.Equal(t, config.MaxIdleConns, transport.MaxIdleConns)
	assert.Equal(t, config.MaxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
	assert.Equal(t, config.MaxConnsPerHost, transport.MaxConnsPerHost)
	assert.Equal(t, config.IdleConnTimeout, transport.IdleConnTimeout)
	assert.Equal(t, tlsConfig, transport.TLSClientConfig)
	assert.True(t, transport.ForceAttemptHTTP2)
	assert.Equal(t, config.WriteBufferSize, transport.WriteBufferSize)
	assert.Equal(t, config.ReadBufferSize, transport.ReadBufferSize)
	assert.Equal(t, 10*time.Second, transport.TLSHandshakeTimeout)
	assert.Equal(t, 1*time.Second, transport.ExpectContinueTimeout)
}
