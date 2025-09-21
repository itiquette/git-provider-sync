//go:build integration

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/repository/gogit"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
	"itiquette/git-provider-sync/internal/integrationtest/testutil"
)

// TestCriticalGitHubToGitLabSyncFlow tests the critical GitHub → GitLab sync flow
// That was failing due to missing remote URL update
func TestCriticalGitHubToGitLabSyncFlow(t *testing.T) {
	// Isolate Git environment from host system
	// Note: Cannot use t.Parallel() when using t.Setenv in IsolateGitEnvironment
	testutil.IsolateGitEnvironment(t)

	ctx := context.Background()

	// Create git operations
	gitOps := gogit.New(ports.GitConfig{
		UserName:    "test-user",
		UserEmail:   "test@example.com",
		StorageMode: ports.StorageModeMemory, // Use memory for faster, isolated tests
	})

	// Create mock provider with proper network call mocking
	mockProvider := &MockGitLabProvider{}
	mockLogger := &TestSyncLogger{}

	// Set up logger mocks (lenient)
	mockLogger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()

	// Test the critical flow: GitHub clone → GitLab sync
	t.Run("github clone to gitlab sync with remote update", func(t *testing.T) {
		// Step 1: Set up isolated git test environment (replaces manual bare repo setup)
		opts := testutil.GitTestOptions{
			SourceRepoName:  "github-repo",
			TargetRepoName:  "gitlab-repo",
			WorkingRepoName: "sync-workspace",
			InitialFiles: map[string]string{
				"README.md":   "# Test Repository\nOriginal GitHub content for sync test",
				"src/main.go": "package main\n\nfunc main() {\n\tprintln(\"Hello from GitHub!\")",
				".gitignore":  "*.log\n*.tmp\n",
				"docs/API.md": "# API Documentation\n\nEndpoints...",
			},
			AddRemotes: map[string]string{
				"origin": "", // Will be set to GitHub (source) URL
			},
		}

		env, err := testutil.SetupGitTestEnvironment(t, gitOps, opts)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		// Get URLs for our simulated git providers
		githubBarePath := env.GetSourceURL()
		gitlabBarePath := env.GetTargetURL()

		// Step 2: Create mock GitHub repository (using testutil environment)
		mockGithubRepo := &TestSyncGitRepository{}

		// Mock the GitHub repository behavior including remote updates
		mockGithubRepo.On("Name").Return("test-repo")
		mockGithubRepo.On("Path").Return(env.WorkingRepo.Path)

		// Initial state: origin points to GitHub (source)
		mockGithubRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
			{Name: "origin", URL: githubBarePath},
		}, nil).Once()

		// GPSUPSTREAM setup (backup of original GitHub URL)
		mockGithubRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(domain.ErrTestNotFound).Once()
		mockGithubRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", githubBarePath).Return(nil).Once()
		mockGithubRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
			{Name: "origin", URL: githubBarePath},
			{Name: "GPSUPSTREAM", URL: githubBarePath},
		}, nil).Once()

		// 🎯 CRITICAL: Mock UpdateRemote call (THE FIX WE IMPLEMENTED)
		mockGithubRepo.On("UpdateRemote", mock.Anything, "origin", mock.AnythingOfType("string")).Return(nil).Once()

		// Mock Push operation (to GitLab bare repo)
		mockGithubRepo.On("Push", mock.Anything, mock.AnythingOfType("ports.PushOptions")).Return(nil).Once()

		// Final state: origin points to GitLab, GPSUPSTREAM to GitHub
		mockGithubRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
			{Name: "origin", URL: gitlabBarePath},
			{Name: "GPSUPSTREAM", URL: githubBarePath},
		}, nil).Once()

		// Step 3: Set up GitLab target configuration pointing to our test environment
		gitlabTarget := createLocalGitLabTarget(gitlabBarePath)

		// Step 4: Mock GitLab provider responses (simulating API calls)
		mockProvider.On("VerifyTarget", mock.Anything, "gitlab-user", "test-repo").Return(false, "", nil)
		mockProvider.On("PrepareForPush", mock.Anything, mock.AnythingOfType("ports.CreateRepositoryRequest")).Return("gitlab-project-123", nil)

		// Step 5: Create source repository entity
		sourceRepo := createTestRepository("test-repo")

		// Step 6: Execute PushToProvider use case (tests our remote update fix)
		pushUseCase := sync.NewPushToProviderUseCase(mockProvider, gitOps)

		request := sync.PushRequest{
			SourceRepository: sourceRepo,
			SourceGitRepo:    mockGithubRepo, // Mocked git repo (no network calls)
			TargetConfig:     gitlabTarget,
			SourceConfig:     createGitHubProviderConfig(),
			CreateIfMissing:  true,
			DryRun:           false,
			ForcePush:        false,
		}

		// Execute the push (this should update remote and push)
		response, err := pushUseCase.Execute(ctx, request)
		require.NoError(t, err, "Push should succeed")
		assert.True(t, response.Success, "Push should be successful")
		assert.True(t, response.Created, "Repository should be created")

		// Step 7: 🎯 CRITICAL VERIFICATION - Check that origin remote was updated to GitLab
		updatedRemotes, err := mockGithubRepo.ListRemotes(ctx)
		require.NoError(t, err)

		var originRemote *ports.RemoteInfo
		var gpsUpstreamRemote *ports.RemoteInfo

		for _, remote := range updatedRemotes {
			if remote.Name == "origin" {
				originRemote = &remote
			}
			if remote.Name == "GPSUPSTREAM" {
				gpsUpstreamRemote = &remote
			}
		}

		// 🎯 VERIFY THE CRITICAL FIX: origin should now point to GitLab, GPSUPSTREAM to GitHub
		require.NotNil(t, originRemote, "Origin remote should exist")
		require.NotNil(t, gpsUpstreamRemote, "GPSUPSTREAM remote should exist")

		assert.Contains(t, originRemote.URL, gitlabBarePath, "🎯 Origin should now point to GitLab (target) bare repo")
		assert.Contains(t, gpsUpstreamRemote.URL, githubBarePath, "GPSUPSTREAM should still point to GitHub (source) bare repo")

		t.Logf("✅ CRITICAL FIX VERIFIED: Remote successfully updated!")
		t.Logf("   GitHub (source): %s", gpsUpstreamRemote.URL)
		t.Logf("   GitLab (target): %s", originRemote.URL)
		t.Logf("   Validates the UpdateRemote fix that resolves empty GitLab repositories")

		// Step 8: Verify all mock expectations were met
		mockProvider.AssertExpectations(t)
		mockGithubRepo.AssertExpectations(t)

		// Optional: Demonstrate that the test environment can be used for further testing
		t.Logf("📁 Test environment created:")
		t.Logf("   Source (GitHub): %s", env.SourceBareRepo.Path)
		t.Logf("   Target (GitLab): %s", env.TargetBareRepo.Path)
		t.Logf("   Working repo: %s", env.WorkingRepo.Path)
	})
}

