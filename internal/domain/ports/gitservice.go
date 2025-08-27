// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"context"
)

// GitOperations defines the interface for git operations (secondary port).
// This port is implemented by adapters that handle git operations like go-git, git binary, etc.
type GitOperations interface {
	// Repository operations
	Clone(ctx context.Context, options CloneOptions) (GitRepository, error)
	Open(ctx context.Context, path string) (GitRepository, error)
	Init(ctx context.Context, path string, options InitOptions) (GitRepository, error)

	// Cleanup operations
	Cleanup(ctx context.Context, path string) error

	// Temporary directory operations
	CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error)
	GetTmpDirPath(ctx context.Context) (string, error)
	DeleteTmpDir(ctx context.Context) error

	// Factory methods
	SupportsURL(url string) bool
	GetName() string
}
