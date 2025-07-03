// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Service implements authentication methods for git operations.
// This restores the gitlib/authservice.go functionality from main branch.
type Service struct{}

// NewService creates a new authentication service.
func NewService() *Service {
	return &Service{}
}

// GetAuthMethod returns the appropriate authentication method based on configuration.
// This restores the exact main branch GetAuthMethod functionality.
func (s *Service) GetAuthMethod(ctx context.Context, authConfig ports.AuthenticationConfiguration) (transport.AuthMethod, error) {
	protocol := strings.ToLower(authConfig.Protocol)

	switch protocol {
	case "ssh":
		// Use SSH agent authentication as in main branch
		return ssh.NewSSHAgentAuth("git")
	case "tls", "https", "":
		// Use basic auth with token as in main branch
		return &http.BasicAuth{
			Username: "anyUser", // Main branch uses "anyUser" as username
			Password: authConfig.Token,
		}, nil
	default:
		return nil, fmt.Errorf("invalid authentication protocol: %s", protocol)
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
func (a *AuthenticationAdapter) GetTransportAuth(ctx context.Context, authConfig ports.AuthenticationConfiguration) (transport.AuthMethod, error) {
	return a.service.GetAuthMethod(ctx, authConfig)
}

// ValidateAuthConfig validates the authentication configuration.
func (a *AuthenticationAdapter) ValidateAuthConfig(authConfig ports.AuthenticationConfiguration) error {
	if authConfig.Token == "" && authConfig.SSHKeyPath == "" && authConfig.SSHKey == "" {
		return fmt.Errorf("authentication configuration requires either token, SSH key path, or SSH key content")
	}

	protocol := strings.ToLower(authConfig.Protocol)
	switch protocol {
	case "ssh":
		if authConfig.SSHKeyPath == "" && authConfig.SSHKey == "" {
			return fmt.Errorf("SSH protocol requires SSH key path or SSH key content")
		}
	case "tls", "https", "":
		if authConfig.Token == "" {
			return fmt.Errorf("HTTPS/TLS protocol requires token")
		}
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return nil
}

// SupportsSSHAgent checks if SSH agent authentication is available.
func (a *AuthenticationAdapter) SupportsSSHAgent() bool {
	// SSH agent support depends on system configuration
	// For now, we assume it's available if SSH is being used
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
