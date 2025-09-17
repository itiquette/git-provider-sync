// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package mirror

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Constants for testing.
const (
	mockRepoURL  = "https://example.com/repo.git"
	mockTempPath = "/test/mock/tmp" // Mock temp path for testing
)

// Mock implementations for interpreter testing

type mockGitRepository struct {
	pushError  error
	closeError error
}

func (m *mockGitRepository) Push(_ context.Context, _ ports.PushOptions) error {
	return m.pushError
}

func (m *mockGitRepository) Close() error {
	return m.closeError
}

// GitRepositoryInfo interface.
func (m *mockGitRepository) Path() string     { return "/test/path" }
func (m *mockGitRepository) URL() string      { return mockRepoURL }
func (m *mockGitRepository) Name() string     { return "test-repo" }
func (m *mockGitRepository) IsBare() bool     { return false }
func (m *mockGitRepository) IsClean() bool    { return true }
func (m *mockGitRepository) HasChanges() bool { return false }

// GitBranchOperations interface.
func (m *mockGitRepository) CurrentBranch() (string, error) { return "main", nil }
func (m *mockGitRepository) ListBranches(_ context.Context) ([]ports.BranchInfo, error) {
	return []ports.BranchInfo{{Name: "main", Hash: "abc123", IsCurrent: true}}, nil
}
func (m *mockGitRepository) CreateBranch(_ context.Context, _, _ string) error { return nil }
func (m *mockGitRepository) CheckoutBranch(_ context.Context, _ string) error  { return nil }
func (m *mockGitRepository) DeleteBranch(_ context.Context, _ string, _ bool) error {
	return nil
}
func (m *mockGitRepository) SetDefaultBranch(_ context.Context, _ string) error { return nil }

// GitRemoteOperations interface.
func (m *mockGitRepository) ListRemotes(_ context.Context) ([]ports.RemoteInfo, error) {
	return []ports.RemoteInfo{{Name: "origin", URL: mockRepoURL}}, nil
}
func (m *mockGitRepository) AddRemote(_ context.Context, _, _ string) error    { return nil }
func (m *mockGitRepository) RemoveRemote(_ context.Context, _ string) error    { return nil }
func (m *mockGitRepository) UpdateRemote(_ context.Context, _, _ string) error { return nil }

// GitSyncOperations interface.
func (m *mockGitRepository) Fetch(_ context.Context, _ ports.FetchOptions) error { return nil }
func (m *mockGitRepository) Pull(_ context.Context, _ ports.PullOptions) error   { return nil }

// GitCommitOperations interface.
func (m *mockGitRepository) GetCommit(_ context.Context, _ string) (ports.CommitInfo, error) {
	return ports.CommitInfo{Hash: "abc123", Message: "test commit"}, nil
}
func (m *mockGitRepository) ListCommits(_ context.Context, _ ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	return []ports.CommitInfo{{Hash: "abc123", Message: "test commit"}}, nil
}

// GitTagOperations interface.
func (m *mockGitRepository) ListTags(_ context.Context) ([]ports.TagInfo, error) {
	return []ports.TagInfo{}, nil
}
func (m *mockGitRepository) CreateTag(_ context.Context, _, _, _ string) error {
	return nil
}
func (m *mockGitRepository) DeleteTag(_ context.Context, _ string) error { return nil }

// GitStatusOperations interface.
func (m *mockGitRepository) Status(_ context.Context) (ports.StatusResult, error) {
	return ports.StatusResult{}, nil
}
func (m *mockGitRepository) Diff(_ context.Context, _ ports.DiffOptions) (string, error) {
	return "", nil
}

type mockGitOpsWithErrors struct {
	cloneError   error
	openError    error
	cleanupError error
	returnRepo   *mockGitRepository
}

func (m *mockGitOpsWithErrors) Clone(_ context.Context, _ ports.CloneOptions) (ports.GitRepository, error) {
	if m.cloneError != nil {
		return nil, m.cloneError
	}

	if m.returnRepo != nil {
		return m.returnRepo, nil
	}

	return &mockGitRepository{}, nil
}

