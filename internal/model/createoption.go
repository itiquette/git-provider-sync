// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

// CreateOption represents options for creating a new repository.
// It includes the repository name, visibility settings, description,
// and the name of the default branch.
type CreateOption struct {
	RepositoryName string // The name of the new repository
	Visibility     string // The visibility setting (e.g., "public", "private")
	Description    string // A description of the repository
	DefaultBranch  string // The name of the default branch (e.g., "main", "master")
}

// NewCreateOption creates a new CreateOption.
//
// Parameters:
//   - repoName: The name of the new repository.
//   - visibility: The visibility setting for the repository.
//   - description: A description of the repository.
//   - defaultBranch: The name of the default branch.
//
// Returns:
//   - A new CreateOption struct configured with the provided options.
func NewCreateOption(repoName, visibility, description, defaultBranch string) CreateOption {
	return CreateOption{
		RepositoryName: repoName,
		Visibility:     visibility,
		Description:    description,
		DefaultBranch:  defaultBranch,
	}
}
