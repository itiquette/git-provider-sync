// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	protocolSSH   = "ssh"
	protocolHTTPS = "https"
	protocolTLS   = "tls"
)

// Static errors for err113 compliance.
var (
	ErrInvalidAuthProtocol           = errors.New("invalid authentication protocol")
	ErrAuthConfigRequiresCredentials = errors.New("authentication configuration requires either token, SSH key path, or SSH key content")
	ErrSSHRequiresKey                = errors.New("SSH protocol requires SSH key path or SSH key content")
	ErrHTTPSRequiresToken            = errors.New("HTTPS/TLS protocol requires token")
	ErrUnsupportedProtocol           = errors.New("unsupported protocol")
	ErrInvalidTokenFormat            = errors.New("invalid token format")
)

// Service implements authentication methods for git operations.
type Service struct{}

// NewService creates a new authentication service.
func NewService() *Service {
	return &Service{}
}

// GetAuthMethod returns the appropriate authentication method based on configuration
//
//	exact main branch GetAuthMethod functionality.
func (s *Service) GetAuthMethod(_ context.Context, authConfig ports.AuthenticationConfiguration) (transport.AuthMethod, error) { //nolint:ireturn
	protocol := strings.ToLower(authConfig.Protocol)

	switch protocol {
	case protocolSSH:
		// Use SSH agent authentication as in main branch
		auth, err := ssh.NewSSHAgentAuth("git")
		if err != nil {
			return nil, fmt.Errorf("error creating SSH agent: %w", err)
		}

		return auth, nil
	case protocolTLS, protocolHTTPS, "":
		// Use standard token authentication with conventional username
		return &http.BasicAuth{
			Username: "token", // Standard convention for token-based auth
			Password: authConfig.Token,
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidAuthProtocol, protocol)
	}
}

// AuthenticationAdapter adapts the domain authentication to transport layer.
type AuthenticationAdapter struct {
	service *Service
}

// NewAuthenticationAdapter creates a new authentication adapter.
func NewAuthenticationAdapter() *AuthenticationAdapter {
	return &AuthenticationAdapter{
		service: NewService(),
	}
}

// GetTransportAuth returns go-git transport authentication method.
func (a *AuthenticationAdapter) GetTransportAuth(ctx context.Context, authConfig ports.AuthenticationConfiguration) (transport.AuthMethod, error) { //nolint:ireturn
	return a.service.GetAuthMethod(ctx, authConfig)
}

// ValidateTokenBasic performs basic token validation - non-empty and reasonable length.
func (a *AuthenticationAdapter) validateTokenBasic(token string) error {
	if token == "" {
		return nil // Empty token validation handled elsewhere
	}

	// Basic sanity checks only - avoid security theatre of format validation
	if len(token) < 4 {
		return fmt.Errorf("%w: token too short", ErrInvalidTokenFormat)
	}

	if len(token) > 512 {
		return fmt.Errorf("%w: token too long", ErrInvalidTokenFormat)
	}

	// Check for obvious non-token strings
	if strings.Contains(token, " ") || strings.Contains(token, "\n") {
		return fmt.Errorf("%w: token contains invalid characters", ErrInvalidTokenFormat)
	}

	return nil
}

// ValidateAuthConfig validates the authentication configuration.
func (a *AuthenticationAdapter) ValidateAuthConfig(authConfig ports.AuthenticationConfiguration) error {
	if authConfig.Token == "" && authConfig.SSHKeyPath == "" && authConfig.SSHKey == "" {
		return ErrAuthConfigRequiresCredentials
	}

	protocol := strings.ToLower(authConfig.Protocol)
	switch protocol {
	case "ssh":
		if authConfig.SSHKeyPath == "" && authConfig.SSHKey == "" {
			return ErrSSHRequiresKey
		}
	case protocolTLS, protocolHTTPS, "":
		if authConfig.Token == "" {
			return ErrHTTPSRequiresToken
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedProtocol, protocol)
	}

	return nil
}

// ValidateAuthConfigWithProvider validates the authentication configuration with basic token validation.
func (a *AuthenticationAdapter) ValidateAuthConfigWithProvider(authConfig ports.AuthenticationConfiguration, _ string) error {
	// First run basic validation
	if err := a.ValidateAuthConfig(authConfig); err != nil {
		return err
	}

	// Add basic token validation for HTTPS/TLS protocols - avoid format validation theatre
	protocol := strings.ToLower(authConfig.Protocol)
	if (protocol == protocolTLS || protocol == protocolHTTPS || protocol == "") && authConfig.Token != "" {
		if err := a.validateTokenBasic(authConfig.Token); err != nil {
			return err
		}
	}

	return nil
}

// SupportsSSHAgent checks if SSH agent authentication is available.
func (a *AuthenticationAdapter) SupportsSSHAgent() bool {
	// Check if SSH_AUTH_SOCK environment variable is set
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock == "" {
		return false
	}

	// Try to connect to the SSH agent socket to verify it's actually available
	dialer := &net.Dialer{}

	conn, err := dialer.Dial("unix", sshAuthSock)
	if err != nil {
		return false
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// Intentionally ignore close error - connection test was successful
			_ = closeErr
		}
	}()

	return true
}

// GetProtocolForURL determines the appropriate protocol for a given URL.
func (a *AuthenticationAdapter) GetProtocolForURL(url string) string {
	if strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://") {
		return "ssh"
	}

	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return "https"
	}

	// Default to HTTPS for unknown formats
	return "https"
}
