// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package utilities

import (
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/domain/constants"
)

// VisibilityMapper provides sophisticated visibility mapping between providers.
// This restores the mapVisibility functionality from main branch in hexagonal style.
type VisibilityMapper struct{}

// NewVisibilityMapper creates a new visibility mapper.
func NewVisibilityMapper() *VisibilityMapper {
	return &VisibilityMapper{}
}

// MapVisibility maps repository visibility between different providers.
// This restores the exact functionality from main branch mapVisibility function.
func (vm *VisibilityMapper) MapVisibility(fromProvider, toProvider, visibility string) (string, error) {
	// Normalize inputs to lowercase
	fromProvider = strings.ToLower(fromProvider)
	toProvider = strings.ToLower(toProvider)
	visibility = strings.ToLower(visibility)

	// If providers are the same, no mapping needed
	if strings.EqualFold(fromProvider, toProvider) {
		return visibility, nil
	}

	// Define mapping based on the provider visibility compatibility tables (restored from main branch)
	mappings := map[string]map[string]map[string]string{
		"gitlab": {
			"github": {"public": "public", "internal": "private", "private": "private"},
			"gitea":  {"public": "public", "internal": "private", "private": "private"},
		},
		"github": {
			"gitlab": {"public": "public", "private": "private"},
			"gitea":  {"public": "public", "private": "private"},
		},
		"gitea": {
			"gitlab": {"public": "public", "private": "private", "limited": "private"},
			"github": {"public": "public", "private": "private", "limited": "private"},
		},
	}

	// Check if the fromProvider is valid
	providerMap, isOK := mappings[fromProvider]
	if !isOK {
		return "", fmt.Errorf("invalid source provider: %s", fromProvider)
	}

	// Check if the toProvider is valid
	targetMap, isOK := providerMap[toProvider]
	if !isOK {
		return "", fmt.Errorf("invalid target provider: %s", toProvider)
	}

	// Check if the visibility is valid and get the mapped visibility
	mappedVisibility, isOK := targetMap[visibility]
	if !isOK {
		return "", fmt.Errorf("invalid visibility for %s: %s", fromProvider, visibility)
	}

	return mappedVisibility, nil
}

// MapToConstants maps provider-specific visibility to domain constants.
func (vm *VisibilityMapper) MapToConstants(provider, visibility string) string {
	provider = strings.ToLower(provider)
	visibility = strings.ToLower(visibility)

	switch provider {
	case "gitlab":
		switch visibility {
		case "public":
			return constants.VisibilityPublic
		case "internal", "private":
			return constants.VisibilityPrivate
		default:
			return constants.VisibilityPrivate // Default to private for safety
		}
	case "github":
		switch visibility {
		case "public":
			return constants.VisibilityPublic
		case "private":
			return constants.VisibilityPrivate
		default:
			return constants.VisibilityPrivate // Default to private for safety
		}
	case "gitea":
		switch visibility {
		case "public":
			return constants.VisibilityPublic
		case "private", "limited":
			return constants.VisibilityPrivate
		default:
			return constants.VisibilityPrivate // Default to private for safety
		}
	default:
		return constants.VisibilityPrivate // Default to private for unknown providers
	}
}

// MapFromConstants maps domain constants to provider-specific visibility.
func (vm *VisibilityMapper) MapFromConstants(provider, visibility string) string {
	provider = strings.ToLower(provider)

	switch provider {
	case "gitlab":
		switch visibility {
		case constants.VisibilityPublic:
			return "public"
		case constants.VisibilityPrivate:
			return "private"
		default:
			return "private" // Default to private for safety
		}
	case "github":
		switch visibility {
		case constants.VisibilityPublic:
			return "public"
		case constants.VisibilityPrivate:
			return "private"
		default:
			return "private" // Default to private for safety
		}
	case "gitea":
		switch visibility {
		case constants.VisibilityPublic:
			return "public"
		case constants.VisibilityPrivate:
			return "private"
		default:
			return "private" // Default to private for safety
		}
	default:
		return "private" // Default to private for unknown providers
	}
}

