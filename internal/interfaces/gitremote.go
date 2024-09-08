// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"itiquette/git-provider-sync/internal/model"
)

// GitRemote defines the interface for managing Git remotes.
// This interface provides methods for creating, deleting, and retrieving
// information about Git remotes, which are essential for synchronizing
// repositories across different Git providers.
type GitRemote interface {
	// CreateRemote adds a new remote to the Git repository.
	//
	// Parameters:
	//   - name: A string representing the name of the new remote (e.g., "origin", "upstream").
	//   - url: A string containing the URL of the remote repository.
	//   - isMirror: A boolean indicating whether this remote should be treated as a mirror.
	//
	// Returns:
	//   - error: An error if the creation fails, or nil if successful.
	//
	// This method should handle scenarios such as:
	//   - Validating the remote name and URL.
	//   - Checking for existing remotes with the same name.
	//   - Configuring the remote with the appropriate settings based on the isMirror flag.
	CreateRemote(name string, url string, isMirror bool) error

	// DeleteRemote removes an existing remote from the Git repository.
	//
	// Parameters:
	//   - name: A string representing the name of the remote to be deleted.
	//
	// Returns:
	//   - error: An error if the deletion fails, or nil if successful.
	//
	// This method should handle scenarios such as:
	//   - Checking if the remote exists before attempting deletion.
	//   - Cleaning up any associated configurations or cached data.
	DeleteRemote(name string) error

	// Remote retrieves information about a specific Git remote.
	//
	// Parameters:
	//   - name: A string representing the name of the remote to retrieve.
	//
	// Returns:
	//   - model.Remote: A struct containing information about the remote.
	//   - error: An error if the remote cannot be found or accessed, or nil if successful.
	//
	// This method should provide details such as:
	//   - The remote's URL.
	//   - Any specific configurations associated with the remote.
	Remote(name string) (model.Remote, error)
}

// Example usage:
//
//	type MyGitRepo struct {
//		// Internal fields for managing the repository
//	}
//
//	func (r *MyGitRepo) CreateRemote(name, url string, isMirror bool) error {
//		// Implementation for creating a new remote
//		// Example:
//		// return r.repo.CreateRemote(&config.RemoteConfig{
//		//     Name: name,
//		//     URLs: []string{url},
//		//     Mirror: isMirror,
//		// })
//	}
//
//	func (r *MyGitRepo) DeleteRemote(name string) error {
//		// Implementation for deleting a remote
//		// Example:
//		// return r.repo.DeleteRemote(name)
//	}
//
//	func (r *MyGitRepo) Remote(name string) (model.Remote, error) {
//		// Implementation for retrieving remote information
//		// Example:
//		// remote, err := r.repo.Remote(name)
//		// if err != nil {
//		//     return model.Remote{}, err
//		// }
//		// return model.Remote{URL: remote.Config().URLs[0]}, nil
//	}
//
//	func SyncRemotes(repo GitRemote) error {
//		// Add a new remote
//		if err := repo.CreateRemote("upstream", "https://github.com/example/repo.git", false); err != nil {
//			return err
//		}
//
//		// Get information about a remote
//		origin, err := repo.Remote("origin")
//		if err != nil {
//			return err
//		}
//		fmt.Printf("Origin URL: %s\n", origin.URL)
//
//		// Delete a remote
//		return repo.DeleteRemote("old-remote")
//	}
