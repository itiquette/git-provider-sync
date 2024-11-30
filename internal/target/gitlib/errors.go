// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlib

import "errors"

var (
	ErrAuthMethod       = errors.New("failed to get auth method")
	ErrBranchCheckout   = errors.New("failed to checkout branch")
	ErrCloneRepository  = errors.New("failed to clone repository")
	ErrFetchBranches    = errors.New("failed to fetch branches")
	ErrWorktree         = errors.New("failed to get worktree")
	ErrHeadSet          = errors.New("failed to set HEAD reference")
	ErrInvalidAuth      = errors.New("invalid authentication configuration")
	ErrOpenRepository   = errors.New("failed to open repository")
	ErrUncleanWorkspace = errors.New("workspace is unclean, aborting")
	ErrPullRepository   = errors.New("failed to pull repository")
	ErrPushRepository   = errors.New("failed to push repository")
	ErrRemoteCreation   = errors.New("failed to set remote in target repository")
	ErrRepositoryOpen   = errors.New("failed to open repository")
	ErrWorktreeStatus   = errors.New("failed to get worktree status")
)