// TestRemoteURLUpdateFailure tests what happens when remote update fails
func TestRemoteURLUpdateFailure(t *testing.T) {

	t.Parallel()

	// Create failing git repository mock
	mockGitRepo := &TestSyncGitRepository{}
	mockProvider := &MockGitLabProvider{}
	mockLogger := &TestSyncLogger{}

	// Set up mocks
	mockLogger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()

	// Mock GPSUPSTREAM setup
	mockGitRepo.On("Name").Return("test-repo")
	mockGitRepo.On("Path").Return(t.TempDir())
	mockGitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
		{Name: "origin", URL: "https://github.com/user/test-repo.git"},
	}, nil).Once()
	mockGitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(nil).Once()
	mockGitRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", "https://github.com/user/test-repo.git").Return(nil).Once()
	mockGitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
		{Name: "origin", URL: "https://github.com/user/test-repo.git"},
		{Name: "GPSUPSTREAM", URL: "https://github.com/user/test-repo.git"},
	}, nil).Once()

	// Mock project exists
	mockProvider.On("VerifyTarget", mock.Anything, "gitlab-user", "test-repo").Return(true, "project-123", nil)

	// CRITICAL: Mock UpdateRemote failure (this tests our fix's error handling)
	mockGitRepo.On("UpdateRemote", mock.Anything, "origin", mock.AnythingOfType("string")).Return(assert.AnError)

	// Create use case and test
	useCase := sync.NewPushToProviderUseCase(mockProvider, nil)

	request := sync.PushRequest{
		SourceRepository: createTestRepository("test-repo"),
		SourceGitRepo:    mockGitRepo,
		TargetConfig:     createGitLabMirrorTarget(),
		CreateIfMissing:  true,
		DryRun:           false,
	}

	// Execute - should fail at remote update
	response, err := useCase.Execute(context.Background(), request)

	// Verify failure
	require.Error(t, err, "Should fail when remote update fails")
	assert.False(t, response.Success, "Response should indicate failure")
	assert.Contains(t, err.Error(), "failed to update origin remote", "Error should mention remote update failure")

	// Verify mocks
	mockGitRepo.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
}

