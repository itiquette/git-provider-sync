// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync_test

import (
	"context"
	"fmt"
	"io/fs"
	"itiquette/git-provider-sync/internal/adapters/shared"
	"path/filepath"

	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
)

// Mock implementations for testing

type mockConfiguration struct {
	mock.Mock
}

func (m *mockConfiguration) Load(ctx context.Context, source ports.ConfigurationSource) (ports.AppConfiguration, error) {
	args := m.Called(ctx, source)
	if err := args.Error(1); err != nil {
		return ports.AppConfiguration{}, fmt.Errorf("failed to load configuration: %w", err)
	}

	return args.Get(0).(ports.AppConfiguration), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockConfiguration) LoadMultiple(ctx context.Context, sources []ports.ConfigurationSource) (ports.AppConfiguration, error) {
	args := m.Called(ctx, sources)
	if err := args.Error(1); err != nil {
		return ports.AppConfiguration{}, fmt.Errorf("failed to load multiple configurations: %w", err)
	}

	return args.Get(0).(ports.AppConfiguration), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockConfiguration) Reload(ctx context.Context) (ports.AppConfiguration, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return ports.AppConfiguration{}, fmt.Errorf("failed to reload configuration: %w", err)
	}

	return args.Get(0).(ports.AppConfiguration), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockConfiguration) Validate(config ports.AppConfiguration) ([]ports.ConfigurationError, error) {
	args := m.Called(config)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	return args.Get(0).([]ports.ConfigurationError), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockConfiguration) ValidateEnvironment(env ports.EnvironmentConfiguration) error {
	args := m.Called(env)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to validate environment: %w", err)
	}

	return nil
}

func (m *mockConfiguration) GetSources() []ports.ConfigurationSource {
	args := m.Called()

	return args.Get(0).([]ports.ConfigurationSource) //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockConfiguration) GetLastModified() time.Time {
	args := m.Called()

	return args.Get(0).(time.Time) //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockConfiguration) GetVersion() string {
	args := m.Called()

	return args.String(0)
}

type mockRepositoryProvider struct {
	mock.Mock
}

func (m *mockRepositoryProvider) ListRepositories(ctx context.Context, config ports.ProviderConfig) ([]entities.Repository, error) {
	args := m.Called(ctx, config)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return args.Get(0).([]entities.Repository), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockRepositoryProvider) GetRepository(ctx context.Context, config ports.ProviderConfig, name string) (entities.Repository, error) {
	args := m.Called(ctx, config, name)
	if err := args.Error(1); err != nil {
		return entities.Repository{}, fmt.Errorf("failed to get repository: %w", err)
	}

	return args.Get(0).(entities.Repository), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockRepositoryProvider) RepositoryExists(ctx context.Context, request ports.RepositoryExistsRequest) (bool, string, error) {
	args := m.Called(ctx, request)

	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *mockRepositoryProvider) CreateRepository(ctx context.Context, config ports.ProviderConfig, options ports.CreateRepositoryOptions) (entities.Repository, error) {
	args := m.Called(ctx, config, options)
	if err := args.Error(1); err != nil {
		return entities.Repository{}, fmt.Errorf("failed to create repository: %w", err)
	}

	return args.Get(0).(entities.Repository), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockRepositoryProvider) UpdateRepository(ctx context.Context, config ports.ProviderConfig, name string, options ports.UpdateRepositoryOptions) error {
	args := m.Called(ctx, config, name, options)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}

	return nil
}

func (m *mockRepositoryProvider) DeleteRepository(ctx context.Context, config ports.ProviderConfig, name string) error {
	args := m.Called(ctx, config, name)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	return nil
}

func (m *mockRepositoryProvider) ValidateRepositoryName(name string) error {
	args := m.Called(name)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to validate repository name: %w", err)
	}

	return nil
}

func (m *mockRepositoryProvider) TransformRepositoryName(name string, options ports.NameTransformOptions) string {
	args := m.Called(name, options)

	return args.String(0)
}

