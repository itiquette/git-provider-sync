// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"

	"github.com/rs/zerolog"
)

// CreateProjectOption represents options for creating a new repository.
// It includes the repository name, visibility settings, description,
// and the name of the default branch.
type CreateProjectOption struct {
	Owner          string
	IsGroup        bool
	RepositoryName string // The name of the new repository
	Visibility     string // The visibility setting (e.g., "public", "private")
	Description    string // A description of the repository
	DefaultBranch  string // The name of the default branch (e.g., "main", "master")
	Disabled       bool
}

// String provides a string representation of CreateOption.
func (co CreateProjectOption) String() string {
	return fmt.Sprintf("CreateOption{RepositoryName: %s, Visibility: %s, Description: %s, DefaultBranch: %s, Disabled: %t}",
		co.RepositoryName,
		co.Visibility,
		co.Description,
		co.DefaultBranch,
		co.Disabled)
}

// DebugLog creates a debug log event with repository creation options.
func (co CreateProjectOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("repository_name", co.RepositoryName).
				Str("visibility", co.Visibility).
				Str("description", co.Description).
				Str("default_branch", co.DefaultBranch).
				Bool("Disabled", co.Disabled)
}

// NewCreateOption creates a new CreateOption.
func NewCreateOption(repoName, visibility, description, defaultBranch string, disabled bool) CreateProjectOption {
	return CreateProjectOption{
		RepositoryName: repoName,
		Visibility:     visibility,
		Description:    description,
		DefaultBranch:  defaultBranch,
		Disabled:       disabled,
	}
}
