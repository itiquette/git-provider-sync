// SPDX-FileCopyrightText: 2024 Josef Andersson
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
	Meta            RepositoryMetainfo
}

// NewGitGoRemoteOption creates a new RemoteConfig for go-git.
// This function is used internally to create remote configurations.
//
// Parameters:
//   - name: The name of the remote (e.g., "origin", "upstream").
//   - urls: A slice of URLs associated with the remote.
//   - isMirror: A boolean indicating if this is a mirror repository.
//
// Returns:
//   - A config.RemoteConfig struct ready to be used with go-git.
func NewGitGoRemoteOption(name string, urls []string, isMirror bool) config.RemoteConfig {
	return config.RemoteConfig{
		Name:   name,
		URLs:   urls,
		Mirror: isMirror,
	}
}

// GoGitRepository returns the underlying go-git Repository.
// This method provides access to the full functionality of go-git if needed.
func (r Repository) GoGitRepository() *git.Repository {
	return r.goGitRepository
}

// Metainfo returns the repository metadata.
// This includes information such as the repository name, description, and URLs.
func (r Repository) Metainfo() RepositoryMetainfo {
	return r.Meta
}

// Remote retrieves a remote by name.
//
// Parameters:
//   - name: The name of the remote to retrieve.
//
// Returns:
//   - A Remote struct containing the URL of the specified remote.
//   - An error if the remote doesn't exist or has no URL.
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
//
// Parameters:
//   - name: The name of the remote to delete.
//
// Returns:
//   - An error if the deletion fails for reasons other than the remote not existing.
func (r Repository) DeleteRemote(name string) error {
	err := r.goGitRepository.DeleteRemote(name)
	if err != nil && !errors.Is(err, git.ErrRemoteNotFound) {
		return fmt.Errorf("failed to delete remote '%s': %w", name, err)
	}

	return nil
}

// CreateRemote adds a new remote to the repository.
//
// Parameters:
//   - name: The name of the new remote.
//   - url: The URL of the new remote.
//   - isMirror: A boolean indicating if this is a mirror repository.
//
// Returns:
//   - An error if the creation fails.
func (r Repository) CreateRemote(name, url string, isMirror bool) error {
	gitRemote := NewGitGoRemoteOption(name, []string{url}, isMirror)

	_, err := r.goGitRepository.CreateRemote(&gitRemote)
	if err != nil {
		return fmt.Errorf("failed to create remote '%s': %w", name, err)
	}

	return nil
}

// NewRepository creates a new Repository instance.
// It wraps a go-git Repository pointer with additional metadata.
//
// Parameters:
//   - repositoryPtr: A pointer to a go-git Repository.
//
// Returns:
//   - A new Repository instance.
//   - An error if the provided pointer is nil.
func NewRepository(repositoryPtr *git.Repository) (Repository, error) {
	if repositoryPtr == nil {
		return Repository{}, ErrNullRepositoryPtr
	}

	return Repository{goGitRepository: repositoryPtr}, nil
}
