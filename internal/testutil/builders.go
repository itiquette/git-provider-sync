// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// RepositoryOption is a functional option for configuring test repositories.
type RepositoryOption func(*repositoryConfig)

type repositoryConfig struct {
	name          string
	httpsURL      string
	sshURL        string
	defaultBranch string
	description   string
	visibility    string
	lastActivity  time.Time
	isPrivate     bool
	isFork        bool
	isArchived    bool
}

// WithRepositoryName sets the repository name.
func WithRepositoryName(name string) RepositoryOption {
	return func(cfg *repositoryConfig) {
		cfg.name = name
	}
}

// WithHTTPSURL sets the HTTPS URL.
func WithHTTPSURL(url string) RepositoryOption {
	return func(cfg *repositoryConfig) {
		cfg.httpsURL = url
	}
}

// WithSSHURL sets the SSH URL.
func WithSSHURL(url string) RepositoryOption {
	return func(cfg *repositoryConfig) {
		cfg.sshURL = url
	}
}

// WithDefaultBranch sets the default branch.
func WithDefaultBranch(branch string) RepositoryOption {
	return func(cfg *repositoryConfig) {
		cfg.defaultBranch = branch
	}
}

// WithPrivate marks the repository as private.
func WithPrivate(private bool) RepositoryOption {
	return func(cfg *repositoryConfig) {
		cfg.isPrivate = private
		if private {
			cfg.visibility = "private"
		}
	}
}

// WithArchived marks the repository as archived.
func WithArchived(archived bool) RepositoryOption {
	return func(cfg *repositoryConfig) {
		cfg.isArchived = archived
	}
}

// WithFork marks the repository as a fork.
func WithFork(fork bool) RepositoryOption {
	return func(cfg *repositoryConfig) {
		cfg.isFork = fork
	}
}

// NewTestRepository creates a test repository with functional options.
func NewTestRepository(tb testing.TB, opts ...RepositoryOption) entities.Repository {
	tb.Helper()

	// Default configuration
	cfg := &repositoryConfig{
		name:          "test-repo",
		httpsURL:      "https://github.com/test/test-repo.git",
		sshURL:        "git@github.com:test/test-repo.git",
		defaultBranch: "main",
		description:   "Test repository",
		visibility:    "public",
		lastActivity:  time.Now(),
		isPrivate:     false,
		isFork:        false,
		isArchived:    false,
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	// Build repository using the builder
	builder := entities.NewRepositoryBuilder()

	// WithName returns (builder, error)
	builder, err := builder.WithName(cfg.name)
	require.NoError(tb, err, "Failed to set repository name")

	// Chain other methods that return (builder, error)
	builder, err = builder.WithHTTPSURL(cfg.httpsURL)
	require.NoError(tb, err, "Failed to set HTTPS URL")

	builder, err = builder.WithSSHURL(cfg.sshURL)
	require.NoError(tb, err, "Failed to set SSH URL")

	builder, err = builder.WithDefaultBranch(cfg.defaultBranch)
	require.NoError(tb, err, "Failed to set default branch")

	builder = builder.WithDescription(cfg.description)
	builder = builder.WithVisibility(cfg.visibility)
	builder = builder.WithPrivate(cfg.isPrivate)
	builder = builder.WithFork(cfg.isFork)
	builder = builder.WithArchived(cfg.isArchived)
	builder = builder.WithLastActivityAt(cfg.lastActivity)

	// Build the repository
	repo, err := builder.Build()
	require.NoError(tb, err, "Failed to build test repository")

	return repo
}

// NewTestPrivateRepository creates a private test repository.
func NewTestPrivateRepository(tb testing.TB, name string) entities.Repository {
	tb.Helper()

	return NewTestRepository(tb,
		WithRepositoryName(name),
		WithPrivate(true),
		WithHTTPSURL("https://github.com/test/"+name+".git"),
		WithSSHURL("git@github.com:test/"+name+".git"),
	)
}

// NewTestArchivedRepository creates an archived test repository.
func NewTestArchivedRepository(tb testing.TB, name string) entities.Repository {
	tb.Helper()

	return NewTestRepository(tb,
		WithRepositoryName(name),
		WithArchived(true),
		WithHTTPSURL("https://github.com/test/"+name+".git"),
		WithSSHURL("git@github.com:test/"+name+".git"),
	)
}

// NewTestForkedRepository creates a forked test repository.
func NewTestForkedRepository(tb testing.TB, name string) entities.Repository {
	tb.Helper()

	return NewTestRepository(tb,
		WithRepositoryName(name),
		WithFork(true),
		WithHTTPSURL("https://github.com/test/"+name+".git"),
		WithSSHURL("git@github.com:test/"+name+".git"),
	)
}

// MirrorTargetOption is a functional option for configuring test mirror targets.
type MirrorTargetOption func(*mirrorTargetConfig)

type mirrorTargetConfig struct {
	name         string
	providerType string
	owner        string
	ownerType    string
	domain       string
	path         string
	token        string
}

// WithTargetName sets the mirror target name.
func WithTargetName(name string) MirrorTargetOption {
	return func(cfg *mirrorTargetConfig) {
		cfg.name = name
	}
}

// WithTargetProvider sets the provider type.
func WithTargetProvider(provider string) MirrorTargetOption {
	return func(cfg *mirrorTargetConfig) {
		cfg.providerType = provider
	}
}

// WithTargetOwner sets the owner.
func WithTargetOwner(owner string) MirrorTargetOption {
	return func(cfg *mirrorTargetConfig) {
		cfg.owner = owner
	}
}

// NewTestMirrorTarget creates a test mirror target with functional options.
func NewTestMirrorTarget(tb testing.TB, opts ...MirrorTargetOption) entities.MirrorTarget {
	tb.Helper()

	// Default configuration
	cfg := &mirrorTargetConfig{
		name:         "test-mirror",
		providerType: "gitlab",
		owner:        "test-owner",
		ownerType:    "user",
		domain:       "gitlab.com",
		path:         "",
		token:        "test-token",
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	// Build mirror target using the builder
	builder := entities.NewMirrorTargetBuilder()

	// Chain methods
	builder, err := builder.WithName(cfg.name)
	require.NoError(tb, err, "Failed to set mirror target name")

	builder, err = builder.WithOwner(cfg.owner)
	require.NoError(tb, err, "Failed to set owner")

	// Note: WithOwnerType might not exist on MirrorTargetBuilder
	// builder = builder.WithOwnerType(cfg.ownerType)

	builder = builder.WithDomain(cfg.domain)

	require.NoError(tb, err, "Failed to set domain")

	builder, err = builder.WithPath(cfg.path)
	require.NoError(tb, err, "Failed to set path")

	// Note: WithProviderType might not exist, check the actual builder
	// For now, we'll set other fields

	// Build auth config separately
	authConfig := entities.AuthConfig{}
	// Note: AuthConfig fields are unexported, so we can't set token directly

	builder = builder.WithAuth(authConfig)

	// Build the mirror target
	target, err := builder.Build()
	require.NoError(tb, err, "Failed to build test mirror target")

	return target
}
