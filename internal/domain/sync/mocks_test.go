// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/mock"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// SharedMockRepositoryProvider for testing with all interface methods.
type SharedMockRepositoryProvider struct {
	mock.Mock
}

func (m *SharedMockRepositoryProvider) ListRepositories(ctx context.Context, config ports.ProviderConfig) ([]entities.Repository, error) {
	args := m.Called(ctx, config)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return args.Get(0).([]entities.Repository), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *SharedMockRepositoryProvider) ProjectExists(ctx context.Context, owner, name string) (bool, string, error) {
	args := m.Called(ctx, owner, name)

	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *SharedMockRepositoryProvider) CreateRepositoryForPush(ctx context.Context, request ports.CreateRepositoryRequest) (string, error) {
	args := m.Called(ctx, request)

	return args.String(0), args.Error(1)
}

func (m *SharedMockRepositoryProvider) Protect(ctx context.Context, owner, branch, projectID string) error {
	args := m.Called(ctx, owner, branch, projectID)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to protect branch: %w", err)
	}

	return nil
}

func (m *SharedMockRepositoryProvider) Unprotect(ctx context.Context, branch, projectID string) error {
	args := m.Called(ctx, branch, projectID)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to unprotect branch: %w", err)
	}

	return nil
}

func (m *SharedMockRepositoryProvider) GetRepository(_ context.Context, _ ports.ProviderConfig, _ string) (entities.Repository, error) {
	return entities.Repository{}, nil
}

func (m *SharedMockRepositoryProvider) RepositoryExists(_ context.Context, _ ports.RepositoryExistsRequest) (bool, string, error) {
	return false, "", nil
}

func (m *SharedMockRepositoryProvider) CreateRepository(_ context.Context, _ ports.ProviderConfig, _ ports.CreateRepositoryOptions) (entities.Repository, error) {
	return entities.Repository{}, nil
}

func (m *SharedMockRepositoryProvider) UpdateRepository(_ context.Context, _ ports.ProviderConfig, _ string, _ ports.UpdateRepositoryOptions) error {
	return nil
}

func (m *SharedMockRepositoryProvider) DeleteRepository(_ context.Context, _ ports.ProviderConfig, _ string) error {
	return nil
}

func (m *SharedMockRepositoryProvider) SetDefaultBranch(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *SharedMockRepositoryProvider) ValidateRepositoryName(_ string) error {
	return nil
}

func (m *SharedMockRepositoryProvider) IsValidProjectName(_ context.Context, _ string) bool {
	return true
}

func (m *SharedMockRepositoryProvider) TransformRepositoryName(name string, _ ports.NameTransformOptions) string {
	return name
}

func (m *SharedMockRepositoryProvider) GetBranchProtection(_ context.Context, _ ports.ProviderConfig, _, _ string) (ports.BranchProtection, error) {
	return ports.BranchProtection{}, nil
}

func (m *SharedMockRepositoryProvider) SetBranchProtection(_ context.Context, _ ports.ProviderConfig, _, _ string, _ ports.BranchProtection) error {
	return nil
}

func (m *SharedMockRepositoryProvider) RemoveBranchProtection(_ context.Context, _ ports.ProviderConfig, _, _ string) error {
	return nil
}

func (m *SharedMockRepositoryProvider) ListProtectedBranches(_ context.Context, _ ports.ProviderConfig, _ string) ([]string, error) {
	return []string{}, nil
}

func (m *SharedMockRepositoryProvider) GetProviderInfo() ports.ProviderInfo {
	return ports.ProviderInfo{}
}

func (m *SharedMockRepositoryProvider) SupportsFeature(_ ports.ProviderFeature) bool {
	return true
}

// SharedMockGitRepository for testing.
type SharedMockGitRepository struct {
	mock.Mock
}

func (m *SharedMockGitRepository) Name() string {
	args := m.Called()

	return args.String(0)
}

func (m *SharedMockGitRepository) Path() string {
	args := m.Called()

	return args.String(0)
}

