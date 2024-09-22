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
	// ProviderType is the name or identifier of the git provider service (e.g., "github", "gitlab").
	ProviderType string

	HTTPClient HTTPClientOption

	// Domain is the domain of the git provider service (e.g., "github.com", "gitlab.com").
	Domain string

	// Scheme is the scheme of the git provider service (e.g., "https", "http". Default if empty is https).
	Scheme string

	Repositories RepositoriesOption
}

// String provides a string representation of GitProviderClientOption.
// It formats the fields into a human-readable string, masking the Token for security.
//
// Returns:
//   - A string representation of the GitProviderClientOption instance.
func (gpo GitProviderClientOption) String() string {
	return fmt.Sprintf("GitProviderClientOption{ProviderType: %s, HTTPClient: %+v, Domain: %s, Scheme: %s}",
		gpo.ProviderType, gpo.HTTPClient, gpo.Domain, gpo.Scheme)
}

func (gpo GitProviderClientOption) DomainWithScheme(scheme string) string {
	if len(scheme) > 0 {
		return scheme + "://" + gpo.Domain
	}

	httpsPrefix := "https://"

	if !strings.HasPrefix(gpo.Domain, httpsPrefix) {
		return httpsPrefix + gpo.Domain
	}

	return gpo.Domain
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
func NewGitProviderClientOption(providerType string, httpClient HTTPClientOption, domain string) GitProviderClientOption {
	return GitProviderClientOption{
		ProviderType: providerType,
		HTTPClient:   httpClient,
		Domain:       domain,
	}
}

// Example usage:
//
//	option := NewGitProviderClientOption("github", "ghp_1234567890abcdef", "github.com")
//	fmt.Println(option)
//	// Output: GitProviderClientOption{Provider: github, Token: ************cdef, Domain: github.com}
