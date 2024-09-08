// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"strings"
)

// GitProviderClientOption holds configuration options for a git provider client.
// It encapsulates essential information needed to connect to and interact with
// a specific git provider service.
type GitProviderClientOption struct {
	// Provider is the name or identifier of the git provider service (e.g., "github", "gitlab").
	Provider string

	// Token is the authentication token used to access the git provider's API.
	// This field is sensitive and should be handled securely.
	Token string

	// Domain is the domain of the git provider service (e.g., "github.com", "gitlab.com").
	Domain string
}

// String provides a string representation of GitProviderClientOption.
// It formats the fields into a human-readable string, masking the Token for security.
//
// Returns:
//   - A string representation of the GitProviderClientOption instance.
func (gpo GitProviderClientOption) String() string {
	return fmt.Sprintf("GitProviderClientOption{Provider: %s, Token: %s, Domain: %s}",
		gpo.Provider, maskToken(gpo.Token), gpo.Domain)
}

// maskToken is a helper function that masks all but the last 4 characters of a token.
// If the token is 4 characters or less, it masks all characters.
func maskToken(token string) string {
	if len(token) <= 4 {
		return strings.Repeat("*", len(token))
	}

	return strings.Repeat("*", len(token)-4) + token[len(token)-4:]
}

// NewGitProviderClientOption creates a new GitProviderClientOption instance.
//
// Parameters:
//   - provider: A string identifying the git provider service.
//   - token: An authentication token for the git provider's API.
//   - domain: The domain of the git provider service.
//
// Returns:
//   - A new GitProviderClientOption instance.
func NewGitProviderClientOption(provider, token, domain string) GitProviderClientOption {
	return GitProviderClientOption{
		Provider: provider,
		Token:    token,
		Domain:   domain,
	}
}

// Example usage:
//
//	option := NewGitProviderClientOption("github", "ghp_1234567890abcdef", "github.com")
//	fmt.Println(option)
//	// Output: GitProviderClientOption{Provider: github, Token: ************cdef, Domain: github.com}
