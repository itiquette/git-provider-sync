// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package auth

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Test GetAuthMethod - core security function

func TestService_GetAuthMethod_SSH(t *testing.T) {
	t.Parallel()

	// Mock SSH agent by setting SSH_AUTH_SOCK environment variable
	originalSSHAuthSock := os.Getenv("SSH_AUTH_SOCK")

	defer func() {
		if originalSSHAuthSock != "" {
			_ = os.Setenv("SSH_AUTH_SOCK", originalSSHAuthSock)
		} else {
			_ = os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	// Create a mock socket file in temp directory for test isolation
	tempDir := t.TempDir()
	mockSocketPath := tempDir + "/ssh-agent.sock"

	// Set the mock SSH_AUTH_SOCK for this test
	_ = os.Setenv("SSH_AUTH_SOCK", mockSocketPath)

	service := NewService()
	ctx := context.Background()

	authConfig := ports.AuthenticationConfiguration{
		Type:       ports.AuthenticationTypeSSH,
		Protocol:   protocolSSH,
		SSHKeyPath: "/path/to/key",
	}

	// The SSH agent creation may fail due to mock socket, but we test the method logic
	authMethod, err := service.GetAuthMethod(ctx, authConfig)

	// We expect this to work since we're testing the method selection logic
	// The actual SSH agent connection failure is acceptable in tests
	if err != nil {
		// SSH agent creation failed as expected with mock - verify error message
		require.Contains(t, err.Error(), "error creating SSH agent")

		return
	}

	require.NotNil(t, authMethod)

	// Verify it's SSH agent auth if creation succeeded
	sshAuth, ok := authMethod.(*ssh.PublicKeysCallback)
	require.True(t, ok, "Expected SSH agent authentication")
	assert.NotNil(t, sshAuth)
}

func TestService_GetAuthMethod_HTTPS(t *testing.T) {
	t.Parallel()

	service := NewService()
	ctx := context.Background()

	tests := []struct {
		name     string
		protocol string
		token    string
	}{
		{
			name:     "https protocol",
			protocol: "https",
			token:    "github_pat_test123",
		},
		{
			name:     "tls protocol",
			protocol: "tls",
			token:    "gitlab_token_456",
		},
		{
			name:     "empty protocol (defaults to https)",
			protocol: "",
			token:    "token_789",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			authConfig := ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeToken,
				Protocol: test.protocol,
				Token:    test.token,
			}

			authMethod, err := service.GetAuthMethod(ctx, authConfig)

			require.NoError(t, err)
			require.NotNil(t, authMethod)

			// Verify it's basic auth with correct credentials
			basicAuth, ok := authMethod.(*http.BasicAuth)
			require.True(t, ok, "Expected HTTP basic authentication")
			assert.Equal(t, "token", basicAuth.Username)
			assert.Equal(t, test.token, basicAuth.Password)
		})
	}
}

func TestService_GetAuthMethod_InvalidProtocol(t *testing.T) {
	t.Parallel()

	service := NewService()
	ctx := context.Background()

	authConfig := ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeToken,
		Protocol: "invalid-protocol",
		Token:    "token123",
	}

	authMethod, err := service.GetAuthMethod(ctx, authConfig)

	require.Error(t, err)
	assert.Nil(t, authMethod)
	require.ErrorIs(t, err, ErrInvalidAuthProtocol)
	assert.Contains(t, err.Error(), "invalid-protocol")
}

