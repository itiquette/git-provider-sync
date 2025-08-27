// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

var (
	// ErrInvalidLengthURL is returned when a remote has no associated URL.
	ErrInvalidLengthURL = errors.New("remote has no URL")

	// ErrNullRepositoryPtr is returned when attempting to create a Repository with a nil pointer.
	ErrNullRepositoryPtr = errors.New("parameter repositoryPtr was null")
)

// Repository represents a Git repository with additional metadata.
// It encapsulates a go-git Repository and provides methods for common operations.
type Repository struct {
	goGitRepository *git.Repository
	ProjectMetaInfo *ProjectInfo
}

// NewGitGoRemoteOption creates a new RemoteConfig for go-git.
func NewGitGoRemoteOption(name string, urls []string, isMirror bool) config.RemoteConfig {
	return config.RemoteConfig{
		Name:   name,
		URLs:   urls,
		Mirror: isMirror,
	}
}

// GoGitRepository returns the underlying go-git Repository.
func (r Repository) GoGitRepository() *git.Repository {
	return r.goGitRepository
}

// ProjectInfo returns the repository metadata.
// This includes information such as the repository name, description, and URLs.
func (r Repository) ProjectInfo() *ProjectInfo {
	return r.ProjectMetaInfo
}

// Remote retrieves a remote by name.
func (r Repository) Remote(name string) (Remote, error) {
	rem, err := r.goGitRepository.Remote(name)
	if err != nil {
		return Remote{}, fmt.Errorf("failed to get remote '%s': %w", name, err)
	}

	urls := rem.Config().URLs
	if len(urls) == 0 {
		return Remote{}, ErrInvalidLengthURL
	}

	return Remote{URL: urls[0]}, nil
}

// DeleteRemote removes a remote by name.
// If the remote doesn't exist, this operation is treated as successful.
func (r Repository) DeleteRemote(name string) error {
	err := r.goGitRepository.DeleteRemote(name)
	if err != nil && !errors.Is(err, git.ErrRemoteNotFound) {
		return fmt.Errorf("failed to delete remote '%s': %w", name, err)
	}

	return nil
}

// CreateRemote adds a new remote to the repository.
func (r Repository) CreateRemote(name, url string, isMirror bool) error {
	gitRemote := NewGitGoRemoteOption(name, []string{url}, isMirror)

	_, err := r.goGitRepository.CreateRemote(&gitRemote)
	if err != nil {
		return fmt.Errorf("failed to create remote '%s': %w", name, err)
	}

	return nil
}

// NewRepository creates a new Repository instance.
