// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package sync_test

import (
	"context"
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
	return args.Get(0).(ports.AppConfiguration), args.Error(1)
}

func (m *mockConfiguration) LoadMultiple(ctx context.Context, sources []ports.ConfigurationSource) (ports.AppConfiguration, error) {
	args := m.Called(ctx, sources)
	return args.Get(0).(ports.AppConfiguration), args.Error(1)
}

func (m *mockConfiguration) Reload(ctx context.Context) (ports.AppConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(ports.AppConfiguration), args.Error(1)
}

func (m *mockConfiguration) Validate(config ports.AppConfiguration) ([]ports.ConfigurationError, error) {
	args := m.Called(config)
	return args.Get(0).([]ports.ConfigurationError), args.Error(1)
}

func (m *mockConfiguration) ValidateEnvironment(env ports.EnvironmentConfiguration) error {
	args := m.Called(env)
	return args.Error(0)
}

func (m *mockConfiguration) Watch(ctx context.Context, callback ports.ConfigurationChangeCallback) error {
	args := m.Called(ctx, callback)
	return args.Error(0)
}

func (m *mockConfiguration) StopWatching() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockConfiguration) GetSources() []ports.ConfigurationSource {
	args := m.Called()
	return args.Get(0).([]ports.ConfigurationSource)
}

func (m *mockConfiguration) GetLastModified() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
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
	return args.Get(0).([]entities.Repository), args.Error(1)
}

func (m *mockRepositoryProvider) GetRepository(ctx context.Context, config ports.ProviderConfig, name string) (entities.Repository, error) {
	args := m.Called(ctx, config, name)
	return args.Get(0).(entities.Repository), args.Error(1)
}

func (m *mockRepositoryProvider) RepositoryExists(ctx context.Context, request ports.RepositoryExistsRequest) (bool, string, error) {
	args := m.Called(ctx, request)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *mockRepositoryProvider) CreateRepository(ctx context.Context, config ports.ProviderConfig, options ports.CreateRepositoryOptions) (entities.Repository, error) {
	args := m.Called(ctx, config, options)
	return args.Get(0).(entities.Repository), args.Error(1)
}

func (m *mockRepositoryProvider) UpdateRepository(ctx context.Context, config ports.ProviderConfig, name string, options ports.UpdateRepositoryOptions) error {
	args := m.Called(ctx, config, name, options)
	return args.Error(0)
}

func (m *mockRepositoryProvider) DeleteRepository(ctx context.Context, config ports.ProviderConfig, name string) error {
	args := m.Called(ctx, config, name)
	return args.Error(0)
}

func (m *mockRepositoryProvider) ValidateRepositoryName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *mockRepositoryProvider) TransformRepositoryName(name string, options ports.NameTransformOptions) string {
	args := m.Called(name, options)
	return args.String(0)
}

func (m *mockRepositoryProvider) GetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) (ports.BranchProtection, error) {
	args := m.Called(ctx, config, repoName, branch)
	return args.Get(0).(ports.BranchProtection), args.Error(1)
}

func (m *mockRepositoryProvider) SetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string, protection ports.BranchProtection) error {
	args := m.Called(ctx, config, repoName, branch, protection)
	return args.Error(0)
}

func (m *mockRepositoryProvider) RemoveBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) error {
	args := m.Called(ctx, config, repoName, branch)
	return args.Error(0)
}

func (m *mockRepositoryProvider) ListProtectedBranches(ctx context.Context, config ports.ProviderConfig, repoName string) ([]string, error) {
	args := m.Called(ctx, config, repoName)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockRepositoryProvider) GetProviderInfo() ports.ProviderInfo {
	args := m.Called()
	return args.Get(0).(ports.ProviderInfo)
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
	return args.Error(0)
}

func (m *mockRepositoryProvider) Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error {
	args := m.Called(ctx, defaultBranch, projectIDStr)
	return args.Error(0)
}

func (m *mockRepositoryProvider) SetDefaultBranch(ctx context.Context, owner, name, branch string) error {
	args := m.Called(ctx, owner, name, branch)
	return args.Error(0)
}

func (m *mockRepositoryProvider) IsValidProjectName(ctx context.Context, name string) bool {
	args := m.Called(ctx, name)
	return args.Bool(0)
}