func (m *mockGitOpsWithErrors) Open(_ context.Context, _ string) (ports.GitRepository, error) {
	if m.openError != nil {
		return nil, m.openError
	}

	if m.returnRepo != nil {
		return m.returnRepo, nil
	}

	return &mockGitRepository{}, nil
}

func (m *mockGitOpsWithErrors) Cleanup(_ context.Context, _ string) error {
	return m.cleanupError
}

func (m *mockGitOpsWithErrors) Init(_ context.Context, _ string, _ ports.InitOptions) (ports.GitRepository, error) {
	return &mockGitRepository{}, nil
}

func (m *mockGitOpsWithErrors) SupportsURL(_ string) bool {
	return true
}

func (m *mockGitOpsWithErrors) GetName() string {
	return "mock-git-ops"
}

func (m *mockGitOpsWithErrors) CreateTmpDir(ctx context.Context, _, _ string) (context.Context, error) {
	return ctx, nil
}

func (m *mockGitOpsWithErrors) GetTmpDirPath(_ context.Context) (string, error) {
	return mockTempPath, nil
}

func (m *mockGitOpsWithErrors) DeleteTmpDir(_ context.Context) error {
	return nil
}

type mockRepoProviderWithErrors struct {
	createError error
	updateError error
}

func (m *mockRepoProviderWithErrors) CreateRepositoryForPush(_ context.Context, _ ports.CreateRepositoryRequest) (string, error) {
	if m.createError != nil {
		return "", m.createError
	}

	return mockRepoURL, nil
}

func (m *mockRepoProviderWithErrors) UpdateRepository(_ context.Context, _ ports.ProviderConfig, _ string, _ ports.UpdateRepositoryOptions) error {
	return m.updateError
}

func (m *mockRepoProviderWithErrors) GetRepository(_ context.Context, _ ports.ProviderConfig, name string) (entities.Repository, error) {
	repo, err := entities.NewRepositoryBuilder().WithName(name)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository name: %w", err)
	}

	// Add a mock URL so the repository is valid
	repo, err = repo.WithHTTPSURL("https://mock-provider.example.com/owner/" + name + ".git")
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set HTTPS URL: %w", err)
	}

	built, err := repo.Build()
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to build repository: %w", err)
	}

	return built, nil
}

func (m *mockRepoProviderWithErrors) ListRepositories(_ context.Context, _ ports.ProviderConfig) ([]entities.Repository, error) {
	return []entities.Repository{}, nil
}

func (m *mockRepoProviderWithErrors) DeleteRepository(_ context.Context, _ ports.ProviderConfig, _ string) error {
	return nil
}

// Missing RepositoryProvider interface methods.
func (m *mockRepoProviderWithErrors) CreateRepository(_ context.Context, _ ports.ProviderConfig, options ports.CreateRepositoryOptions) (entities.Repository, error) {
	if m.createError != nil {
		return entities.Repository{}, m.createError
	}

	repo, err := entities.NewRepositoryBuilder().WithName(options.Name)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository name: %w", err)
	}

	// Add a mock URL so the repository is valid
	repo, err = repo.WithHTTPSURL("https://mock-provider.example.com/owner/" + options.Name + ".git")
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set HTTPS URL: %w", err)
	}

	built, err := repo.Build()
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to build repository: %w", err)
	}

	return built, nil
}

func (m *mockRepoProviderWithErrors) SetDefaultBranch(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockRepoProviderWithErrors) RepositoryExists(_ context.Context, _ ports.RepositoryExistsRequest) (bool, string, error) {
	return true, mockRepoURL, nil
}

func (m *mockRepoProviderWithErrors) ValidateRepositoryName(_ string) error {
	return nil
}

func (m *mockRepoProviderWithErrors) GetBranchProtection(_ context.Context, _ ports.ProviderConfig, _, _ string) (ports.BranchProtection, error) {
	return ports.BranchProtection{}, nil
}

