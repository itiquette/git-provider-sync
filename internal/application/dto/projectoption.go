// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package dto

import "strconv"

// ProjectOption represents project configuration options.
type ProjectOption struct {
	Description string `koanf:"description"` // Human-readable description for mirrored repositories
	Disabled    bool   `koanf:"disabled"`    // When true, project is skipped during sync (default: false)
	Visibility  string `koanf:"visibility"`  // Repository visibility: "public", "private", or "internal"
}

// NewProjectOption creates a new ProjectOption with default values.
func NewProjectOption() *ProjectOption {
	return &ProjectOption{
		Description: "",
		Disabled:    false,
		Visibility:  "",
	}
}

func (p ProjectOption) String() string {
	return "ProjectOption: Type: " + p.Description + ", Disabled: " + strconv.FormatBool(p.Disabled) + ", Visibility: " + p.Visibility
}
