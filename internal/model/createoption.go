// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"

	"github.com/rs/zerolog"
)

// CreateOption represents options for creating a new repository.
// It includes the repository name, visibility settings, description,
// and the name of the default branch.
type CreateOption struct {
	RepositoryName string // The name of the new repository
	Visibility     string // The visibility setting (e.g., "public", "private")
	Description    string // A description of the repository
	DefaultBranch  string // The name of the default branch (e.g., "main", "master")
	CIEnabled      bool
}

// String provides a string representation of CreateOption.
func (co CreateOption) String() string {
	return fmt.Sprintf("CreateOption{RepositoryName: %s, Visibility: %s, Description: %s, DefaultBranch: %s, CIEnabled: %t}",
		co.RepositoryName,
		co.Visibility,
		co.Description,
		co.DefaultBranch,
		co.CIEnabled)
}

// DebugLog creates a debug log event with repository creation options.
func (co CreateOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("repository_name", co.RepositoryName).
				Str("visibility", co.Visibility).
				Str("description", co.Description).
				Str("default_branch", co.DefaultBranch).
				Bool("CIEnabled", co.CIEnabled)
}

// NewCreateOption creates a new CreateOption.
func NewCreateOption(repoName, visibility, description, defaultBranch string, ciEnabled bool) CreateOption {
	return CreateOption{
		RepositoryName: repoName,
		Visibility:     visibility,
		Description:    description,
		DefaultBranch:  defaultBranch,
		CIEnabled:      ciEnabled,
	}
}