func (m *mockRepoProviderWithErrors) SetBranchProtection(_ context.Context, _ ports.ProviderConfig, _, _ string, _ ports.BranchProtection) error {
	return nil
}

func (m *mockRepoProviderWithErrors) RemoveBranchProtection(_ context.Context, _ ports.ProviderConfig, _, _ string) error {
	return nil
}

func (m *mockRepoProviderWithErrors) GetProviderInfo() ports.ProviderInfo {
	return ports.ProviderInfo{Name: "mock-provider", Type: "git"}
}

func (m *mockRepoProviderWithErrors) GetCapabilities() ports.ProviderCapabilities {
	return ports.ProviderCapabilities{SupportsPrivateRepos: true}
}

func (m *mockRepoProviderWithErrors) PushToProvider(_ context.Context, _ ports.ProviderConfig, _ string, _ ports.SyncOptions) (ports.SyncResult, error) {
	return ports.SyncResult{}, nil
}

func (m *mockRepoProviderWithErrors) ProjectExists(_ context.Context, _, _ string) (bool, string, error) {
	return true, "123", nil
}

func (m *mockRepoProviderWithErrors) Protect(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockRepoProviderWithErrors) Unprotect(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockRepoProviderWithErrors) IsValidProjectName(_ context.Context, _ string) bool {
	return true
}

func (m *mockRepoProviderWithErrors) SupportsFeature(_ ports.ProviderFeature) bool {
	return true
}

func (m *mockRepoProviderWithErrors) TransformRepositoryName(name string, _ ports.NameTransformOptions) string {
	return name
}

func (m *mockRepoProviderWithErrors) ListProtectedBranches(_ context.Context, _ ports.ProviderConfig, _ string) ([]string, error) {
	return []string{}, nil
}

type mockLoggerWithCapture struct {
	debugMessages []string
	infoMessages  []string
	errorMessages []string
}

func (m *mockLoggerWithCapture) Debug(_ context.Context, msg string, _ map[string]any) {
	m.debugMessages = append(m.debugMessages, msg)
}

func (m *mockLoggerWithCapture) Info(_ context.Context, msg string, _ map[string]any) {
	m.infoMessages = append(m.infoMessages, msg)
}

func (m *mockLoggerWithCapture) Error(_ context.Context, msg string, _ map[string]any) {
	m.errorMessages = append(m.errorMessages, msg)
}

func (m *mockLoggerWithCapture) Fatal(_ context.Context, msg string, _ map[string]any) {
	m.errorMessages = append(m.errorMessages, msg)
}

func (m *mockLoggerWithCapture) IsLevelEnabled(_ ports.LogLevel) bool {
	return true
}

func (m *mockLoggerWithCapture) Trace(_ context.Context, msg string, _ map[string]any) {
	m.debugMessages = append(m.debugMessages, msg)
}

func (m *mockLoggerWithCapture) Warn(_ context.Context, msg string, _ map[string]any) {
	m.infoMessages = append(m.infoMessages, msg)
}

// Test EffectInterpreter constructor

func TestNewEffectInterpreter(t *testing.T) {
	t.Parallel()

	gitOps := &mockGitOpsWithErrors{}
	repoProvider := &mockRepoProviderWithErrors{}
	logger := &mockLoggerWithCapture{}

	interpreter := NewEffectInterpreter(gitOps, repoProvider, logger)

	require.NotNil(t, interpreter)
	assert.Equal(t, gitOps, interpreter.gitOps)
	assert.Equal(t, repoProvider, interpreter.repoProvider)
	assert.Equal(t, logger, interpreter.logger)
}

// Test ExecuteOperation

func TestEffectInterpreter_ExecuteOperation_Success(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	operation := createValidOperation()

	result := interpreter.ExecuteOperation(ctx, operation)

	assert.True(t, result.Success)
	require.NoError(t, result.Error)
	assert.Equal(t, operation, result.Operation)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.NotEmpty(t, result.Effects)
}