func (m *SharedMockGitRepository) ListRemotes(ctx context.Context) ([]ports.RemoteInfo, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	return args.Get(0).([]ports.RemoteInfo), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *SharedMockGitRepository) AddRemote(ctx context.Context, name, url string) error {
	args := m.Called(ctx, name, url)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

func (m *SharedMockGitRepository) RemoveRemote(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to remove remote: %w", err)
	}

	return nil
}

func (m *SharedMockGitRepository) Push(ctx context.Context, options ports.PushOptions) error {
	args := m.Called(ctx, options)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

func (m *SharedMockGitRepository) URL() string                    { return "" }
func (m *SharedMockGitRepository) IsBare() bool                   { return false }
func (m *SharedMockGitRepository) IsClean() bool                  { return true }
func (m *SharedMockGitRepository) HasChanges() bool               { return false }
func (m *SharedMockGitRepository) Close() error                   { return nil }
func (m *SharedMockGitRepository) CurrentBranch() (string, error) { return "main", nil }
func (m *SharedMockGitRepository) ListBranches(_ context.Context) ([]ports.BranchInfo, error) {
	return []ports.BranchInfo{}, nil
}
func (m *SharedMockGitRepository) CreateBranch(_ context.Context, _, _ string) error {
	return nil
}
func (m *SharedMockGitRepository) CheckoutBranch(_ context.Context, _ string) error { return nil }
func (m *SharedMockGitRepository) DeleteBranch(_ context.Context, _ string, _ bool) error {
	return nil
}
func (m *SharedMockGitRepository) SetDefaultBranch(_ context.Context, _ string) error {
	return nil
}
func (m *SharedMockGitRepository) UpdateRemote(ctx context.Context, name, url string) error {
	args := m.Called(ctx, name, url)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to update remote: %w", err)
	}

	return nil
}
func (m *SharedMockGitRepository) Fetch(_ context.Context, _ ports.FetchOptions) error {
	return nil
}
func (m *SharedMockGitRepository) Pull(_ context.Context, _ ports.PullOptions) error {
	return nil
}
func (m *SharedMockGitRepository) GetCommit(_ context.Context, _ string) (ports.CommitInfo, error) {
	return ports.CommitInfo{}, nil
}
func (m *SharedMockGitRepository) ListCommits(_ context.Context, _ ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	return []ports.CommitInfo{}, nil
}
func (m *SharedMockGitRepository) ListTags(_ context.Context) ([]ports.TagInfo, error) {
	return []ports.TagInfo{}, nil
}
func (m *SharedMockGitRepository) CreateTag(_ context.Context, _, _, _ string) error {
	return nil
}
func (m *SharedMockGitRepository) DeleteTag(_ context.Context, _ string) error { return nil }
func (m *SharedMockGitRepository) Status(_ context.Context) (ports.StatusResult, error) {
	return ports.StatusResult{}, nil
}
func (m *SharedMockGitRepository) Diff(_ context.Context, _ ports.DiffOptions) (string, error) {
	return "", nil
}

// SharedMockGitOperations for testing.
type SharedMockGitOperations struct {
	mock.Mock
}

//nolint:ireturn // Test mock returns interface
func (m *SharedMockGitOperations) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) {
	args := m.Called(ctx, options)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return args.Get(0).(ports.GitRepository), nil //nolint:forcetypeassert // Test mock - controlled return values
}

//nolint:ireturn // Test mock returns interface
func (m *SharedMockGitOperations) Open(_ context.Context, _ string) (ports.GitRepository, error) {
	return &SharedMockGitRepository{}, nil
}

//nolint:ireturn // Test mock returns interface
func (m *SharedMockGitOperations) Init(_ context.Context, _ string, _ ports.InitOptions) (ports.GitRepository, error) {
	return &SharedMockGitRepository{}, nil
}

func (m *SharedMockGitOperations) Cleanup(_ context.Context, _ string) error {
	return nil
}