// GetSupportedVisibilities returns the supported visibility options for a provider.
func (vm *VisibilityMapper) GetSupportedVisibilities(provider string) []string {
	provider = strings.ToLower(provider)

	switch provider {
	case "gitlab":
		return []string{"public", "internal", "private"}
	case "github":
		return []string{"public", "private"}
	case "gitea":
		return []string{"public", "private", "limited"}
	default:
		return []string{"public", "private"} // Default common set
	}
}

// IsValidVisibility checks if a visibility is valid for a provider.
func (vm *VisibilityMapper) IsValidVisibility(provider, visibility string) bool {
	supportedVisibilities := vm.GetSupportedVisibilities(provider)
	visibility = strings.ToLower(visibility)

	for _, supported := range supportedVisibilities {
		if strings.EqualFold(visibility, supported) {
			return true
		}
	}

	return false
}

// GetDefaultVisibility returns the default visibility for a provider.
func (vm *VisibilityMapper) GetDefaultVisibility(provider string) string {
	provider = strings.ToLower(provider)

	switch provider {
	case "gitlab", "github", "gitea":
		return "private" // Most secure default
	default:
		return "private" // Safe default for unknown providers
	}
}

// VisibilityMappingInfo provides information about visibility mapping capabilities.
type VisibilityMappingInfo struct {
	SupportedProviders []string                            `json:"supported_providers"`
	ProviderMappings   map[string]map[string]string        `json:"provider_mappings"`
	SupportedOptions   map[string][]string                 `json:"supported_options"`
	DefaultOptions     map[string]string                   `json:"default_options"`
	MappingRules       map[string]map[string][]MappingRule `json:"mapping_rules"`
}

// MappingRule describes how visibility is mapped between providers.
type MappingRule struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Description string `json:"description"`
}

// GetMappingInfo returns comprehensive information about visibility mapping.
func (vm *VisibilityMapper) GetMappingInfo() VisibilityMappingInfo {
	return VisibilityMappingInfo{
		SupportedProviders: []string{"gitlab", "github", "gitea"},
		ProviderMappings: map[string]map[string]string{
			"gitlab_to_github": {"public": "public", "internal": "private", "private": "private"},
			"gitlab_to_gitea":  {"public": "public", "internal": "private", "private": "private"},
			"github_to_gitlab": {"public": "public", "private": "private"},
			"github_to_gitea":  {"public": "public", "private": "private"},
			"gitea_to_gitlab":  {"public": "public", "private": "private", "limited": "private"},
			"gitea_to_github":  {"public": "public", "private": "private", "limited": "private"},
		},
		SupportedOptions: map[string][]string{
			"gitlab": {"public", "internal", "private"},
			"github": {"public", "private"},
			"gitea":  {"public", "private", "limited"},
		},
		DefaultOptions: map[string]string{
			"gitlab": "private",
			"github": "private",
			"gitea":  "private",
		},
		MappingRules: map[string]map[string][]MappingRule{
			"gitlab": {
				"github": {
					{From: "public", To: "public", Description: "Public repositories remain public"},
					{From: "internal", To: "private", Description: "Internal repositories become private (GitHub doesn't support internal)"},
					{From: "private", To: "private", Description: "Private repositories remain private"},
				},
				"gitea": {
					{From: "public", To: "public", Description: "Public repositories remain public"},
					{From: "internal", To: "private", Description: "Internal repositories become private (Gitea doesn't support internal)"},
					{From: "private", To: "private", Description: "Private repositories remain private"},
				},
			},
			"github": {
				"gitlab": {
					{From: "public", To: "public", Description: "Public repositories remain public"},
					{From: "private", To: "private", Description: "Private repositories remain private"},
				},
				"gitea": {
					{From: "public", To: "public", Description: "Public repositories remain public"},
					{From: "private", To: "private", Description: "Private repositories remain private"},
				},
			},
			"gitea": {
				"gitlab": {
					{From: "public", To: "public", Description: "Public repositories remain public"},
					{From: "private", To: "private", Description: "Private repositories remain private"},
					{From: "limited", To: "private", Description: "Limited repositories become private (GitLab doesn't support limited)"},
				},
				"github": {
					{From: "public", To: "public", Description: "Public repositories remain public"},
					{From: "private", To: "private", Description: "Private repositories remain private"},
					{From: "limited", To: "private", Description: "Limited repositories become private (GitHub doesn't support limited)"},
				},
			},
		},
	}
}