func TestEffectInterpreter_ExecuteOperation_ValidationFailure(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	// Create operation with invalid source URL
	operation := Operation{
		Type: OperationTypeCloneAndMirror,
		Source: RepositorySpec{
			URL:  "", // Invalid - empty URL
			Name: "source",
		},
		Target: RepositorySpec{
			URL:  "https://target.com/repo.git",
			Name: "target",
		},
		Options: OperationOptions{DryRun: false},
		Validations: []ValidationRule{
			{Name: "ValidSourceURL", Predicate: validateSourceURL},
		},
		Effects: []Effect{
			{Type: EffectTypeCloneRepository, Description: "Clone"},
		},
		Metadata: OperationMetadata{ID: "test-op"},
	}

	result := interpreter.ExecuteOperation(ctx, operation)

	assert.False(t, result.Success)
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "validation failed")
	assert.Empty(t, result.Effects) // No effects executed due to validation failure
}

func TestEffectInterpreter_ExecuteOperation_EffectFailure(t *testing.T) {
	t.Parallel()

	// Create interpreter with git ops that will fail
	gitOps := &mockGitOpsWithErrors{cloneError: errors.New("clone failed")}
	repoProvider := &mockRepoProviderWithErrors{}
	logger := &mockLoggerWithCapture{}
	interpreter := NewEffectInterpreter(gitOps, repoProvider, logger)

	ctx := context.Background()
	operation := createValidOperation()

	result := interpreter.ExecuteOperation(ctx, operation)

	assert.False(t, result.Success)
	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "clone failed")
	assert.NotEmpty(t, result.Effects)
	assert.False(t, result.Effects[0].Success)
}

// Test individual effect execution

func TestEffectInterpreter_executeCloneRepository_Success(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	effect := Effect{
		Type:        EffectTypeCloneRepository,
		Description: "Clone repository",
		Parameters: map[string]any{
			"url":        "https://github.com/owner/repo.git",
			"local_path": mockTempPath + "/repo",
			"auth":       AuthSpec{Type: ports.AuthTypeToken, Token: "token"},
			"branch":     "main",
		},
	}

	operation := Operation{
		Options:  OperationOptions{DryRun: false},
		Metadata: OperationMetadata{ID: "test-op"},
	}

	result, err := interpreter.executeCloneRepository(ctx, effect, operation)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestEffectInterpreter_executeCloneRepository_DryRun(t *testing.T) {
	t.Parallel()

	logger := &mockLoggerWithCapture{}
	interpreter := NewEffectInterpreter(&mockGitOpsWithErrors{}, &mockRepoProviderWithErrors{}, logger)
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypeCloneRepository,
		Parameters: map[string]any{
			"url":        "https://github.com/owner/repo.git",
			"local_path": mockTempPath + "/repo",
		},
	}

	operation := Operation{
		Options: OperationOptions{DryRun: true},
	}

	result, err := interpreter.executeCloneRepository(ctx, effect, operation)

	require.NoError(t, err)
	assert.Equal(t, "dry_run_clone", result)
	assert.Contains(t, logger.infoMessages, "[DRY RUN] Would clone repository")
}

func TestEffectInterpreter_executeCloneRepository_MissingParameters(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	tests := []struct {
		name          string
		parameters    map[string]any
		expectedError error
	}{
		{
			name:          "missing URL",
			parameters:    map[string]any{"local_path": mockTempPath, "auth": AuthSpec{}},
			expectedError: domain.ErrCloneEffectMissingURL,
		},
		{
			name:          "missing local path",
			parameters:    map[string]any{"url": "https://example.com", "auth": AuthSpec{}},
			expectedError: domain.ErrCloneEffectMissingPath,
		},
		{
			name:          "missing auth",
			parameters:    map[string]any{"url": "https://example.com", "local_path": mockTempPath},
			expectedError: domain.ErrCloneEffectMissingAuth,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			effect := Effect{
				Type:       EffectTypeCloneRepository,
				Parameters: test.parameters,
			}

			operation := Operation{Options: OperationOptions{DryRun: false}}

			result, err := interpreter.executeCloneRepository(ctx, effect, operation)

			require.Error(t, err)
			assert.Equal(t, test.expectedError, err)
			assert.Nil(t, result)
		})
	}
}

