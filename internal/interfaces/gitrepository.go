// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"itiquette/git-provider-sync/internal/model"

	"github.com/go-git/go-git/v5"
)

// GitRepository defines the interface for interacting with a Git repository.
// It extends the GitRemote interface and provides additional methods for
// accessing the underlying go-git repository and repository metadata.
type GitRepository interface {
	// GitRemote embeds all methods from the GitRemote interface.
	// This allows GitRepository to handle remote-related operations.
	GitRemote

	// GoGitRepository returns the underlying go-git Repository object.
	// This method provides access to the full functionality of go-git
	// for operations not covered by higher-level abstractions.
	//
	// Returns:
	//   - *git.Repository: A pointer to the underlying go-git Repository.
	GoGitRepository() *git.Repository

	// ProjectInfo returns metadata about the repository.
	// This method provides access to additional information about the repository
	// that may not be directly available from the go-git Repository object.
	//
	// Returns:
	//   - model.RepositoryMetainfo: A struct containing metadata about the repository.
	ProjectInfo() *model.ProjectInfo
}

// Note: The GitRemote interface is not defined in this snippet.
// It should be defined elsewhere in the package, containing methods
// for interacting with Git remotes.

// Example usage:
//
//	type MyGitRepo struct {
//		repo *git.Repository
//		meta model.RepositoryMetainfo
//	}
//
//	func (r *MyGitRepo) GoGitRepository() *git.Repository {
//		return r.repo
//	}
//
//	func (r *MyGitRepo) Metainfo() model.RepositoryMetainfo {
//		return r.meta
//	}
//
//	// Implement other methods from GitRemote interface...
//
//	func UseGitRepository(repo GitRepository) {
//		// Access go-git repository
//		gitRepo := repo.GoGitRepository()
//		head, _ := gitRepo.Head()
//		fmt.Println("Current HEAD:", head.Hash())
//
//		// Access metadata
//		meta := repo.Metainfo()
//		fmt.Println("Repository name:", meta.Name)
//
//		// Use GitRemote methods
//		remote, _ := repo.Remote("origin")
//		fmt.Println("Origin URL:", remote.URL)
//	}
