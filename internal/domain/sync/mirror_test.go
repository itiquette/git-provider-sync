// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	sharedadapters "itiquette/git-provider-sync/internal/adapters/shared"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestToMirrorsUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockProvider := &SharedMockRepositoryProvider{}
	mockGitOps := &SharedMockGitOperations{}
	mockLogger := &SharedMockLogger{}

	// Create a mock repository
	mockRepo := &SharedMockGitRepository{}
	mockRepo.On("Name").Return("test-repo")
	mockRepo.On("Path").Return("/tmp/test-repo")
	mockRepo.On("URL").Return("https://github.com/user/test-repo.git")

	// Create mirror target using builder
	mirrorTarget := entities.NewMirrorTarget(
		"gitlab-mirror",
		entities.ProviderTypeGitLab,
		"gitlab.com",
		"backup-user",
		"",
		entities.AuthConfig{},
		true,
	)

	request := ToMirrorsRequest{
		SourceRepositories: []ports.GitRepository{mockRepo},
		MirrorTargets:      []entities.MirrorTarget{mirrorTarget},
		SourceConfig: ports.ProviderConfig{
			ProviderType: "github",
			Domain:       "github.com",
			Owner:        "user",
		},
		DryRun: false,
		Options: Options{
			CreateIfNotExists: true,
		},
	}

	// Setup expectations
	mockProvider.On("ProjectExists", mock.Anything, "backup-user", "test-repo").
		Return(false, "", nil)

	mockProvider.On("CreateRepositoryForPush", mock.Anything, mock.MatchedBy(func(req ports.CreateRepositoryRequest) bool {
		return req.Name == "test-repo"
	})).Return("project-123", nil)

	// Setup logger (lenient for all log levels)
	mockLogger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()

	// Create filesystem
	fs := filesystem.NewOSFileSystem()

	// Create string utils
	stringUtils := sharedadapters.NewStringUtilsAdapter()

	// Create use case
	useCase := NewToMirrorsUseCase(
		mockProvider,
		mockGitOps,
		nil, // archive ops not needed for this test
		fs,
		mockLogger,
		stringUtils,
	)

	// Execute
	ctx := context.Background()
	response, err := useCase.Execute(ctx, request)

	// Assertions
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, 1, response.TotalRepositories)
	assert.Equal(t, 1, response.SuccessfulSyncs)
	assert.Equal(t, 0, response.FailedSyncs)
	assert.Len(t, response.Results, 1)

	result := response.Results[0]
	assert.Equal(t, "test-repo", result.RepositoryName)
	assert.Equal(t, "gitlab-mirror", result.MirrorName)
	assert.True(t, result.Success)
	assert.False(t, result.Skipped)
	assert.NoError(t, result.Error)

	// Verify mock expectations
	mockProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestToMirrorsUseCase_Execute_DryRun(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockProvider := &SharedMockRepositoryProvider{}
	mockGitOps := &SharedMockGitOperations{}
	mockLogger := &SharedMockLogger{}

	// Create a mock repository
	mockRepo := &SharedMockGitRepository{}
	mockRepo.On("Name").Return("test-repo")
	mockRepo.On("Path").Return("/tmp/test-repo")
	mockRepo.On("URL").Return("https://github.com/user/test-repo.git")

	// Create mirror target
	mirrorTarget := entities.NewMirrorTarget(
		"gitlab-mirror",
		entities.ProviderTypeGitLab,
		"gitlab.com",
		"backup-user",
		"",
		entities.AuthConfig{},
		true,
	)

	request := ToMirrorsRequest{
		SourceRepositories: []ports.GitRepository{mockRepo},
		MirrorTargets:      []entities.MirrorTarget{mirrorTarget},
		SourceConfig: ports.ProviderConfig{
			ProviderType: "github",
			Domain:       "github.com",
			Owner:        "user",
		},
		DryRun: true, // Dry run mode
		Options: Options{
			CreateIfNotExists: true,
		},
	}

	// In dry run, we might check existence but don't create
	mockProvider.On("ProjectExists", mock.Anything, "backup-user", "test-repo").
		Return(false, "", nil).Maybe()

	// Setup logger (lenient for all log levels)
	mockLogger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()

	// Create filesystem
	fs := filesystem.NewOSFileSystem()

	// Create string utils
	stringUtils := sharedadapters.NewStringUtilsAdapter()

	// Create use case
	useCase := NewToMirrorsUseCase(
		mockProvider,
		mockGitOps,
		nil,
		fs,
		mockLogger,
		stringUtils,
	)

	// Execute
	ctx := context.Background()
	response, err := useCase.Execute(ctx, request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, 1, response.TotalRepositories)
	assert.Equal(t, 0, response.SuccessfulSyncs) // Nothing actually synced in dry run
	assert.Equal(t, 1, response.SkippedSyncs)    // Skipped due to dry run

	// Verify that CreateRepositoryForPush was NOT called
	mockProvider.AssertNotCalled(t, "CreateRepositoryForPush", mock.Anything, mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestToMirrorsUseCase_Execute_ProviderError(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockProvider := &SharedMockRepositoryProvider{}
	mockGitOps := &SharedMockGitOperations{}
	mockLogger := &SharedMockLogger{}

	// Create a mock repository
	mockRepo := &SharedMockGitRepository{}
	mockRepo.On("Name").Return("test-repo")
	mockRepo.On("Path").Return("/tmp/test-repo")
	mockRepo.On("URL").Return("https://github.com/user/test-repo.git")

	// Create mirror target
	mirrorTarget := entities.NewMirrorTarget(
		"gitlab-mirror",
		entities.ProviderTypeGitLab,
		"gitlab.com",
		"backup-user",
		"",
		entities.AuthConfig{},
		true,
	)

	request := ToMirrorsRequest{
		SourceRepositories: []ports.GitRepository{mockRepo},
		MirrorTargets:      []entities.MirrorTarget{mirrorTarget},
		SourceConfig: ports.ProviderConfig{
			ProviderType: "github",
			Domain:       "github.com",
			Owner:        "user",
		},
		DryRun: false,
		Options: Options{
			CreateIfNotExists: true,
		},
	}

	// Simulate API error
	apiError := errors.New("API rate limit exceeded")
	mockProvider.On("ProjectExists", mock.Anything, "backup-user", "test-repo").
		Return(false, "", apiError)

	// Setup logger (lenient for all log levels)
	mockLogger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()

	// Create filesystem
	fs := filesystem.NewOSFileSystem()

	// Create string utils
	stringUtils := sharedadapters.NewStringUtilsAdapter()

	// Create use case
	useCase := NewToMirrorsUseCase(
		mockProvider,
		mockGitOps,
		nil,
		fs,
		mockLogger,
		stringUtils,
	)

	// Execute
	ctx := context.Background()
	response, err := useCase.Execute(ctx, request)

	// Assertions - operation continues despite individual failures
	require.NoError(t, err) // Overall operation doesn't fail
	assert.False(t, response.Success)
	assert.Equal(t, 1, response.TotalRepositories)
	assert.Equal(t, 0, response.SuccessfulSyncs)
	assert.Equal(t, 1, response.FailedSyncs)
	assert.Len(t, response.Results, 1)

	result := response.Results[0]
	assert.Equal(t, "test-repo", result.RepositoryName)
	assert.Equal(t, "gitlab-mirror", result.MirrorName)
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "API rate limit exceeded")

	// Verify mock expectations
	mockProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestToMirrorsUseCase_Execute_SkipExisting(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockProvider := &SharedMockRepositoryProvider{}
	mockGitOps := &SharedMockGitOperations{}
	mockLogger := &SharedMockLogger{}

	// Create a mock repository
	mockRepo := &SharedMockGitRepository{}
	mockRepo.On("Name").Return("test-repo")
	mockRepo.On("Path").Return("/tmp/test-repo")
	mockRepo.On("URL").Return("https://github.com/user/test-repo.git")

	// Create mirror target
	mirrorTarget := entities.NewMirrorTarget(
		"gitlab-mirror",
		entities.ProviderTypeGitLab,
		"gitlab.com",
		"backup-user",
		"",
		entities.AuthConfig{},
		true,
	)

	request := ToMirrorsRequest{
		SourceRepositories: []ports.GitRepository{mockRepo},
		MirrorTargets:      []entities.MirrorTarget{mirrorTarget},
		SourceConfig: ports.ProviderConfig{
			ProviderType: "github",
			Domain:       "github.com",
			Owner:        "user",
		},
		DryRun: false,
		Options: Options{
			CreateIfNotExists: false, // Don't create if not exists
			UpdateDescription: false, // Don't update existing
		},
	}

	// Repository already exists
	mockProvider.On("ProjectExists", mock.Anything, "backup-user", "test-repo").
		Return(true, "project-123", nil)

	// Setup logger (lenient for all log levels)
	mockLogger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()

	// Create filesystem
	fs := filesystem.NewOSFileSystem()

	// Create string utils
	stringUtils := sharedadapters.NewStringUtilsAdapter()

	// Create use case
	useCase := NewToMirrorsUseCase(
		mockProvider,
		mockGitOps,
		nil,
		fs,
		mockLogger,
		stringUtils,
	)

	// Execute
	ctx := context.Background()
	response, err := useCase.Execute(ctx, request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, 1, response.TotalRepositories)
	assert.Equal(t, 0, response.SuccessfulSyncs)
	assert.Equal(t, 1, response.SkippedSyncs)

	result := response.Results[0]
	assert.True(t, result.Skipped)
	assert.Equal(t, "skipped", result.Action)

	// Verify that CreateRepositoryForPush was NOT called
	mockProvider.AssertNotCalled(t, "CreateRepositoryForPush", mock.Anything, mock.Anything)

	mockProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestToMirrorsUseCase_Execute_EmptyRequest(t *testing.T) {
	t.Parallel()

	// Setup minimal mocks
	mockLogger := &SharedMockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()

	fs := filesystem.NewOSFileSystem()
	stringUtils := sharedadapters.NewStringUtilsAdapter()

	// Create use case
	useCase := NewToMirrorsUseCase(
		&SharedMockRepositoryProvider{},
		&SharedMockGitOperations{},
		nil,
		fs,
		mockLogger,
		stringUtils,
	)

	// Test with empty repositories
	request := ToMirrorsRequest{
		SourceRepositories: []ports.GitRepository{},
		MirrorTargets: []entities.MirrorTarget{
			entities.NewMirrorTarget("mirror", entities.ProviderTypeGitLab, "gitlab.com", "user", "", entities.AuthConfig{}, true),
		},
	}

	ctx := context.Background()
	response, err := useCase.Execute(ctx, request)

	// Should handle gracefully
	require.NoError(t, err)
	assert.Equal(t, 0, response.TotalRepositories)
	assert.Equal(t, 0, response.SuccessfulSyncs)
	assert.Empty(t, response.Results)
}

func TestToMirrorsUseCase_Execute_MultipleMirrors(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockProvider := &SharedMockRepositoryProvider{}
	mockGitOps := &SharedMockGitOperations{}
	mockLogger := &SharedMockLogger{}

	// Create a mock repository
	mockRepo := &SharedMockGitRepository{}
	mockRepo.On("Name").Return("test-repo")
	mockRepo.On("Path").Return("/tmp/test-repo")
	mockRepo.On("URL").Return("https://github.com/user/test-repo.git")

	// Create multiple mirror targets
	gitlabMirror := entities.NewMirrorTarget(
		"gitlab-mirror",
		entities.ProviderTypeGitLab,
		"gitlab.com",
		"backup-user",
		"",
		entities.AuthConfig{},
		true,
	)

	giteaMirror := entities.NewMirrorTarget(
		"gitea-mirror",
		entities.ProviderTypeGitea,
		"gitea.com",
		"archive-user",
		"",
		entities.AuthConfig{},
		true,
	)

	request := ToMirrorsRequest{
		SourceRepositories: []ports.GitRepository{mockRepo},
		MirrorTargets:      []entities.MirrorTarget{gitlabMirror, giteaMirror},
		SourceConfig: ports.ProviderConfig{
			ProviderType: "github",
			Domain:       "github.com",
			Owner:        "user",
		},
		DryRun: false,
		Options: Options{
			CreateIfNotExists: true,
		},
	}

	// Setup expectations for both mirrors
	mockProvider.On("ProjectExists", mock.Anything, "backup-user", "test-repo").
		Return(false, "", nil).Once()
	mockProvider.On("CreateRepositoryForPush", mock.Anything, mock.MatchedBy(func(req ports.CreateRepositoryRequest) bool {
		return req.Name == "test-repo"
	})).Return("gitlab-project-123", nil).Once()

	mockProvider.On("ProjectExists", mock.Anything, "archive-user", "test-repo").
		Return(false, "", nil).Once()
	mockProvider.On("CreateRepositoryForPush", mock.Anything, mock.MatchedBy(func(req ports.CreateRepositoryRequest) bool {
		return req.Name == "test-repo"
	})).Return("gitea-project-456", nil).Once()

	// Setup logger (lenient for all log levels)
	mockLogger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()

	// Create filesystem and string utils
	fs := filesystem.NewOSFileSystem()
	stringUtils := sharedadapters.NewStringUtilsAdapter()

	// Create use case
	useCase := NewToMirrorsUseCase(
		mockProvider,
		mockGitOps,
		nil,
		fs,
		mockLogger,
		stringUtils,
	)

	// Execute
	ctx := context.Background()
	response, err := useCase.Execute(ctx, request)

	// Assertions
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, 1, response.TotalRepositories)
	assert.Equal(t, 2, response.SuccessfulSyncs) // Both mirrors succeeded
	assert.Equal(t, 0, response.FailedSyncs)
	assert.Len(t, response.Results, 2)

	// Check both results
	for _, result := range response.Results {
		assert.Equal(t, "test-repo", result.RepositoryName)
		assert.True(t, result.Success)
		assert.False(t, result.Skipped)
		assert.NoError(t, result.Error)
		assert.Equal(t, "created", result.Action)
	}

	// Verify mock expectations
	mockProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