type mockGitOperations struct {
	mock.Mock
}

func (m *mockGitOperations) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(ports.GitRepository), args.Error(1)
}

func (m *mockGitOperations) Open(ctx context.Context, path string) (ports.GitRepository, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(ports.GitRepository), args.Error(1)
}

func (m *mockGitOperations) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) {
	args := m.Called(ctx, path, options)
	return args.Get(0).(ports.GitRepository), args.Error(1)
}

func (m *mockGitOperations) Cleanup(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *mockGitOperations) SupportsURL(url string) bool {
	args := m.Called(url)
	return args.Bool(0)
}

func (m *mockGitOperations) GetName() string {
	args := m.Called()
	return args.String(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(ctx context.Context, message string, fields map[string]interface{}) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) Warn(ctx context.Context, message string, fields map[string]interface{}) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) Fatal(ctx context.Context, message string, fields map[string]interface{}) {
	m.Called(ctx, message, fields)
}

func (m *mockLogger) IsLevelEnabled(level ports.LogLevel) bool {
	args := m.Called(level)
	return args.Bool(0)
}

func (m *mockLogger) Trace(ctx context.Context, message string, fields map[string]interface{}) {
	m.Called(ctx, message, fields)
}

// Test Suite

func TestSyncRepositoriesUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		request        sync.SyncRequest
		setupMocks     func(*mockConfiguration, *mockRepositoryProvider, *mockGitOperations, *mockLogger)
		expectedResult func(*testing.T, sync.SyncResponse, error)
	}{
		{
			name: "successful_sync_with_single_environment",
			request: sync.SyncRequest{
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
									Path:         "/tmp/backup",
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

				// Setup logging expectations - accept any log levels
				mockLog.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Warn", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			expectedResult: func(t *testing.T, response sync.SyncResponse, err error) {
				require.NoError(t, err)
				require.True(t, response.Success)
				require.Empty(t, response.Errors)
			},
		},
		{
			name: "empty_configuration_environments",
			request: sync.SyncRequest{
				ConfigPath:  "/test/config.yaml",
				Environment: "test",
				DryRun:      false,
			},
			setupMocks: func(mockConfig *mockConfiguration, mockRepo *mockRepositoryProvider, mockGit *mockGitOperations, mockLog *mockLogger) {
				// Setup configuration with no environments
				testConfig := ports.AppConfiguration{
					Environments: map[string]ports.EnvironmentConfiguration{},
				}

				mockConfig.On("Load", mock.Anything, mock.AnythingOfType("ports.ConfigurationSource")).
					Return(testConfig, nil)

				// Setup logging expectations - accept any log levels
				mockLog.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Warn", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			expectedResult: func(t *testing.T, response sync.SyncResponse, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "no environments configured")
			},
		},
		{
			name: "dry_run_mode",
			request: sync.SyncRequest{
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

				// Setup logging expectations - accept any log levels
				mockLog.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Warn", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				mockLog.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			expectedResult: func(t *testing.T, response sync.SyncResponse, err error) {
				require.NoError(t, err)
				require.True(t, response.Success)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockConfig := &mockConfiguration{}
			mockRepo := &mockRepositoryProvider{}
			mockGit := &mockGitOperations{}
			mockLogger := &mockLogger{}

			// Setup expectations
			tt.setupMocks(mockConfig, mockRepo, mockGit, mockLogger)

			// Create use case
			useCase := sync.NewSyncRepositoriesUseCase(
				mockConfig,
				mockRepo,
				mockGit,
				mockLogger,
			)

			// Execute
			ctx := context.Background()
			response, err := useCase.Execute(ctx, tt.request)

			// Verify
			tt.expectedResult(t, response, err)

			// Assert all expectations were met
			mockConfig.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
			mockGit.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestSyncRepositoriesUseCase_NewSyncRepositoriesUseCase(t *testing.T) {
	mockConfig := &mockConfiguration{}
	mockRepo := &mockRepositoryProvider{}
	mockGit := &mockGitOperations{}
	mockLogger := &mockLogger{}

	useCase := sync.NewSyncRepositoriesUseCase(mockConfig, mockRepo, mockGit, mockLogger)

	require.NotNil(t, useCase)
}