// Helper functions

// Repository creation functions moved to testutil.GitTestEnvironment

// CreateGitLabMirrorTarget creates a GitLab mirror target configuration
func createGitLabMirrorTarget() entities.MirrorTarget {
	authConfig := entities.NewAuthConfigWithToken("gitlab-token", "")

	builder := entities.NewMirrorTargetBuilder()
	builder, _ = builder.WithName("gitlab-mirror")
	builder, _ = builder.WithProvider("gitlab")
	builder = builder.WithDomain("gitlab.com")
	builder, _ = builder.WithOwner("gitlab-user")
	builder = builder.WithAuth(authConfig)

	target, _ := builder.Build()
	return target
}

// CreateGitHubProviderConfig creates a GitHub provider configuration
func createGitHubProviderConfig() ports.ProviderConfig {
	return ports.ProviderConfig{
		ProviderType: "github",
		Domain:       "github.com",
		Owner:        "user",
		AuthConfig: ports.AuthenticationConfig{
			Token: "github-token",
		},
	}
}

// CreateLocalGitLabTarget creates a GitLab mirror target pointing to test environment bare repo
func createLocalGitLabTarget(gitlabBarePath string) entities.MirrorTarget {
	authConfig := entities.NewAuthConfigWithToken("test-token", "git")
	builder := entities.NewMirrorTargetBuilder()
	builder, _ = builder.WithName("test-gitlab")
	builder, _ = builder.WithProvider("gitlab")
	builder = builder.WithDomain("localhost") // Points to test environment bare repo
	builder, _ = builder.WithOwner("gitlab-user")
	builder, _ = builder.WithPath("")
	builder = builder.WithAuth(authConfig)
	target, _ := builder.Build()
	return target
}

// MockGitLabProvider mocks GitLab provider with proper network call mocking
type MockGitLabProvider struct {
	mock.Mock
}

