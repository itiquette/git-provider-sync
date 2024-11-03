// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	model "itiquette/git-provider-sync/internal/model/configuration"
	"strings"

	"github.com/rs/zerolog"
)

// GitProviderClientOption holds configuration options for a git provider client.
// It encapsulates essential information needed to connect to and interact with
// a specific git provider service.
type GitProviderClientOption struct {
	// ProviderType is the name or identifier of the git provider service (e.g., "github", "gitlab").
	ProviderType string

	HTTPClient model.HTTPClientOption

	// Domain is the domain of the git provider service (e.g., "github.com", "gitlab.com").
	Domain string

	Repositories model.RepositoriesOption

	UploadURL string
}

// String provides a safe string representation without exposing sensitive data.
func (gpo GitProviderClientOption) String() string {
	return fmt.Sprintf("GitProviderClientOption{type: %s, domain: %s, httpClient: %v}",
		gpo.ProviderType,
		gpo.Domain,
		gpo.HTTPClient.String(),
	)
}

// DebugLog provides detailed logging while protecting sensitive information.
func (gpo GitProviderClientOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint
				Str("provider_type", gpo.ProviderType).
				Str("domain", gpo.Domain).
				Str("httpclient", gpo.String()).
				Interface("repositories", gpo.Repositories)
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

// NewGitProviderClientOption creates a new GitProviderClientOption instance.
//
// Parameters:
//   - provider: A string identifying the git provider service.
//   - token: An authentication token for the git provider's API.
//   - domain: The domain of the git provider service.
//
// Returns:
//   - A new GitProviderClientOption instance.
func NewGitProviderClientOption(providerType string, httpClient model.HTTPClientOption, domain string) GitProviderClientOption {
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