func TestService_GetAuthMethod_CaseInsensitive(t *testing.T) {
	t.Parallel()

	service := NewService()
	ctx := context.Background()

	tests := []struct {
		name     string
		protocol string
	}{
		{"uppercase SSH", "SSH"},
		{"mixed case HTTPS", "HttpS"},
		{"uppercase TLS", "TLS"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock SSH environment for SSH tests
			if strings.ToLower(test.protocol) == protocolSSH {
				originalSSHAuthSock := os.Getenv("SSH_AUTH_SOCK")

				defer func() {
					if originalSSHAuthSock != "" {
						_ = os.Setenv("SSH_AUTH_SOCK", originalSSHAuthSock)
					} else {
						_ = os.Unsetenv("SSH_AUTH_SOCK")
					}
				}()

				// Set mock SSH_AUTH_SOCK for test isolation
				mockSocketPath := t.TempDir() + "/ssh-agent.sock"
				_ = os.Setenv("SSH_AUTH_SOCK", mockSocketPath)
			}

			authConfig := ports.AuthenticationConfiguration{
				Protocol: test.protocol,
				Token:    "test-token",
			}

			// For SSH tests without SSH_AUTH_SOCK, provide an SSH key path
			if strings.ToLower(test.protocol) == protocolSSH {
				authConfig.SSHKeyPath = "/path/to/key" // Mock SSH key path for testing
			}

			authMethod, err := service.GetAuthMethod(ctx, authConfig)

			// Handle SSH agent errors in mocked environment
			if strings.ToLower(test.protocol) == protocolSSH && err != nil {
				// SSH agent creation failed as expected with mock - verify error message
				require.Contains(t, err.Error(), "error creating SSH agent")

				return
			}

			require.NoError(t, err)
			require.NotNil(t, authMethod)
		})
	}
}

// Test AuthenticationAdapter

func TestNewAuthenticationAdapter(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	require.NotNil(t, adapter)
	require.NotNil(t, adapter.service)
}

func TestAuthenticationAdapter_GetTransportAuth(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()
	ctx := context.Background()

	authConfig := ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeToken,
		Protocol: "https",
		Token:    "test-token",
	}

	authMethod, err := adapter.GetTransportAuth(ctx, authConfig)

	require.NoError(t, err)
	require.NotNil(t, authMethod)

	basicAuth, ok := authMethod.(*http.BasicAuth)
	require.True(t, ok)
	assert.Equal(t, "token", basicAuth.Username)
	assert.Equal(t, "test-token", basicAuth.Password)
}

// Test ValidateAuthConfig - critical security validation

func TestAuthenticationAdapter_ValidateAuthConfig_Valid(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	tests := []struct {
		name       string
		authConfig ports.AuthenticationConfiguration
	}{
		{
			name: "valid HTTPS token auth",
			authConfig: ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeToken,
				Protocol: "https",
				Token:    "github_pat_123",
			},
		},
		{
			name: "valid TLS token auth",
			authConfig: ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeToken,
				Protocol: "tls",
				Token:    "gitlab_token_456",
			},
		},
		{
			name: "valid SSH key path",
			authConfig: ports.AuthenticationConfiguration{
				Type:       ports.AuthenticationTypeSSH,
				Protocol:   protocolSSH,
				SSHKeyPath: filepath.Join(t.TempDir(), ".ssh", "id_rsa"),
			},
		},
		{
			name: "valid SSH key content",
			authConfig: ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeSSH,
				Protocol: protocolSSH,
				SSHKey:   "-----BEGIN OPENSSH PRIVATE KEY-----",
			},
		},
		{
			name: "empty protocol defaults to HTTPS",
			authConfig: ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeToken,
				Protocol: "",
				Token:    "token123",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := adapter.ValidateAuthConfig(test.authConfig)
			require.NoError(t, err)
		})
	}
}

func TestAuthenticationAdapter_ValidateAuthConfig_MissingCredentials(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	authConfig := ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeToken,
		Protocol: "https",
		// Missing token, SSH key path, and SSH key content
	}

	err := adapter.ValidateAuthConfig(authConfig)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrAuthConfigRequiresCredentials)
}

func TestAuthenticationAdapter_ValidateAuthConfig_SSH_MissingKey(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	authConfig := ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeSSH,
		Protocol: protocolSSH,
		// Missing SSH key path and SSH key content (and no token either)
	}

	err := adapter.ValidateAuthConfig(authConfig)

	require.Error(t, err)
	// Will fail the general credentials check first
	require.ErrorIs(t, err, ErrAuthConfigRequiresCredentials)
}

