// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
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
