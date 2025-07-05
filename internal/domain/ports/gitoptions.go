// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"time"
)

// CloneOptions contains options for cloning repositories.
type CloneOptions struct {
	URL          string
	Path         string
	Branch       string
	SingleBranch bool
	Depth        int
	Mirror       bool
	Bare         bool
	Auth         AuthOptions
	Progress     ProgressCallback
	Tags         TagMode
	Timeout      time.Duration
}

// NewCloneOptions creates default clone options.
func NewCloneOptions() CloneOptions {
	return CloneOptions{
		SingleBranch: false,
		Depth:        0,
		Mirror:       false,
		Bare:         false,
		Tags:         TagModeDefault,
		Timeout:      time.Minute * 5,
	}
}

// InitOptions contains options for initializing repositories.
type InitOptions struct {
	Bare          bool
	Template      string
	DefaultBranch string
}

// NewInitOptions creates default init options.
func NewInitOptions() InitOptions {
	return InitOptions{
		Bare:          false,
		DefaultBranch: "main",
	}
}

// FetchOptions contains options for fetching from remotes.
type FetchOptions struct {
	Remote    string
	RefSpecs  []string
	Depth     int
	Auth      AuthOptions
	Progress  ProgressCallback
	Prune     bool
	PruneTags bool
	Force     bool
	Timeout   time.Duration
}

// NewFetchOptions creates default fetch options.
func NewFetchOptions() FetchOptions {
	return FetchOptions{
		Remote:    "origin",
		Depth:     0,
		Prune:     false,
		PruneTags: false,
		Force:     false,
		Timeout:   time.Minute * 5,
	}
}

// PullOptions contains options for pulling from remotes.
type PullOptions struct {
	Remote      string
	Branch      string
	Auth        AuthOptions
	Progress    ProgressCallback
	Rebase      bool
	FastForward FastForwardMode
	Strategy    MergeStrategy
	Timeout     time.Duration
}

// NewPullOptions creates default pull options.
func NewPullOptions() PullOptions {
	return PullOptions{
		Remote:      "origin",
		Rebase:      false,
		FastForward: FastForwardDefault,
		Strategy:    MergeStrategyDefault,
		Timeout:     time.Minute * 5,
	}
}

// PushOptions contains options for pushing to remotes.
type PushOptions struct {
	Remote     string
	RefSpecs   []string
	Auth       AuthOptions
	Progress   ProgressCallback
	Force      bool
	Atomic     bool
	Tags       bool
	FollowTags bool
	Timeout    time.Duration
}

// NewPushOptions creates default push options.
func NewPushOptions() PushOptions {
	return PushOptions{
		Remote:     "origin",
		Force:      false,
		Atomic:     false,
		Tags:       false,
		FollowTags: false,
		Timeout:    time.Minute * 5,
	}
}

// AuthOptions contains authentication options for git operations.
type AuthOptions struct {
	Type       AuthType
	Username   string
	Password   string
	Token      string
	SSHKeyPath string
	SSHKey     []byte
	Passphrase string
}

// NewAuthOptions creates default auth options.
func NewAuthOptions() AuthOptions {
	return AuthOptions{
		Type: AuthTypeNone,
	}
}

// ListCommitsOptions contains options for listing commits.
type ListCommitsOptions struct {
	From      string
	To        string
	MaxCount  int
	Since     *time.Time
	Until     *time.Time
	Author    string
	Committer string
	Message   string
	Paths     []string
}

// NewListCommitsOptions creates default list commits options.
func NewListCommitsOptions() ListCommitsOptions {
	return ListCommitsOptions{
		MaxCount: 100,
	}
}

// DiffOptions contains options for generating diffs.
type DiffOptions struct {
	From        string
	To          string
	Paths       []string
	Cached      bool
	NameOnly    bool
	NameStatus  bool
	Context     int
	IgnoreSpace bool
}

// NewDiffOptions creates default diff options.
func NewDiffOptions() DiffOptions {
	return DiffOptions{
		Context:     3,
		IgnoreSpace: false,
	}
}

// AuthType defines authentication methods.
type AuthType int

const (
	AuthTypeNone AuthType = iota
	AuthTypeBasic
	AuthTypeToken
	AuthTypeSSH
	AuthTypeSSHAgent
	AuthTypeSSHKey
)

// ProgressCallback is called during long-running operations to report progress.
type ProgressCallback func(current, total int64, message string)

// GitConfig represents git configuration settings.
type GitConfig struct {
	UserName                string
	UserEmail               string
	SignKey                 string
	SignOff                 bool
	PreferredImplementation string
	MaxConcurrent           int
	VerifySSL               bool
	Debug                   bool
	CacheSize               int64
	Timeout                 time.Duration
	TrustDomains            []string
	LogFile                 string
}