func TestAuthenticationAdapter_ValidateAuthConfig_HTTPS_MissingToken(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	tests := []struct {
		name     string
		protocol string
	}{
		{"https protocol", "https"},
		{"tls protocol", "tls"},
		{"empty protocol", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			authConfig := ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeToken,
				Protocol: test.protocol,
				// Missing token (and no SSH keys either)
			}

			err := adapter.ValidateAuthConfig(authConfig)

			require.Error(t, err)
			// Will fail the general credentials check first
			require.ErrorIs(t, err, ErrAuthConfigRequiresCredentials)
		})
	}
}

func TestAuthAdapter_ValidateSSHAuth_MissingKey_ReturnsError(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	authConfig := ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeSSH,
		Protocol: protocolSSH,
		Token:    "some-token", // Has token but SSH protocol requires SSH key
		// Missing SSH key path and SSH key content
	}

	err := adapter.ValidateAuthConfig(authConfig)

	require.Error(t, err)
	// Should fail protocol-specific validation since we have general credentials
	require.ErrorIs(t, err, ErrSSHRequiresKey)
}

func TestAuthAdapter_ValidateHTTPSAuth_MissingToken_ReturnsError(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	tests := []struct {
		name     string
		protocol string
	}{
		{"https protocol", "https"},
		{"tls protocol", "tls"},
		{"empty protocol", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			authConfig := ports.AuthenticationConfiguration{
				Type:       ports.AuthenticationTypeToken,
				Protocol:   test.protocol,
				SSHKeyPath: "/path/to/key", // Has SSH key but HTTPS protocol requires token
				// Missing token
			}

			err := adapter.ValidateAuthConfig(authConfig)

			require.Error(t, err)
			// Should fail protocol-specific validation since we have general credentials
			require.ErrorIs(t, err, ErrHTTPSRequiresToken)
		})
	}
}

func TestAuthAdapter_ValidateAuth_UnsupportedProtocol_ReturnsError(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	authConfig := ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeToken,
		Protocol: "ftp", // Unsupported protocol
		Token:    "token123",
	}

	err := adapter.ValidateAuthConfig(authConfig)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedProtocol)
	assert.Contains(t, err.Error(), "ftp")
}

func TestAuthenticationAdapter_ValidateAuthConfig_CaseInsensitive(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	tests := []struct {
		name     string
		protocol string
		token    string
	}{
		{"uppercase SSH", "SSH", ""},
		{"mixed case HTTPS", "HttpS", "token123"},
		{"uppercase TLS", "TLS", "token456"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			authConfig := ports.AuthenticationConfiguration{
				Protocol: test.protocol,
				Token:    test.token,
			}

			if test.protocol == "SSH" {
				authConfig.SSHKeyPath = "/path/to/key"
			}

			err := adapter.ValidateAuthConfig(authConfig)
			require.NoError(t, err)
		})
	}
}

// Test SupportsSSHAgent

func TestAuthenticationAdapter_SupportsSSHAgent(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	tests := []struct {
		name            string
		sshAuthSock     string
		setupSocket     bool
		expectedSupport bool
		description     string
	}{
		{
			name:            "no SSH_AUTH_SOCK environment variable",
			sshAuthSock:     "",
			setupSocket:     false,
			expectedSupport: false,
			description:     "should return false when SSH_AUTH_SOCK is not set",
		},
		{
			name:            "SSH_AUTH_SOCK set but socket doesn't exist",
			sshAuthSock:     "/tmp/non-existent-socket",
			setupSocket:     false,
			expectedSupport: false,
			description:     "should return false when socket doesn't exist",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// Set up environment
			originalSSHAuthSock := os.Getenv("SSH_AUTH_SOCK")

			defer func() {
				// Restore original environment
				if originalSSHAuthSock != "" {
					_ = os.Setenv("SSH_AUTH_SOCK", originalSSHAuthSock)
				} else {
					_ = os.Unsetenv("SSH_AUTH_SOCK")
				}
			}()

			if test.sshAuthSock != "" {
				_ = os.Setenv("SSH_AUTH_SOCK", test.sshAuthSock)
			} else {
				_ = os.Unsetenv("SSH_AUTH_SOCK")
			}

			result := adapter.SupportsSSHAgent()
			assert.Equal(t, test.expectedSupport, result, test.description)
		})
	}
}

