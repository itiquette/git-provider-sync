// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package dto

import (
	"fmt"
)

// RepositoriesOption represents configuration for repository filtering.
type RepositoriesOption struct {
	Exclude []string `koanf:"exclude"`
	Include []string `koanf:"include"`
}

func (r RepositoriesOption) String() string {
	return fmt.Sprintf("RepositoryOption: Exclude %v, Include: %v",
		r.Exclude, r.Include)
}
