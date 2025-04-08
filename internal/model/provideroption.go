// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"strings"
)

type ProviderOption struct {
	ExcludedRepositories []string
	IncludeForks         bool
	IncludedRepositories []string
	Owner                string
	OwnerType            string
	User                 string
}

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
	USER  string = "user"
	GROUP string = "group"
)
