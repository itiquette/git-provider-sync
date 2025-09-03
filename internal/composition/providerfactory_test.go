// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/transport"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Creates a test HTTP factory.
func createTestHTTPFactory() (*transport.HTTPFactory, error) {
	config := transport.GetDefaultHTTPConfig()
	config.SkipTLSVerify = false // Security: Always verify TLS

	factory, err := transport.NewHTTPFactory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create test HTTP factory: %w", err)
	}

	return factory, nil
}

// Mock logger for testing.
type mockLogger struct {
	mu       sync.Mutex
	messages []string
}

func (m *mockLogger) Debug(_ context.Context, msg string, _ map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, msg)
}

func (m *mockLogger) Info(_ context.Context, msg string, _ map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, msg)
}

func (m *mockLogger) Error(_ context.Context, msg string, _ map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, msg)
}

func (m *mockLogger) Warn(_ context.Context, _ string, _ map[string]any) {}

func (m *mockLogger) Fatal(_ context.Context, _ string, _ map[string]any) {}

func (m *mockLogger) Trace(_ context.Context, _ string, _ map[string]any) {}

func (m *mockLogger) IsLevelEnabled(_ ports.LogLevel) bool { return true }

func TestNewProviderFactory(t *testing.T) {
	t.Parallel()

	httpFactory, err := createTestHTTPFactory()
	require.NoError(t, err)

	logger := &mockLogger{}

	factory := NewProviderFactory(httpFactory, logger)

	require.NotNil(t, factory)
	assert.Equal(t, httpFactory, factory.httpFactory)
	assert.Equal(t, logger, factory.logger)
}

func TestProviderFactory_GetSupportedProviders(t *testing.T) {
	t.Parallel()

	httpFactory, err := createTestHTTPFactory()
	require.NoError(t, err)

	logger := &mockLogger{}

	factory := NewProviderFactory(httpFactory, logger)

	providers := factory.GetSupportedProviders()

	assert.Contains(t, providers, "github")
	assert.Contains(t, providers, "gitlab")
	assert.Contains(t, providers, "gitea")
	assert.Len(t, providers, 3)
}

func TestProviderFactory_ValidateProviderType(t *testing.T) {
	t.Parallel()

	httpFactory, err := createTestHTTPFactory()
	require.NoError(t, err)

	logger := &mockLogger{}

	factory := NewProviderFactory(httpFactory, logger)

	tests := []struct {
		name         string
		providerType string
		expectError  bool
	}{
		{"github valid", "github", false},
		{"GitHub case insensitive", "GitHub", false},
		{"gitlab valid", "gitlab", false},
		{"gitea valid", "gitea", false},
		{"unsupported", "unsupported", true},
		{"empty", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := factory.ValidateProviderType(test.providerType)

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProviderFactory_ValidationErrors(t *testing.T) {
	t.Parallel()

	httpFactory, err := createTestHTTPFactory()
	require.NoError(t, err)

	logger := &mockLogger{}

	factory := NewProviderFactory(httpFactory, logger)

	tests := []struct {
		name          string
		config        ProviderConfig
		expectedError string
	}{
		{
			name: "missing provider type",
			config: ProviderConfig{
				Domain: "github.com",
				Owner:  "testuser",
				Token:  "test-token",
			},
			expectedError: "provider type is required",
		},
		{
			name: "missing owner",
			config: ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com",
				Token:        "test-token",
			},
			expectedError: "owner is required",
		},
		{
			name: "missing authentication",
			config: ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com",
				Owner:        "testuser",
			},
			expectedError: "authentication is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Test validation directly without creating provider (avoids network calls)
			err := factory.validateConfig(test.config)

			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedError)
		})
	}
}

func TestProviderFactory_CreateProviderFromConfig(t *testing.T) {
	t.Parallel()

	httpFactory, err := createTestHTTPFactory()
	require.NoError(t, err)

	logger := &mockLogger{}

	factory := NewProviderFactory(httpFactory, logger)

	providerConfig := ports.ProviderConfig{
		ProviderType: "github",
		Domain:       "github.com",
		Owner:        "testuser",
		AuthConfig: ports.AuthenticationConfig{
			Token: "test-token",
		},
	}

	// Note: Only verifies the config conversion,
	// actual provider creation might require network calls
	clientConfig := ProviderConfig{
		ProviderType: providerConfig.ProviderType,
		Domain:       providerConfig.Domain,
		Owner:        providerConfig.Owner,
		Token:        providerConfig.AuthConfig.Token,
		Username:     providerConfig.AuthConfig.Username,
		SSHKeyPath:   providerConfig.AuthConfig.SSHKeyPath,
		SSHKey:       providerConfig.AuthConfig.SSHKey,
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		UserAgent:    "git-provider-sync/1.0",
	}

	// Test config validation instead of actual creation to avoid network calls
	err = factory.validateConfig(clientConfig)
	require.NoError(t, err)
}

func TestEnsureHTTPSScheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		domain   string
		expected string
	}{
		{
			name:     "default https",
			domain:   "example.com",
			expected: "https://example.com",
		},
		{
			name:     "already has scheme",
			domain:   "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "http scheme",
			domain:   "http://example.com",
			expected: "http://example.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ensureHTTPSScheme(test.domain)
			assert.Equal(t, test.expected, result)
		})
	}
}