func TestEffectInterpreter_executeCreateRepository_Success(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypeCreateRepository,
		Parameters: map[string]any{
			"name":        "test-repo",
			"owner":       "test-owner",
			"description": "Test repository",
			"private":     true,
		},
	}

	operation := Operation{
		Target: RepositorySpec{
			Provider:   "github",
			RemotePath: "github.com",
		},
		Options: OperationOptions{DryRun: false},
	}

	result, err := interpreter.executeCreateRepository(ctx, effect, operation)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestEffectInterpreter_executeCreateRepository_DryRun(t *testing.T) {
	t.Parallel()

	logger := &mockLoggerWithCapture{}
	interpreter := NewEffectInterpreter(&mockGitOpsWithErrors{}, &mockRepoProviderWithErrors{}, logger)
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypeCreateRepository,
		Parameters: map[string]any{
			"name":  "test-repo",
			"owner": "test-owner",
		},
	}

	operation := Operation{Options: OperationOptions{DryRun: true}}

	result, err := interpreter.executeCreateRepository(ctx, effect, operation)

	require.NoError(t, err)
	assert.Equal(t, "dry_run_create", result)
	assert.Contains(t, logger.infoMessages, "[DRY RUN] Would create repository")
}

func TestEffectInterpreter_executePushToRepository_Success(t *testing.T) {
	t.Parallel()

	mockRepo := &mockGitRepository{}
	gitOps := &mockGitOpsWithErrors{returnRepo: mockRepo}
	interpreter := NewEffectInterpreter(gitOps, &mockRepoProviderWithErrors{}, &mockLoggerWithCapture{})
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypePushToRepository,
		Parameters: map[string]any{
			"url":        "https://github.com/owner/repo.git",
			"local_path": mockTempPath + "/repo",
			"auth":       AuthSpec{Type: ports.AuthTypeToken, Token: "token"},
			"force":      true,
		},
	}

	operation := Operation{Options: OperationOptions{DryRun: false}}

	result, err := interpreter.executePushToRepository(ctx, effect, operation)

	require.NoError(t, err)
	assert.Equal(t, "push_success", result)
}

func TestEffectInterpreter_executePushToRepository_DryRun(t *testing.T) {
	t.Parallel()

	logger := &mockLoggerWithCapture{}
	interpreter := NewEffectInterpreter(&mockGitOpsWithErrors{}, &mockRepoProviderWithErrors{}, logger)
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypePushToRepository,
		Parameters: map[string]any{
			"url":        "https://github.com/owner/repo.git",
			"local_path": mockTempPath + "/repo",
		},
	}

	operation := Operation{Options: OperationOptions{DryRun: true}}

	result, err := interpreter.executePushToRepository(ctx, effect, operation)

	require.NoError(t, err)
	assert.Equal(t, "dry_run_push", result)
	assert.Contains(t, logger.infoMessages, "[DRY RUN] Would push to repository")
}

func TestEffectInterpreter_executeUpdateDescription_Success(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypeUpdateDescription,
		Parameters: map[string]any{
			"repository":  "test-repo",
			"description": "Updated description",
		},
	}

	operation := Operation{
		Target: RepositorySpec{
			Provider:   "github",
			RemotePath: "github.com",
			Owner:      "test-owner",
		},
		Options: OperationOptions{DryRun: false},
	}

	result, err := interpreter.executeUpdateDescription(ctx, effect, operation)

	require.NoError(t, err)
	assert.Equal(t, "description_updated", result)
}

func TestEffectInterpreter_executeUpdateTopics_Success(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypeUpdateTopics,
		Parameters: map[string]any{
			"repository": "test-repo",
			"topics":     []string{"topic1", "topic2"},
		},
	}

	operation := Operation{
		Target: RepositorySpec{
			Provider:   "github",
			RemotePath: "github.com",
			Owner:      "test-owner",
		},
		Options: OperationOptions{DryRun: false},
	}

	result, err := interpreter.executeUpdateTopics(ctx, effect, operation)

	require.NoError(t, err)
	assert.Equal(t, "topics_updated", result)
}