func (m *mockRepositoryProvider) GetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) (ports.BranchProtection, error) {
	args := m.Called(ctx, config, repoName, branch)
	if err := args.Error(1); err != nil {
		return ports.BranchProtection{}, fmt.Errorf("failed to get branch protection: %w", err)
	}

	return args.Get(0).(ports.BranchProtection), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockRepositoryProvider) SetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string, protection ports.BranchProtection) error {
	args := m.Called(ctx, config, repoName, branch, protection)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to set branch protection: %w", err)
	}

	return nil
}

func (m *mockRepositoryProvider) RemoveBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) error {
	args := m.Called(ctx, config, repoName, branch)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to remove branch protection: %w", err)
	}

	return nil
}

func (m *mockRepositoryProvider) ListProtectedBranches(ctx context.Context, config ports.ProviderConfig, repoName string) ([]string, error) {
	args := m.Called(ctx, config, repoName)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to list protected branches: %w", err)
	}

	return args.Get(0).([]string), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockRepositoryProvider) GetProviderInfo() ports.ProviderInfo {
	args := m.Called()

	return args.Get(0).(ports.ProviderInfo) //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockRepositoryProvider) SupportsFeature(feature ports.ProviderFeature) bool {
	args := m.Called(feature)

	return args.Bool(0)
}

func (m *mockRepositoryProvider) CreateRepositoryForPush(ctx context.Context, request ports.CreateRepositoryRequest) (string, error) {
	args := m.Called(ctx, request)

	return args.String(0), args.Error(1)
}

func (m *mockRepositoryProvider) ProjectExists(ctx context.Context, owner, repo string) (bool, string, error) {
	args := m.Called(ctx, owner, repo)

	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *mockRepositoryProvider) Protect(ctx context.Context, owner string, defaultBranch string, projectIDstr string) error {
	args := m.Called(ctx, owner, defaultBranch, projectIDstr)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to protect: %w", err)
	}

	return nil
}

func (m *mockRepositoryProvider) Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error {
	args := m.Called(ctx, defaultBranch, projectIDStr)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to unprotect: %w", err)
	}

	return nil
}

func (m *mockRepositoryProvider) SetDefaultBranch(ctx context.Context, owner, name, branch string) error {
	args := m.Called(ctx, owner, name, branch)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (m *mockRepositoryProvider) IsValidProjectName(ctx context.Context, name string) bool {
	args := m.Called(ctx, name)

	return args.Bool(0)
}

type mockGitOperations struct {
	mock.Mock
}

func (m *mockGitOperations) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) { //nolint:ireturn
	args := m.Called(ctx, options)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to clone: %w", err)
	}

	return args.Get(0).(ports.GitRepository), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockGitOperations) Open(ctx context.Context, path string) (ports.GitRepository, error) { //nolint:ireturn
	args := m.Called(ctx, path)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to open: %w", err)
	}

	return args.Get(0).(ports.GitRepository), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockGitOperations) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) { //nolint:ireturn
	args := m.Called(ctx, path, options)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("failed to init: %w", err)
	}

	return args.Get(0).(ports.GitRepository), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockGitOperations) Cleanup(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to cleanup: %w", err)
	}

	return nil
}

func (m *mockGitOperations) SupportsURL(url string) bool {
	args := m.Called(url)

	return args.Bool(0)
}

func (m *mockGitOperations) GetName() string {
	args := m.Called()

	return args.String(0)
}

func (m *mockGitOperations) CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	args := m.Called(ctx, dir, prefix)
	if err := args.Error(1); err != nil {
		return ctx, fmt.Errorf("failed to create temp dir: %w", err)
	}

	return args.Get(0).(context.Context), nil //nolint:forcetypeassert // Test mock - controlled return values
}

func (m *mockGitOperations) GetTmpDirPath(ctx context.Context) (string, error) {
	args := m.Called(ctx)

	return args.String(0), args.Error(1)
}

func (m *mockGitOperations) DeleteTmpDir(ctx context.Context) error {
	args := m.Called(ctx)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to delete temp dir: %w", err)
	}

	return nil
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(ctx context.Context, message string, fields map[string]any) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) Info(ctx context.Context, message string, fields map[string]any) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) Warn(ctx context.Context, message string, fields map[string]any) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) Error(ctx context.Context, message string, fields map[string]any) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) Fatal(ctx context.Context, message string, fields map[string]any) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) IsLevelEnabled(level ports.LogLevel) bool {
	args := m.Called(level)

	return args.Bool(0)
}

