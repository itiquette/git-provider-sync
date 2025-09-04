// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"strings"
)

// ProviderOption represents configuration options for a provider.
type ProviderOption struct {
	ExcludedRepositories []string
	IncludeForks         bool
	IncludedRepositories []string
	Owner                string
	OwnerType            string
	User                 string
}

// NewProviderOption creates a new provider option with the given parameters.
func NewProviderOption(
	includeForks bool,
	owner string,
	ownerType string,
	included,
	excluded []string,
) ProviderOption {
	return ProviderOption{
		ExcludedRepositories: excluded,
		IncludeForks:         includeForks,
		IncludedRepositories: included,
		Owner:                owner,
		OwnerType:            ownerType,
	}
}

// IsGroup returns true if the provider option represents a group owner type.
func (pr ProviderOption) IsGroup() bool {
	return strings.EqualFold(pr.OwnerType, GROUP)
}

func (pr ProviderOption) String() string {
	return fmt.Sprintf("ProviderOption{Owner: %s, OwnerType: %s, IncludeForks: %v, IncludedRepositories: %v, ExcludedRepositories: %v}",
		pr.Owner,
		pr.OwnerType,
		pr.IncludeForks,
		pr.IncludedRepositories,
		pr.ExcludedRepositories)
}

const (
	// USER represents user owner type.
	USER string = "user"
	// GROUP represents group owner type.
	GROUP string = "group"
)