func TestEffectInterpreter_executeCleanupTempFiles_Success(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypeCleanupTempFiles,
		Parameters: map[string]any{
			"local_path": mockTempPath + "/repo",
		},
	}

	operation := Operation{Options: OperationOptions{DryRun: false}}

	result, err := interpreter.executeCleanupTempFiles(ctx, effect, operation)

	require.NoError(t, err)
	assert.Equal(t, "cleanup_success", result)
}

func TestEffectInterpreter_executeCleanupTempFiles_DryRun(t *testing.T) {
	t.Parallel()

	logger := &mockLoggerWithCapture{}
	interpreter := NewEffectInterpreter(&mockGitOpsWithErrors{}, &mockRepoProviderWithErrors{}, logger)
	ctx := context.Background()

	effect := Effect{
		Type: EffectTypeCleanupTempFiles,
		Parameters: map[string]any{
			"local_path": mockTempPath + "/repo",
		},
	}

	operation := Operation{Options: OperationOptions{DryRun: true}}

	result, err := interpreter.executeCleanupTempFiles(ctx, effect, operation)

	require.NoError(t, err)
	assert.Equal(t, "dry_run_cleanup", result)
	assert.Contains(t, logger.infoMessages, "[DRY RUN] Would cleanup temp files")
}

// Test dependency management

func TestEffectInterpreter_dependenciesSatisfied(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()

	tests := []struct {
		name            string
		effect          Effect
		completed       map[string]bool
		expectSatisfied bool
	}{
		{
			name: "no dependencies",
			effect: Effect{
				Type:      EffectTypeCloneRepository,
				DependsOn: []string{},
			},
			completed:       map[string]bool{},
			expectSatisfied: true,
		},
		{
			name: "dependencies satisfied",
			effect: Effect{
				Type:      EffectTypePushToRepository,
				DependsOn: []string{"clone_repository"},
			},
			completed: map[string]bool{
				"clone_repository": true,
			},
			expectSatisfied: true,
		},
		{
			name: "dependencies not satisfied",
			effect: Effect{
				Type:      EffectTypePushToRepository,
				DependsOn: []string{"clone_repository"},
			},
			completed:       map[string]bool{},
			expectSatisfied: false,
		},
		{
			name: "multiple dependencies partially satisfied",
			effect: Effect{
				Type:      EffectTypeUpdateDescription,
				DependsOn: []string{"clone_repository", "create_repository"},
			},
			completed: map[string]bool{
				"clone_repository": true,
				// Create_repository is missing
			},
			expectSatisfied: false,
		},
		{
			name: "multiple dependencies fully satisfied",
			effect: Effect{
				Type:      EffectTypeUpdateDescription,
				DependsOn: []string{"clone_repository", "create_repository"},
			},
			completed: map[string]bool{
				"clone_repository":  true,
				"create_repository": true,
			},
			expectSatisfied: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := interpreter.dependenciesSatisfied(test.effect, test.completed)
			assert.Equal(t, test.expectSatisfied, result)
		})
	}
}

// Test metrics updating