// Test GetProtocolForURL - security-critical URL parsing

func TestAuthenticationAdapter_GetProtocolForURL(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	tests := []struct {
		name             string
		url              string
		expectedProtocol string
	}{
		{
			name:             "SSH git@ URL",
			url:              "git@github.com:owner/repo.git",
			expectedProtocol: protocolSSH,
		},
		{
			name:             "SSH protocol URL",
			url:              "ssh://git@gitlab.com/owner/repo.git",
			expectedProtocol: protocolSSH,
		},
		{
			name:             "HTTPS URL",
			url:              "https://github.com/owner/repo.git",
			expectedProtocol: "https",
		},
		{
			name:             "HTTP URL",
			url:              "http://example.com/repo.git",
			expectedProtocol: "https",
		},
		{
			name:             "unknown format defaults to HTTPS",
			url:              "file:///path/to/repo.git",
			expectedProtocol: "https",
		},
		{
			name:             "empty URL defaults to HTTPS",
			url:              "",
			expectedProtocol: "https",
		},
		{
			name:             "malformed URL defaults to HTTPS",
			url:              "not-a-url",
			expectedProtocol: "https",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := adapter.GetProtocolForURL(test.url)
			assert.Equal(t, test.expectedProtocol, result)
		})
	}
}

// Authentication workflow tests

func TestAuthenticationAdapter_HTTPSWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		url          string
		token        string
		expectedUser string
		expectedPass string
	}{
		{
			name:         "GitHub HTTPS authentication",
			url:          "https://github.com/owner/repo.git",
			token:        "github_pat_11ABCDEFG0123456789",
			expectedUser: "token",
			expectedPass: "github_pat_11ABCDEFG0123456789",
		},
		{
			name:         "GitLab HTTPS authentication",
			url:          "https://gitlab.com/group/project.git",
			token:        "glpat-xxxxxxxxxxxxxxxxxxxx",
			expectedUser: "token",
			expectedPass: "glpat-xxxxxxxxxxxxxxxxxxxx",
		},
		{
			name:         "Custom domain HTTPS authentication",
			url:          "https://git.company.com/team/project.git",
			token:        "custom_token_12345",
			expectedUser: "token",
			expectedPass: "custom_token_12345",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewAuthenticationAdapter()
			ctx := context.Background()

			// Arrange: Protocol detection from URL
			protocol := adapter.GetProtocolForURL(test.url)
			authConfig := ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeToken,
				Protocol: protocol,
				Token:    test.token,
			}

			// Act: Complete HTTPS authentication workflow
			err := adapter.ValidateAuthConfig(authConfig)
			require.NoError(t, err, "Config validation should succeed for HTTPS")

			authMethod, err := adapter.GetTransportAuth(ctx, authConfig)
			require.NoError(t, err, "Auth method creation should succeed")
			require.NotNil(t, authMethod, "Auth method should be created")

			// Assert: Verify HTTPS auth method configuration
			basicAuth, ok := authMethod.(*http.BasicAuth)
			require.True(t, ok, "Should create BasicAuth for HTTPS")
			assert.Equal(t, test.expectedUser, basicAuth.Username)
			assert.Equal(t, test.expectedPass, basicAuth.Password)
			assert.Equal(t, "https", protocol)
		})
	}
}

