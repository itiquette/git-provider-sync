// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
)

type RepositoriesOption struct {
	Exclude []string `koanf:"exclude"`
	Include []string `koanf:"include"`
}

func (r RepositoriesOption) String() string {
	return fmt.Sprintf("RepositoryOption: Exclude %v, Include: %v",
		r.Exclude, r.Include)
}

// IncludedRepositories returns a slice of included repository names.
func (r RepositoriesOption) IncludedRepositories() []string {
	return r.Include
}

// ExcludedRepositories returns a slice of excluded repository names.
func (r RepositoriesOption) ExcludedRepositories() []string {
	return r.Exclude
}
