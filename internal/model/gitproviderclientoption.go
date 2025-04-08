// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
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

	AuthCfg model.AuthConfig

	// Domain is the domain of the git provider service (e.g., "github.com", "gitlab.com").
	Domain string

	Repositories model.RepositoriesOption

	UploadURL string
}

// String provides a safe string representation without exposing sensitive data.
func (gpo GitProviderClientOption) String() string {
	return fmt.Sprintf("ProviderClientOption{ProviderType: %s, Domain: %s, AuthCfg: %v}",
		gpo.ProviderType,
		gpo.Domain,
		gpo.AuthCfg.String(),
	)
}

// DebugLog provides detailed logging while protecting sensitive information.
func (gpo GitProviderClientOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint
				Str("ProviderType", gpo.ProviderType).
				Str("Domain", gpo.Domain).
				Str("ProviderClientOption", gpo.String()).
				Interface("Repositories", gpo.Repositories)
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
func NewGitProviderClientOption(providerType string, auth model.AuthConfig, domain string) GitProviderClientOption {
	return GitProviderClientOption{
		ProviderType: providerType,
		AuthCfg:      auth,
		Domain:       domain,
	}
}

// Example usage:
//
//	option := NewGitProviderClientOption("github", "ghp_1234567890abcdef", "github.com")
//	fmt.Println(option)
//	// Output: GitProviderClientOption{Provider: github, Token: ************cdef, Domain: github.com}