func TestEffectInterpreter_updateMetrics(t *testing.T) { //nolint:tparallel // Shared metrics variable
	t.Parallel()

	interpreter := createTestInterpreter()
	metrics := &OperationMetrics{}

	tests := []struct {
		name                   string
		effect                 Effect
		completedEffect        CompletedEffect
		expectedNetworkCalls   int
		expectedReposCreated   int
		expectedReposUpdated   int
		expectedReposProcessed int
	}{
		{
			name:                   "clone repository effect",
			effect:                 Effect{Type: EffectTypeCloneRepository},
			completedEffect:        CompletedEffect{Success: true},
			expectedNetworkCalls:   1,
			expectedReposProcessed: 1,
		},
		{
			name:                   "create repository effect - success",
			effect:                 Effect{Type: EffectTypeCreateRepository},
			completedEffect:        CompletedEffect{Success: true},
			expectedNetworkCalls:   1,
			expectedReposCreated:   1,
			expectedReposProcessed: 1,
		},
		{
			name:                   "create repository effect - failure",
			effect:                 Effect{Type: EffectTypeCreateRepository},
			completedEffect:        CompletedEffect{Success: false},
			expectedNetworkCalls:   1,
			expectedReposCreated:   0, // Not incremented on failure
			expectedReposProcessed: 1,
		},
		{
			name:                   "push repository effect - success",
			effect:                 Effect{Type: EffectTypePushToRepository},
			completedEffect:        CompletedEffect{Success: true},
			expectedNetworkCalls:   1,
			expectedReposUpdated:   1,
			expectedReposProcessed: 1,
		},
		{
			name:                   "local effect (cleanup)",
			effect:                 Effect{Type: EffectTypeCleanupTempFiles},
			completedEffect:        CompletedEffect{Success: true},
			expectedNetworkCalls:   0, // Local operation
			expectedReposProcessed: 1,
		},
	}

	for _, test := range tests { //nolint:paralleltest // Shared metrics variable
		t.Run(test.name, func(t *testing.T) {
			// Reset metrics for each test
			metrics = &OperationMetrics{}

			interpreter.updateMetrics(metrics, test.effect, test.completedEffect)

			assert.Equal(t, test.expectedNetworkCalls, metrics.NetworkCalls)
			assert.Equal(t, test.expectedReposCreated, metrics.RepositoriesCreated)
			assert.Equal(t, test.expectedReposUpdated, metrics.RepositoriesUpdated)
			assert.Equal(t, test.expectedReposProcessed, metrics.RepositoriesProcessed)
		})
	}
}

// Test error handling for unknown effect types

func TestEffectInterpreter_executeEffect_UnknownEffectType(t *testing.T) {
	t.Parallel()

	interpreter := createTestInterpreter()
	ctx := context.Background()

	effect := Effect{
		Type:        EffectType("unknown_effect"),
		Description: "Unknown effect",
	}

	operation := Operation{
		Options:  OperationOptions{DryRun: false},
		Metadata: OperationMetadata{ID: "test-op"},
	}

	completedEffect := interpreter.executeEffect(ctx, effect, operation)

	assert.False(t, completedEffect.Success)
	require.Error(t, completedEffect.Error)
	assert.Contains(t, completedEffect.Error.Error(), "unknown effect type")
}

// Helper functions

func createTestInterpreter() *EffectInterpreter {
	return NewEffectInterpreter(
		&mockGitOpsWithErrors{},
		&mockRepoProviderWithErrors{},
		&mockLoggerWithCapture{},
	)
}

func createValidOperation() Operation {
	return Operation{
		Type: OperationTypeCloneAndMirror,
		Source: RepositorySpec{
			URL:       "https://github.com/owner/source.git",
			Name:      "source",
			LocalPath: mockTempPath + "/source",
			Auth:      AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
		},
		Target: RepositorySpec{
			URL:        "https://gitlab.com/owner/target.git",
			Name:       "target",
			Provider:   "gitlab",
			RemotePath: "gitlab.com",
			Owner:      "owner",
			Auth:       AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
		},
		Options: OperationOptions{DryRun: false},
		Effects: []Effect{
			{
				Type:        EffectTypeCloneRepository,
				Description: "Clone source repository",
				Parameters: map[string]any{
					"url":        "https://github.com/owner/source.git",
					"local_path": mockTempPath + "/source",
					"auth":       AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
					"branch":     "main",
				},
			},
		},
		Validations: []ValidationRule{
			{Name: "ValidSourceURL", Predicate: validateSourceURL},
			{Name: "ValidTargetURL", Predicate: validateTargetURL},
		},
		Metadata: OperationMetadata{
			ID:        "test-operation",
			CreatedAt: time.Now(),
			Priority:  PriorityNormal,
		},
	}
}
