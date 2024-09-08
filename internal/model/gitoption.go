// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"regexp"

	"github.com/rs/zerolog"
)

// basicAuthRegEx is used to identify and replace basic auth information in URLs.
var basicAuthRegEx = regexp.MustCompile(`//.*@`)

// PushOption represents options for a git push operation.
// It encapsulates the target repository, reference specifications,
// and flags for pruning and force pushing.
type PushOption struct {
	Target   string   // The URL of the target repository
	RefSpecs []string // The reference specifications to push
	Prune    bool     // Whether to prune remote branches that no longer exist locally
	Force    bool     // Whether to force push (overwrite remote history)
}

// ScrubTarget removes sensitive information from the target URL.
// This method is useful for logging or displaying the PushOption
// without revealing authentication credentials.
//
// Returns:
//   - A string with the authentication part of the URL replaced by "****:***".
func (po PushOption) ScrubTarget() string {
	return basicAuthRegEx.ReplaceAllString(po.Target, "//****:***@")
}

// String provides a string representation of PushOption.
// It uses ScrubTarget to ensure sensitive information is not exposed.
//
// Returns:
//   - A string representation of the PushOption struct.
func (po PushOption) String() string {
	return fmt.Sprintf("PushOption{Target: %s, RefSpecs: %v, Prune: %t, Force: %t}",
		po.ScrubTarget(), po.RefSpecs, po.Prune, po.Force)
}

// DebugLog creates a debug log event with repository metadata.
// This method is useful for debugging and tracing Push option operations.
func (po PushOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("target", po.ScrubTarget()).
				Strs("refspecs", po.RefSpecs).
				Bool("prune", po.Prune).
				Bool("force", po.Force)
}

// NewPushOption creates a new PushOption with appropriate RefSpecs.
// It automatically sets up the correct reference specifications based on
// whether a force push is requested.
//
// Parameters:
//   - target: The URL of the target repository.
//   - prune: Whether to prune remote branches.
//   - force: Whether to force push.
//
// Returns:
//   - A new PushOption struct configured with the provided options.
func NewPushOption(target string, prune, force bool) PushOption {
	refSpecs := []string{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"}
	if force {
		for i, spec := range refSpecs {
			refSpecs[i] = "+" + spec
		}
	}

	return PushOption{
		Target:   target,
		RefSpecs: refSpecs,
		Prune:    prune,
		Force:    force,
	}
}

// PullOption represents options for a git pull operation.
// It includes the name of the remote, its URL, and the local target path.
type PullOption struct {
	Name       string // The name of the remote (e.g., "origin")
	URL        string // The URL of the remote repository
	TargetPath string // The local path where the repository will be pulled
}

// String provides a string representation of PullOption.
//
// Returns:
//   - A string representation of the PullOption struct.
func (po PullOption) String() string {
	return fmt.Sprintf("PullOption{Name: %s, URL: %s, TargetPath: %s}",
		po.Name, po.URL, po.TargetPath)
}

// DebugLog creates a debug log event with repository metadata.
// This method is useful for debugging and tracing Pull option operations.
func (po PullOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("target", po.Name).
				Str("refspecs", po.TargetPath).
				Str("prune", po.URL)
}

// NewPullOption creates a new PullOption.
//
// Parameters:
//   - name: The name of the remote.
//   - url: The URL of the remote repository.
//   - targetPath: The local path where the repository will be pulled.
//
// Returns:
//   - A new PullOption struct configured with the provided options.
func NewPullOption(name, url, targetPath string) PullOption {
	return PullOption{Name: name, URL: url, TargetPath: targetPath}
}

// CloneOption represents options for a git clone operation.
// It includes flags for cleaning up names, mirroring, and specifies
// the source URL and target path.
type CloneOption struct {
	CleanupName bool   // Whether to clean up the repository name
	URL         string // The URL of the repository to clone
	Mirror      bool   // Whether to create a mirror clone
	TargetPath  string // The local path where the repository will be cloned
}

// NewCloneOption creates a new CloneOption.
//
// Parameters:
//   - url: The URL of the repository to clone.
//   - mirror: Whether to create a mirror clone.
//   - targetPath: The local path where the repository will be cloned.
//
// Returns:
//   - A new CloneOption struct configured with the provided options.
func NewCloneOption(url string, mirror bool, targetPath string) CloneOption {
	return CloneOption{URL: url, Mirror: mirror, TargetPath: targetPath}
}

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