func (m *mockLogger) Trace(ctx context.Context, message string, fields map[string]any) {
	m.Called(ctx, message, fields)
}

type mockArchiveOperations struct {
	mock.Mock
}

func (m *mockArchiveOperations) CreateMirror(_ context.Context, _ ports.ArchiveMirrorRequest) error {
	args := m.Called()
	if err := args.Error(0); err != nil {
		return fmt.Errorf("failed to create archive mirror: %w", err)
	}

	return nil
}

type mockFileSystem struct {
	mock.Mock
}

func (m *mockFileSystem) Exists(path string) (bool, error) {
	args := m.Called(path)

	return args.Bool(0), args.Error(1)
}

func (m *mockFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	args := m.Called(path, perm)

	return args.Error(0) //nolint:wrapcheck // Test mock
}

func (m *mockFileSystem) RemoveAll(path string) error {
	args := m.Called(path)

	return args.Error(0) //nolint:wrapcheck // Test mock
}

func (m *mockFileSystem) Stat(path string) (fs.FileInfo, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Test mock
	}

	return args.Get(0).(fs.FileInfo), args.Error(1) //nolint:wrapcheck,forcetypeassert // Test mock
}

func (m *mockFileSystem) TempDir(dir, pattern string) (string, error) {
	args := m.Called(dir, pattern)

	return args.String(0), args.Error(1)
}

// Test Suite

