// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"context"
	"time"
)

// GitRepositoryInfo provides basic repository information (core interface).
type GitRepositoryInfo interface {
	// Basic information
	Path() string
	URL() string
	Name() string

	// State queries
	IsBare() bool
	IsClean() bool
	HasChanges() bool

	// Cleanup
	Close() error
}

// GitBranchOperations provides branch management capabilities.
type GitBranchOperations interface {
	CurrentBranch() (string, error)
	ListBranches(ctx context.Context) ([]BranchInfo, error)
	CreateBranch(ctx context.Context, name, source string) error
	CheckoutBranch(ctx context.Context, name string) error
	DeleteBranch(ctx context.Context, name string, force bool) error
	SetDefaultBranch(ctx context.Context, name string) error
}

// GitRemoteOperations provides remote management capabilities.
type GitRemoteOperations interface {
	ListRemotes(ctx context.Context) ([]RemoteInfo, error)
	AddRemote(ctx context.Context, name, url string) error
	RemoveRemote(ctx context.Context, name string) error
	UpdateRemote(ctx context.Context, name, url string) error
}

// GitSyncOperations provides sync capabilities.
type GitSyncOperations interface {
	Fetch(ctx context.Context, options FetchOptions) error
	Pull(ctx context.Context, options PullOptions) error
	Push(ctx context.Context, options PushOptions) error
}

// GitCommitOperations provides commit operations.
type GitCommitOperations interface {
	GetCommit(ctx context.Context, ref string) (CommitInfo, error)
	ListCommits(ctx context.Context, options ListCommitsOptions) ([]CommitInfo, error)
}

// GitTagOperations provides tag management capabilities.
type GitTagOperations interface {
	ListTags(ctx context.Context) ([]TagInfo, error)
	CreateTag(ctx context.Context, name, message, ref string) error
	DeleteTag(ctx context.Context, name string) error
}

// GitStatusOperations provides status and diff capabilities.
type GitStatusOperations interface {
	Status(ctx context.Context) (StatusResult, error)
	Diff(ctx context.Context, options DiffOptions) (string, error)
}

// GitRepository represents a git repository that can be operated on.
// This interface composes smaller, focused interfaces following ISP.
type GitRepository interface {
	GitRepositoryInfo
	GitBranchOperations
	GitRemoteOperations
	GitSyncOperations
	GitCommitOperations
	GitTagOperations
	GitStatusOperations
}

// Supporting types

// BranchInfo contains information about a branch.
type BranchInfo struct {
	Name      string
	Hash      string
	Upstream  string
	IsRemote  bool
	IsCurrent bool
	Commit    string
}

// RemoteInfo contains information about a remote.
type RemoteInfo struct {
	Name      string
	URL       string
	FetchURL  string
	PushURL   string
	IsDefault bool
}

// CommitInfo contains information about a commit.
type CommitInfo struct {
	Hash      string
	Author    PersonInfo
	Committer PersonInfo
	Message   string
	Timestamp time.Time
	Parents   []string
}

// PersonInfo contains information about a person (author/committer).
type PersonInfo struct {
	Name      string
	Email     string
	When      time.Time
	Timestamp time.Time
}

// TagInfo contains information about a tag.
type TagInfo struct {
	Name      string
	Hash      string
	Commit    string
	Message   string
	Tagger    PersonInfo
	Timestamp time.Time
	IsSigned  bool
}

// StatusResult contains git status information.
type StatusResult struct {
	IsClean    bool
	Modified   []string
	Added      []string
	Deleted    []string
	Renamed    []string
	Untracked  []string
	Conflicted []string
}

// Enums

// FastForwardMode defines fast-forward merge modes.
type FastForwardMode int

const (
	// FastForwardDefault represents the default fast-forward mode.
	FastForwardDefault FastForwardMode = iota
	// FastForwardOnly represents fast-forward only mode.
	FastForwardOnly
	// FastForwardModeOnly represents fast-forward mode only.
	FastForwardModeOnly
	// FastForwardNever represents never fast-forward mode.
	FastForwardNever
)

// MergeStrategy defines merge strategies.
type MergeStrategy int

const (
	// MergeStrategyDefault represents the default merge strategy.
	MergeStrategyDefault MergeStrategy = iota
	// MergeStrategyRecursive represents recursive merge strategy.
	MergeStrategyRecursive
	// MergeStrategyOctopus represents octopus merge strategy.
	MergeStrategyOctopus
	// MergeStrategyOurs represents ours merge strategy.
	MergeStrategyOurs
	// MergeStrategySubtree represents subtree merge strategy.
	MergeStrategySubtree
)

// TagMode defines tag fetching modes.
type TagMode int

const (
	// TagModeDefault represents the default tag mode.
	TagModeDefault TagMode = iota
	// TagModeAll represents all tags mode.
	TagModeAll
	// TagModeNone represents no tags mode.
	TagModeNone
	// TagModeFollowing represents following tags mode.
	TagModeFollowing
	// TagModeFollow represents follow tags mode.
	TagModeFollow
)
