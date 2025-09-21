// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestToMirrorsUseCase_EmptyRequest(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockProvider := &SharedMockRepositoryProvider{}
	mockGitOps := &SharedMockGitOperations{}
	mockLogger := &SharedMockLogger{}

	// Setup logger
	mockLogger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()

	// Create filesystem
	fs := filesystem.NewOSFileSystem() //nolint:varnamelen // Common abbreviation for filesystem

	// Create use case
	useCase := NewToMirrorsUseCase(
		mockProvider,
		mockGitOps,
		nil,
		fs,
		mockLogger,
	)

	// Execute with empty request
	ctx := context.Background()
	request := ToMirrorsRequest{
		SourceRepositories: []ports.GitRepository{},
		MirrorTargets:      []entities.MirrorTarget{},
	}

	response, err := useCase.Execute(ctx, request)

	// Should handle gracefully
	require.NoError(t, err)
	assert.Equal(t, 0, response.TotalRepositories)
	assert.Equal(t, 0, response.SuccessfulSyncs)
	assert.Empty(t, response.Results)
}

func TestToMirrorsUseCase_NoMirrorTargets(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockProvider := &SharedMockRepositoryProvider{}
	mockGitOps := &SharedMockGitOperations{}
	mockLogger := &SharedMockLogger{}

	// Create a mock repository
	mockRepo := &SharedMockGitRepository{}
	mockRepo.On("Name").Return("test-repo").Maybe()

	// Setup logger
	mockLogger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()

	// Create filesystem
	fs := filesystem.NewOSFileSystem() //nolint:varnamelen // Common abbreviation for filesystem

	// Create use case
	useCase := NewToMirrorsUseCase(
		mockProvider,
		mockGitOps,
		nil,
		fs,
		mockLogger,
	)

	// Execute with no mirror targets
	ctx := context.Background()
	request := ToMirrorsRequest{
		SourceRepositories: []ports.GitRepository{mockRepo},
		MirrorTargets:      []entities.MirrorTarget{}, // No targets
		SourceConfig: ports.ProviderConfig{
			ProviderType: "github",
			Domain:       "github.com",
			Owner:        "user",
		},
	}

	response, err := useCase.Execute(ctx, request)

	// Should handle gracefully
	require.NoError(t, err)
	assert.Equal(t, 1, response.TotalRepositories)
	assert.Equal(t, 0, response.SuccessfulSyncs)
	assert.Equal(t, 0, response.FailedSyncs)
}
