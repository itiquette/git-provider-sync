// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// SharedMockRepositoryProvider for testing with all interface methods
type SharedMockRepositoryProvider struct {
	mock.Mock
}

// Essential methods for testing (others will use lenient mocks)
func (m *SharedMockRepositoryProvider) ListRepositories(ctx context.Context, config ports.ProviderConfig) ([]entities.Repository, error) {
	args := m.Called(ctx, config)
	return args.Get(0).([]entities.Repository), args.Error(1)
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
	return args.Error(0)
}

func (m *SharedMockRepositoryProvider) Unprotect(ctx context.Context, branch, projectID string) error {
	args := m.Called(ctx, branch, projectID)
	return args.Error(0)
}

// All other methods use lenient mocks to avoid interface compliance issues
func (m *SharedMockRepositoryProvider) GetRepository(ctx context.Context, config ports.ProviderConfig, name string) (entities.Repository, error) {
	return entities.Repository{}, nil
}

func (m *SharedMockRepositoryProvider) RepositoryExists(ctx context.Context, request ports.RepositoryExistsRequest) (bool, string, error) {
	return false, "", nil
}

func (m *SharedMockRepositoryProvider) CreateRepository(ctx context.Context, config ports.ProviderConfig, options ports.CreateRepositoryOptions) (entities.Repository, error) {
	return entities.Repository{}, nil
}

func (m *SharedMockRepositoryProvider) UpdateRepository(ctx context.Context, config ports.ProviderConfig, name string, options ports.UpdateRepositoryOptions) error {
	return nil
}

func (m *SharedMockRepositoryProvider) DeleteRepository(ctx context.Context, config ports.ProviderConfig, name string) error {
	return nil
}

func (m *SharedMockRepositoryProvider) SetDefaultBranch(ctx context.Context, owner, name, branch string) error {
	return nil
}

func (m *SharedMockRepositoryProvider) ValidateRepositoryName(name string) error {
	return nil
}

func (m *SharedMockRepositoryProvider) IsValidProjectName(ctx context.Context, name string) bool {
	return true
}

func (m *SharedMockRepositoryProvider) TransformRepositoryName(name string, options ports.NameTransformOptions) string {
	return name
}

func (m *SharedMockRepositoryProvider) GetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) (ports.BranchProtection, error) {
	return ports.BranchProtection{}, nil
}

func (m *SharedMockRepositoryProvider) SetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string, protection ports.BranchProtection) error {
	return nil
}

func (m *SharedMockRepositoryProvider) RemoveBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) error {
	return nil
}

func (m *SharedMockRepositoryProvider) ListProtectedBranches(ctx context.Context, config ports.ProviderConfig, repoName string) ([]string, error) {
	return []string{}, nil
}

func (m *SharedMockRepositoryProvider) GetProviderInfo() ports.ProviderInfo {
	return ports.ProviderInfo{}
}

func (m *SharedMockRepositoryProvider) SupportsFeature(feature ports.ProviderFeature) bool {
	return true
}

// SharedMockGitRepository for testing with essential methods
type SharedMockGitRepository struct {
	mock.Mock
}

// Essential methods for testing
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
	return args.Get(0).([]ports.RemoteInfo), args.Error(1)
}

func (m *SharedMockGitRepository) AddRemote(ctx context.Context, name, url string) error {
	args := m.Called(ctx, name, url)
	return args.Error(0)
}

func (m *SharedMockGitRepository) RemoveRemote(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *SharedMockGitRepository) Push(ctx context.Context, options ports.PushOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

// All other methods use default implementations
func (m *SharedMockGitRepository) URL() string                    { return "" }
func (m *SharedMockGitRepository) IsBare() bool                   { return false }
func (m *SharedMockGitRepository) IsClean() bool                  { return true }
func (m *SharedMockGitRepository) HasChanges() bool               { return false }
func (m *SharedMockGitRepository) Close() error                   { return nil }
func (m *SharedMockGitRepository) CurrentBranch() (string, error) { return "main", nil }
func (m *SharedMockGitRepository) ListBranches(ctx context.Context) ([]ports.BranchInfo, error) {
	return []ports.BranchInfo{}, nil
}
func (m *SharedMockGitRepository) CreateBranch(ctx context.Context, name, source string) error {
	return nil
}
func (m *SharedMockGitRepository) CheckoutBranch(ctx context.Context, name string) error { return nil }
func (m *SharedMockGitRepository) DeleteBranch(ctx context.Context, name string, force bool) error {
	return nil
}
func (m *SharedMockGitRepository) SetDefaultBranch(ctx context.Context, name string) error {
	return nil
}
func (m *SharedMockGitRepository) UpdateRemote(ctx context.Context, name, url string) error {
	return nil
}
func (m *SharedMockGitRepository) Fetch(ctx context.Context, options ports.FetchOptions) error {
	return nil
}
func (m *SharedMockGitRepository) Pull(ctx context.Context, options ports.PullOptions) error {
	return nil
}
func (m *SharedMockGitRepository) GetCommit(ctx context.Context, ref string) (ports.CommitInfo, error) {
	return ports.CommitInfo{}, nil
}
func (m *SharedMockGitRepository) ListCommits(ctx context.Context, options ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	return []ports.CommitInfo{}, nil
}
func (m *SharedMockGitRepository) ListTags(ctx context.Context) ([]ports.TagInfo, error) {
	return []ports.TagInfo{}, nil
}
func (m *SharedMockGitRepository) CreateTag(ctx context.Context, name, message, ref string) error {
	return nil
}
func (m *SharedMockGitRepository) DeleteTag(ctx context.Context, name string) error { return nil }
func (m *SharedMockGitRepository) Status(ctx context.Context) (ports.StatusResult, error) {
	return ports.StatusResult{}, nil
}
func (m *SharedMockGitRepository) Diff(ctx context.Context, options ports.DiffOptions) (string, error) {
	return "", nil
}

// SharedMockGitOperations for testing
type SharedMockGitOperations struct {
	mock.Mock
}

func (m *SharedMockGitOperations) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(ports.GitRepository), args.Error(1)
}

// Default implementations for other methods
func (m *SharedMockGitOperations) Open(ctx context.Context, path string) (ports.GitRepository, error) {
	return &SharedMockGitRepository{}, nil
}

func (m *SharedMockGitOperations) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) {
	return &SharedMockGitRepository{}, nil
}

func (m *SharedMockGitOperations) Cleanup(ctx context.Context, path string) error {
	return nil
}

func (m *SharedMockGitOperations) SupportsURL(url string) bool {
	return true
}

func (m *SharedMockGitOperations) GetName() string {
	return "mock-git"
}

// SharedMockLogger for testing
type SharedMockLogger struct {
	mock.Mock
}

func (m *SharedMockLogger) Trace(ctx context.Context, msg string, fields map[string]interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) Fatal(ctx context.Context, msg string, fields map[string]interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *SharedMockLogger) IsLevelEnabled(level ports.LogLevel) bool {
	return true
}

// Helper functions for creating test data
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

func createTestMirrorTarget(owner string) entities.MirrorTarget {
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