func (m *SharedMockGitOperations) SupportsURL(_ string) bool {
	return true
}

func (m *SharedMockGitOperations) GetName() string {
	return "mock-git"
}

func (m *SharedMockGitOperations) CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	args := m.Called(ctx, dir, prefix)
	if err := args.Error(1); err != nil {
		return ctx, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return args.Get(0).(context.Context), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *SharedMockGitOperations) GetTmpDirPath(ctx context.Context) (string, error) {
	args := m.Called(ctx)

	return args.String(0), args.Error(1)
}

func (m *SharedMockGitOperations) DeleteTmpDir(ctx context.Context) error {
	args := m.Called(ctx)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to delete temp directory: %w", err)
	}

	return nil
}

// SharedMockLogger for testing.
type SharedMockLogger struct {
	mock.Mock
}

func (m *SharedMockLogger) Trace(ctx context.Context, msg string, fields map[string]any) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Debug(ctx context.Context, msg string, fields map[string]any) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Info(ctx context.Context, msg string, fields map[string]any) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Warn(ctx context.Context, msg string, fields map[string]any) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Error(ctx context.Context, msg string, fields map[string]any) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Fatal(ctx context.Context, msg string, fields map[string]any) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) IsLevelEnabled(_ ports.LogLevel) bool {
	return true
}

func createTestRepository(name string) entities.Repository {
	builder := entities.NewRepositoryBuilder()
	builder, _ = builder.WithName(name)
	builder, _ = builder.WithHTTPSURL("https://github.com/source/" + name + ".git")
	builder, _ = builder.WithSSHURL("git@github.com:source/" + name + ".git")
	builder, _ = builder.WithDefaultBranch("main")
	builder = builder.WithDescription("Test repository for " + name)
	builder = builder.WithVisibility("public")
	repo, _ := builder.Build()

	return repo
}

func createTestRepositoryWithActivity(name string, isFork, isArchived bool, lastActivity time.Time) entities.Repository {
	builder := entities.NewRepositoryBuilder()
	builder, _ = builder.WithName(name)
	builder, _ = builder.WithHTTPSURL("https://github.com/test/" + name + ".git")
	builder, _ = builder.WithSSHURL("git@github.com:test/" + name + ".git")
	builder, _ = builder.WithDefaultBranch("main")
	builder = builder.WithDescription("Test repository for " + name)

	visibility := "public"
	if name == "private-repo" {
		visibility = "private"
	}

	builder = builder.WithVisibility(visibility)

	if isFork {
		builder = builder.WithFork(true)
	}

	if isArchived {
		builder = builder.WithArchived(true)
	}

	// Set last activity using the builder method that exists
	builder = builder.WithLastActivityAt(lastActivity)

	repo, _ := builder.Build()

	return repo
}

func createTestMirrorTarget(owner string) entities.MirrorTarget { //nolint:unparam // Used with different values in different test files
	authConfig := entities.NewAuthConfigWithToken("test-token", "git")
	builder := entities.NewMirrorTargetBuilder()
	builder, _ = builder.WithName("test-target")
	builder, _ = builder.WithProvider("github")
	builder = builder.WithDomain("github.com")
	builder, _ = builder.WithOwner(owner)
	builder, _ = builder.WithPath("")
	builder = builder.WithAuth(authConfig)
	target, _ := builder.Build()

	return target
}

// SharedMockMirrorOperations for testing.
type SharedMockMirrorOperations struct {
	mock.Mock
}

func (m *SharedMockMirrorOperations) Mirror(ctx context.Context, opts ports.MirrorOptions) error {
	args := m.Called(ctx, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to mirror: %w", err)
	}

	return nil
}

func createTestProviderConfig() ports.ProviderConfig {
	return ports.ProviderConfig{
		ProviderType: "github",
		Domain:       "github.com",
		Owner:        "test",
		AuthConfig: ports.AuthenticationConfig{
			Token:    "test-token",
			Username: "git",
		},
	}
}
