// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"strings"
)

type RepositoriesOption struct {
	Exclude     string  `koanf:"exclude"`
	Include     string  `koanf:"include"`
	Description *string `koanf:"description"`
	InMem       bool    `koanf:"inmem"`
}

func (r RepositoriesOption) String() string {
	return fmt.Sprintf("RepositoryOption: Exclude %v, Include: %v",
		r.Exclude, r.Include)
}

// IncludedRepositories returns a slice of included repository names.
func (r RepositoriesOption) IncludedRepositories() []string {
	return splitAndTrim(r.Include)
}

// ExcludedRepositories returns a slice of excluded repository names.
func (r RepositoriesOption) ExcludedRepositories() []string {
	return splitAndTrim(r.Exclude)
}

func splitAndTrim(s string) []string {
	return strings.Split(strings.ReplaceAll(s, " ", ""), ",")
}