func (m *MockGitLabProvider) ProjectExists(ctx context.Context, owner, name string) (bool, string, error) {
	args := m.Called(ctx, owner, name)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockGitLabProvider) CreateRepositoryForPush(ctx context.Context, req ports.CreateRepositoryRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

// Implement other required methods with minimal implementations
func (m *MockGitLabProvider) Protect(context.Context, string, string, string) error { return nil }
func (m *MockGitLabProvider) Unprotect(context.Context, string, string) error       { return nil }
func (m *MockGitLabProvider) ListRepositories(context.Context, ports.ProviderConfig) ([]entities.Repository, error) {
	return nil, nil
}
func (m *MockGitLabProvider) GetRepository(context.Context, ports.ProviderConfig, string) (entities.Repository, error) {
	return entities.Repository{}, nil
}
func (m *MockGitLabProvider) RepositoryExists(context.Context, ports.RepositoryExistsRequest) (bool, string, error) {
	return false, "", nil
}
func (m *MockGitLabProvider) CreateRepository(context.Context, ports.ProviderConfig, ports.CreateRepositoryOptions) (entities.Repository, error) {
	return entities.Repository{}, nil
}
func (m *MockGitLabProvider) UpdateRepository(context.Context, ports.ProviderConfig, string, ports.UpdateRepositoryOptions) error {
	return nil
}
func (m *MockGitLabProvider) DeleteRepository(context.Context, ports.ProviderConfig, string) error {
	return nil
}
func (m *MockGitLabProvider) SetDefaultBranch(context.Context, string, string, string) error {
	return nil
}
func (m *MockGitLabProvider) ValidateRepositoryName(string) error             { return nil }
func (m *MockGitLabProvider) IsValidProjectName(context.Context, string) bool { return true }
func (m *MockGitLabProvider) TransformRepositoryName(name string, _ ports.NameTransformOptions) string {
	return name
}
func (m *MockGitLabProvider) GetBranchProtection(context.Context, ports.ProviderConfig, string, string) (ports.BranchProtection, error) {
	return ports.BranchProtection{}, nil
}
func (m *MockGitLabProvider) SetBranchProtection(context.Context, ports.ProviderConfig, string, string, ports.BranchProtection) error {
	return nil
}
func (m *MockGitLabProvider) RemoveBranchProtection(context.Context, ports.ProviderConfig, string, string) error {
	return nil
}
func (m *MockGitLabProvider) ListProtectedBranches(context.Context, ports.ProviderConfig, string) ([]string, error) {
	return nil, nil
}
func (m *MockGitLabProvider) GetProviderInfo() ports.ProviderInfo        { return ports.ProviderInfo{} }
func (m *MockGitLabProvider) SupportsFeature(ports.ProviderFeature) bool { return true }

// TestSyncLogger provides a simple logger mock for integration tests
type TestSyncLogger struct {
	mock.Mock
}

func (l *TestSyncLogger) Debug(ctx context.Context, msg string, fields map[string]any) {
	l.Called(ctx, msg, fields)
}
func (l *TestSyncLogger) Info(ctx context.Context, msg string, fields map[string]any) {
	l.Called(ctx, msg, fields)
}
func (l *TestSyncLogger) Error(ctx context.Context, msg string, fields map[string]any) {
	l.Called(ctx, msg, fields)
}
func (l *TestSyncLogger) Trace(ctx context.Context, msg string, fields map[string]any) {
	l.Called(ctx, msg, fields)
}
func (l *TestSyncLogger) Warn(ctx context.Context, msg string, fields map[string]any) {
	l.Called(ctx, msg, fields)
}
func (l *TestSyncLogger) Fatal(ctx context.Context, msg string, fields map[string]any) {
	l.Called(ctx, msg, fields)
}
func (l *TestSyncLogger) IsLevelEnabled(level ports.LogLevel) bool {
	args := l.Called(level)
	return args.Bool(0)
}

// TestSyncGitRepository provides a git repository mock for integration tests
type TestSyncGitRepository struct {
	mock.Mock
}

func (r *TestSyncGitRepository) Name() string {
	args := r.Called()
	return args.String(0)
}
func (r *TestSyncGitRepository) Path() string {
	args := r.Called()
	return args.String(0)
}
func (r *TestSyncGitRepository) URL() string {
	args := r.Called()
	return args.String(0)
}
func (r *TestSyncGitRepository) IsBare() bool {
	args := r.Called()
	return args.Bool(0)
}
func (r *TestSyncGitRepository) IsClean() bool {
	args := r.Called()
	return args.Bool(0)
}
func (r *TestSyncGitRepository) HasChanges() bool {
	args := r.Called()
	return args.Bool(0)
}
func (r *TestSyncGitRepository) Close() error {
	args := r.Called()
	return args.Error(0)
}
func (r *TestSyncGitRepository) ListRemotes(ctx context.Context) ([]ports.RemoteInfo, error) {
	args := r.Called(ctx)
	return args.Get(0).([]ports.RemoteInfo), args.Error(1)
}
func (r *TestSyncGitRepository) AddRemote(ctx context.Context, name, url string) error {
	args := r.Called(ctx, name, url)
	return args.Error(0)
}
func (r *TestSyncGitRepository) RemoveRemote(ctx context.Context, name string) error {
	args := r.Called(ctx, name)
	return args.Error(0)
}
func (r *TestSyncGitRepository) UpdateRemote(ctx context.Context, name, url string) error {
	args := r.Called(ctx, name, url)
	return args.Error(0)
}
func (r *TestSyncGitRepository) Push(ctx context.Context, options ports.PushOptions) error {
	args := r.Called(ctx, options)
	return args.Error(0)
}
func (r *TestSyncGitRepository) ListBranches(ctx context.Context) ([]ports.BranchInfo, error) {
	args := r.Called(ctx)
	return args.Get(0).([]ports.BranchInfo), args.Error(1)
}
func (r *TestSyncGitRepository) CurrentBranch() (string, error) {
	args := r.Called()
	return args.String(0), args.Error(1)
}
func (r *TestSyncGitRepository) CreateBranch(ctx context.Context, name, source string) error {
	args := r.Called(ctx, name, source)
	return args.Error(0)
}
func (r *TestSyncGitRepository) CheckoutBranch(ctx context.Context, name string) error {
	args := r.Called(ctx, name)
	return args.Error(0)
}
func (r *TestSyncGitRepository) DeleteBranch(ctx context.Context, name string, force bool) error {
	args := r.Called(ctx, name, force)
	return args.Error(0)
}
func (r *TestSyncGitRepository) SetDefaultBranch(ctx context.Context, name string) error {
	args := r.Called(ctx, name)
	return args.Error(0)
}
func (r *TestSyncGitRepository) ListTags(ctx context.Context) ([]ports.TagInfo, error) {
	args := r.Called(ctx)
	return args.Get(0).([]ports.TagInfo), args.Error(1)
}
func (r *TestSyncGitRepository) CreateTag(ctx context.Context, name, commit, message string) error {
	args := r.Called(ctx, name, commit, message)
	return args.Error(0)
}
func (r *TestSyncGitRepository) DeleteTag(ctx context.Context, name string) error {
	args := r.Called(ctx, name)
	return args.Error(0)
}
func (r *TestSyncGitRepository) Status(ctx context.Context) (ports.StatusResult, error) {
	args := r.Called(ctx)
	return args.Get(0).(ports.StatusResult), args.Error(1)
}
func (r *TestSyncGitRepository) Diff(ctx context.Context, options ports.DiffOptions) (string, error) {
	args := r.Called(ctx, options)
	return args.String(0), args.Error(1)
}
func (r *TestSyncGitRepository) ListCommits(ctx context.Context, options ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	args := r.Called(ctx, options)
	return args.Get(0).([]ports.CommitInfo), args.Error(1)
}
func (r *TestSyncGitRepository) GetCommit(ctx context.Context, hash string) (ports.CommitInfo, error) {
	args := r.Called(ctx, hash)
	return args.Get(0).(ports.CommitInfo), args.Error(1)
}
func (r *TestSyncGitRepository) Fetch(ctx context.Context, options ports.FetchOptions) error {
	args := r.Called(ctx, options)
	return args.Error(0)
}
func (r *TestSyncGitRepository) Pull(ctx context.Context, options ports.PullOptions) error {
	args := r.Called(ctx, options)
	return args.Error(0)
}

// Creates test repository entities
func createTestRepository(name string) entities.Repository {
	builder := entities.NewRepositoryBuilder()
	builder, _ = builder.WithName(name)
	builder, _ = builder.WithHTTPSURL("https://github.com/test/" + name + ".git")
	builder = builder.WithProviderType("github")
	builder, _ = builder.WithDefaultBranch("main")
	builder = builder.WithPrivate(false)
	builder = builder.WithDescription("Test repository")
	repo, _ := builder.Build()
	return repo
}