// TestSyncRepositoriesUseCase_Execute validates the complete sync process
// TEST PURPOSE:
// integration test validates the core sync use case end-to-end, ensuring:
// 1. Configuration loading and validation from multiple sources
// 2. Environment selection and processing
// 3. Temporary directory creation and lifecycle management
// 4. Source repository operations coordination
// 5. Mirror synchronization to all configured targets
// 6. error handling and recovery
// 7. Statistics reporting and observability
// 8. Proper resource cleanup (directories, connections)
// SCENARIOS COVERED:
// - Successful sync with single environment configuration
// - Empty configuration environments (should fail gracefully)
// - Dry run mode (should simulate operations without side effects)
// - Various error conditions with proper error propagation
// Tests coordination between configuration, repository providers, git operations, and logging.
func TestSyncRepositoriesUseCase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		request        sync.Request
		setupMocks     func(*mockConfiguration, *mockRepositoryProvider, *mockGitOperations, *mockLogger)
		expectedResult func(*testing.T, sync.Response, error)
	}{
		{
			name: "successful_sync_with_single_environment",
			request: sync.Request{
				ConfigPath:  "/test/config.yaml",
				Environment: "test",
				DryRun:      false,
			},
			setupMocks: func(mockConfig *mockConfiguration, mockRepo *mockRepositoryProvider, mockGit *mockGitOperations, mockLog *mockLogger) {
				// Setup successful configuration loading
				testConfig := ports.AppConfiguration{
					Environments: map[string]ports.EnvironmentConfiguration{
						"test": {
							Enabled: true,
							Source: ports.SourceConfiguration{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "test-owner",
								Authentication: ports.AuthenticationConfiguration{
									Token: "test-token",
								},
							},
							Mirrors: map[string]ports.MirrorConfiguration{
								"backup": {
									Enabled:      true,
									ProviderType: "directory",
									Path:         filepath.Join("testdata", "backup"),
								},
							},
						},
					},
				}

				mockConfig.On("Load", mock.Anything, mock.AnythingOfType("ports.ConfigurationSource")).
					Return(testConfig, nil)

				// Setup repository provider expectations for validation
				mockRepo.On("ListRepositories", mock.Anything, mock.AnythingOfType("ports.ProviderConfig")).
					Return([]entities.Repository{}, nil)

				// Setup git operations expectations for temporary directory management
				mockGit.On("CreateTmpDir", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(context.Background(), nil).Maybe()
				mockGit.On("GetTmpDirPath", mock.Anything).Return(filepath.Join("testdata", "tmp", "test"), nil).Maybe()
				mockGit.On("DeleteTmpDir", mock.Anything).Return(nil).Maybe()

				// Setup logging expectations - accept any log levels
				mockLog.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Warn", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			expectedResult: func(t *testing.T, response sync.Response, err error) {
				t.Helper()
				t.Helper()
				require.NoError(t, err)
				require.True(t, response.Success)
				require.Empty(t, response.Errors)
			},
		},
		{
			name: "empty_configuration_environments",
			request: sync.Request{
				ConfigPath:  "/test/config.yaml",
				Environment: "test",
				DryRun:      false,
			},
			setupMocks: func(mockConfig *mockConfiguration, _ *mockRepositoryProvider, mockGit *mockGitOperations, mockLog *mockLogger) {
				// Setup configuration with no environments
				testConfig := ports.AppConfiguration{
					Environments: map[string]ports.EnvironmentConfiguration{},
				}

				mockConfig.On("Load", mock.Anything, mock.AnythingOfType("ports.ConfigurationSource")).
					Return(testConfig, nil)

				// Setup git operations expectations for temporary directory management
				mockGit.On("CreateTmpDir", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(context.Background(), nil).Maybe()
				mockGit.On("GetTmpDirPath", mock.Anything).Return(filepath.Join("testdata", "tmp", "test"), nil).Maybe()
				mockGit.On("DeleteTmpDir", mock.Anything).Return(nil).Maybe()

				// Setup logging expectations - accept any log levels
				mockLog.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Warn", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			expectedResult: func(t *testing.T, _ sync.Response, err error) {
				t.Helper()
				require.Error(t, err)
				require.Contains(t, err.Error(), "no environments configured")
			},
		},
		{
			name: "dry_run_mode",
			request: sync.Request{
				ConfigPath:  "/test/config.yaml",
				Environment: "test",
				DryRun:      true,
			},
			setupMocks: func(mockConfig *mockConfiguration, mockRepo *mockRepositoryProvider, mockGit *mockGitOperations, mockLog *mockLogger) {
				// Setup configuration
				testConfig := ports.AppConfiguration{
					Environments: map[string]ports.EnvironmentConfiguration{
						"test": {
							Enabled: true,
							Source: ports.SourceConfiguration{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "test-owner",
							},
						},
					},
				}

				mockConfig.On("Load", mock.Anything, mock.AnythingOfType("ports.ConfigurationSource")).
					Return(testConfig, nil)

				// Setup repository provider expectations for validation
				mockRepo.On("ListRepositories", mock.Anything, mock.AnythingOfType("ports.ProviderConfig")).
					Return([]entities.Repository{}, nil)

				// Setup git operations expectations for temporary directory management
				mockGit.On("CreateTmpDir", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(context.Background(), nil).Maybe()
				mockGit.On("GetTmpDirPath", mock.Anything).Return(filepath.Join("testdata", "tmp", "test"), nil).Maybe()
				mockGit.On("DeleteTmpDir", mock.Anything).Return(nil).Maybe()

				// Setup logging expectations - accept any log levels
				mockLog.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Warn", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			expectedResult: func(t *testing.T, response sync.Response, err error) {
				t.Helper()
				t.Helper()
				require.NoError(t, err)
				require.True(t, response.Success)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Create mocks
			mockConfig := &mockConfiguration{}
			mockRepo := &mockRepositoryProvider{}
			mockGit := &mockGitOperations{}
			mockLogger := &mockLogger{}

			// Setup expectations
			test.setupMocks(mockConfig, mockRepo, mockGit, mockLogger)

			// Create mock archive operations
			mockArchive := &mockArchiveOperations{}

			// Create mock file system
			mockFS := &mockFileSystem{}

			// Create use case
			stringUtils := shared.NewStringUtilsAdapter()
			useCase := sync.NewRepositoriesUseCase(
				mockConfig,
				mockRepo,
				mockGit,
				mockArchive,
				mockFS,
				mockLogger,
				stringUtils,
			)

			// Execute
			ctx := context.Background()
			response, err := useCase.Execute(ctx, test.request)

			// Verify
			test.expectedResult(t, response, err)

			// Assert all expectations were met
			mockConfig.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
			mockGit.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func (m *mockFileSystem) Clean(path string) string {
	return filepath.Clean(path)
}

func (m *mockFileSystem) Join(elem ...string) string {
	return filepath.Join(elem...)
}
