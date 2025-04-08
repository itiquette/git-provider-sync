// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import "errors"

var (
	ErrBranchCheckout      = errors.New("failed to checkout branch")
	ErrCloneRepository     = errors.New("failed to clone repository")
	ErrPullRepository      = errors.New("failed to pull repository")
	ErrNewRepositoryModel  = errors.New("failed to create new model.repository")
	ErrFetchBranches       = errors.New("failed to fetch branches")
	ErrAuthMethod          = errors.New("failed to get auth method")
	ErrWorktree            = errors.New("failed to get worktree")
	ErrHeadSet             = errors.New("failed to set HEAD reference")
	ErrOpenRepository      = errors.New("failed to open repository")
	ErrOpenWorktree        = errors.New("failed to open repository worktree")
	ErrRemoteCreation      = errors.New("failed to set remote in target repository")
	ErrRepoInitialization  = errors.New("failed to initialize target repository")
	ErrPushRepository      = errors.New("failed to push to target repository")
	ErrUncleanWorkspace    = errors.New("workspace is unclean, aborting")
	ErrUnsupportedGitType  = errors.New("failed with unsupported git option type")
	ErrGitBinaryNotFound   = errors.New("failed to find a Git executable")
	ErrEmptyBinaryPath     = errors.New("failed to find Git binary path")
	ErrPermissionDenied    = errors.New("failed with permission denied (publickey). Provide correct key in your ssh-agent")
	ErrSetRepositoryConfig = errors.New("failed to set repository config")
	ErrGetRemoteBranches   = errors.New("failed to get remote branches")
	ErrTmpDirPath          = errors.New("failed to get tmpdirpath")
)