func TestAuthenticationAdapter_SSHWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		sshKeyPath string
		sshKey     string
	}{
		{
			name:       "GitLab SSH with key path",
			url:        "git@gitlab.com:owner/repo.git",
			sshKeyPath: filepath.Join(t.TempDir(), ".ssh", "id_ed25519"),
		},
		{
			name:       "GitHub SSH with key path",
			url:        "git@github.com:owner/repo.git",
			sshKeyPath: filepath.Join(t.TempDir(), ".ssh", "id_rsa"),
		},
		{
			name:   "Custom SSH with inline key",
			url:    "git@git.example.com:team/project.git",
			sshKey: "-----BEGIN OPENSSH PRIVATE KEY-----",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock SSH environment for workflow tests
			originalSSHAuthSock := os.Getenv("SSH_AUTH_SOCK")

			defer func() {
				if originalSSHAuthSock != "" {
					_ = os.Setenv("SSH_AUTH_SOCK", originalSSHAuthSock)
				} else {
					_ = os.Unsetenv("SSH_AUTH_SOCK")
				}
			}()

			// Set mock SSH_AUTH_SOCK for test isolation
			mockSocketPath := t.TempDir() + "/ssh-agent.sock"
			_ = os.Setenv("SSH_AUTH_SOCK", mockSocketPath)
			adapter := NewAuthenticationAdapter()
			ctx := context.Background()

			// Arrange: Protocol detection from URL
			protocol := adapter.GetProtocolForURL(test.url)
			authConfig := ports.AuthenticationConfiguration{
				Type:       ports.AuthenticationTypeSSH,
				Protocol:   protocol,
				SSHKeyPath: test.sshKeyPath,
				SSHKey:     test.sshKey,
			}

			// Act: Validate SSH configuration (this should work without SSH agent)
			err := adapter.ValidateAuthConfig(authConfig)
			require.NoError(t, err, "Config validation should succeed for SSH")

			// Test auth method creation with mocked SSH environment
			authMethod, err := adapter.GetTransportAuth(ctx, authConfig)

			// SSH agent creation may fail with mock socket, which is acceptable
			if err != nil {
				// Verify we get expected SSH agent error with mock environment
				assert.Contains(t, err.Error(), "error creating SSH agent")

				return
			}

			// If SSH agent creation succeeded (unlikely with mock), verify method type
			require.NotNil(t, authMethod, "Auth method should be created")
			_, ok := authMethod.(*ssh.PublicKeysCallback)
			require.True(t, ok, "Should create PublicKeysCallback for SSH")
			assert.Equal(t, protocolSSH, protocol)
		})
	}
}

// Protocol validation edge cases

func TestAuthenticationAdapter_ProtocolValidation(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	err := adapter.ValidateAuthConfig(ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeToken,
		Protocol: "HTTPS!@#",
		Token:    "token123",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedProtocol)
}

// Test ValidateAuthConfigWithProvider - enhanced validation with token format checking

func TestAuthenticationAdapter_ValidateAuthConfigWithProvider_ValidTokens(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	tests := []struct {
		name         string
		providerType string
		token        string
		protocol     string
	}{
		// Any valid tokens pass basic validation (no format enforcement)
		{"GitHub-style token", "github", "ghp_123456789012345678901234567890123456", "https"},
		{"GitLab-style token", "gitlab", "glpat-12345678901234567890", "https"},
		{"Gitea-style token", "gitea", "1234567890abcdef1234567890abcdef12345678", "https"},
		{"Generic token", "unknown", "arbitrary_token_12345", "https"},
		{"Bearer token", "api", "Bearer-xyz123", "https"},
		{"Simple token", "custom", "simpletoken", "https"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			authConfig := ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeToken,
				Protocol: test.protocol,
				Token:    test.token,
			}

			err := adapter.ValidateAuthConfigWithProvider(authConfig, test.providerType)
			require.NoError(t, err, "Valid %s token should pass validation", test.providerType)
		})
	}
}

func TestAuthenticationAdapter_ValidateAuthConfigWithProvider_InvalidTokens(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	tests := []struct {
		name         string
		providerType string
		token        string
		protocol     string
		expectedErr  error
	}{
		// Only truly invalid tokens fail basic validation
		{"Token too short", "github", "abc", "https", ErrInvalidTokenFormat},
		{"Token with spaces", "gitlab", "token with spaces", "https", ErrInvalidTokenFormat},
		{"Token with newlines", "gitea", "token\nwith\nnewlines", "https", ErrInvalidTokenFormat},
		{"Token too long", "any", strings.Repeat("x", 513), "https", ErrInvalidTokenFormat},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			authConfig := ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeToken,
				Protocol: test.protocol,
				Token:    test.token,
			}

			err := adapter.ValidateAuthConfigWithProvider(authConfig, test.providerType)
			require.Error(t, err)
			require.ErrorIs(t, err, test.expectedErr)
			assert.Contains(t, err.Error(), "invalid token format")
		})
	}
}

func TestAuthenticationAdapter_ValidateAuthConfigWithProvider_UnknownProvider(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	authConfig := ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeToken,
		Protocol: "https",
		Token:    "any-token-format-should-work",
	}

	err := adapter.ValidateAuthConfigWithProvider(authConfig, "unknown-provider")
	require.NoError(t, err, "Unknown providers should accept any non-empty token")
}

func TestAuthenticationAdapter_ValidateAuthConfigWithProvider_SSHSkipsTokenValidation(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	// SSH should skip token validation even with invalid token format
	authConfig := ports.AuthenticationConfiguration{
		Type:       ports.AuthenticationTypeSSH,
		Protocol:   protocolSSH,
		Token:      "invalid-github-token-format",
		SSHKeyPath: "/path/to/key",
	}

	err := adapter.ValidateAuthConfigWithProvider(authConfig, "github")
	require.NoError(t, err, "SSH protocol should skip token format validation")
}

func TestAuthenticationAdapter_ValidateAuthConfigWithProvider_EmptyToken(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	// Empty token should fail basic validation before format validation
	authConfig := ports.AuthenticationConfiguration{
		Type:     ports.AuthenticationTypeToken,
		Protocol: "https",
		Token:    "",
	}

	err := adapter.ValidateAuthConfigWithProvider(authConfig, "github")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrAuthConfigRequiresCredentials)
}

// Test basic token validation

func TestAuthenticationAdapter_validateTokenBasic(t *testing.T) {
	t.Parallel()

	adapter := NewAuthenticationAdapter()

	validTokens := []string{
		"ghp_123456789012345678901234567890123456", // GitHub format (but we don't enforce format)
		"glpat-12345678901234567890",               // GitLab format (but we don't enforce format)
		"1234567890abcdef1234567890abcdef12345678", // Gitea format (but we don't enforce format)
		"arbitrary_token_string_123",               // Any reasonable token
		"Bearer-token-xyz",                         // Various formats accepted
		"simple123",                                // Short but valid
	}

	for _, token := range validTokens {
		err := adapter.validateTokenBasic(token)
		require.NoError(t, err, "Valid token should pass basic validation: %s", token)
	}

	invalidTokens := []string{
		"abc",                    // too short
		"token with spaces",      // contains spaces
		"token\nwith\nnewlines",  // contains newlines
		strings.Repeat("x", 513), // too long
	}

	for _, token := range invalidTokens {
		err := adapter.validateTokenBasic(token)
		require.Error(t, err, "Invalid token should fail basic validation: %s", token)
		require.ErrorIs(t, err, ErrInvalidTokenFormat)
	}

	// Test empty token (should be allowed - handled elsewhere)
	err := adapter.validateTokenBasic("")
	require.NoError(t, err, "Empty token should be allowed in basic validation")
}

// Test error constants exist and are properly defined

func TestAuthenticationErrors(t *testing.T) {
	t.Parallel()

	// Verify all error constants are defined and non-empty
	errors := []error{
		ErrInvalidAuthProtocol,
		ErrAuthConfigRequiresCredentials,
		ErrSSHRequiresKey,
		ErrHTTPSRequiresToken,
		ErrUnsupportedProtocol,
		ErrInvalidTokenFormat,
	}

	for _, err := range errors {
		require.Error(t, err)
		assert.NotEmpty(t, err.Error())
	}

	// Verify errors are distinct
	errorMessages := make(map[string]bool)

	for _, err := range errors {
		msg := err.Error()
		assert.False(t, errorMessages[msg], "Duplicate error message: %s", msg)
		errorMessages[msg] = true
	}
}
