// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestPushToProviderUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*SharedMockRepositoryProvider, *SharedMockGitRepository, *SharedMockGitOperations, *SharedMockLogger)
		request        PushRequest
		expectedError  bool
		expectedResult func(PushResponse) bool
	}{
		{
			name: "successful_push_to_existing_repository",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitRepo *SharedMockGitRepository, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				// Mock GPSUPSTREAM remote setup
				gitRepo.On("Name").Return("test-repo")
				gitRepo.On("Path").Return("/tmp/test-repo")
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/test-repo.git"},
				}, nil).Once()
				gitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(errors.New("not found")).Once() // Expected
				gitRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", "https://github.com/source/test-repo.git").Return(nil).Once()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/test-repo.git"},
					{Name: "GPSUPSTREAM", URL: "https://github.com/source/test-repo.git"},
				}, nil).Once()

				// Mock repository existence check
				provider.On("ProjectExists", mock.Anything, "target-owner", "test-repo").Return(true, "project-123", nil)

				// Mock push operation
				gitRepo.On("Push", mock.Anything, mock.AnythingOfType("ports.PushOptions")).Return(nil)

				// Mock logger calls (lenient)
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: PushRequest{
				SourceRepository: createTestRepository("test-repo"),
				SourceGitRepo:    nil, // Will be set by mock
				TargetConfig:     createTestMirrorTarget("target-owner"),
				CreateIfMissing:  true,
				DryRun:           false,
			},
			expectedError: false,
			expectedResult: func(resp PushResponse) bool {
				return resp.Success && !resp.Created && resp.ProjectID == "project-123"
			},
		},
		{
			name: "successful_creation_and_push_to_new_repository",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitRepo *SharedMockGitRepository, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				// Mock GPSUPSTREAM remote setup
				gitRepo.On("Name").Return("new-repo")
				gitRepo.On("Path").Return("/tmp/new-repo")
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/new-repo.git"},
				}, nil).Once()
				gitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(errors.New("not found")).Once()
				gitRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", "https://github.com/source/new-repo.git").Return(nil).Once()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/new-repo.git"},
					{Name: "GPSUPSTREAM", URL: "https://github.com/source/new-repo.git"},
				}, nil).Once()

				// Mock repository doesn't exist, needs creation
				provider.On("ProjectExists", mock.Anything, "target-owner", "new-repo").Return(false, "", nil)
				provider.On("CreateRepositoryForPush", mock.Anything, mock.AnythingOfType("ports.CreateRepositoryRequest")).Return("new-project-456", nil)

				// Mock push operation
				gitRepo.On("Push", mock.Anything, mock.AnythingOfType("ports.PushOptions")).Return(nil)

				// Mock logger calls (lenient)
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: PushRequest{
				SourceRepository: createTestRepository("new-repo"),
				SourceGitRepo:    nil, // Will be set by mock
				TargetConfig:     createTestMirrorTarget("target-owner"),
				CreateIfMissing:  true,
				DryRun:           false,
			},
			expectedError: false,
			expectedResult: func(resp PushResponse) bool {
				return resp.Success && resp.Created && resp.ProjectID == "new-project-456"
			},
		},
		{
			name: "dry_run_mode_simulation",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitRepo *SharedMockGitRepository, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				// No mocks needed for dry run - it should simulate everything
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: PushRequest{
				SourceRepository: createTestRepository("dry-run-repo"),
				SourceGitRepo:    nil,
				TargetConfig:     createTestMirrorTarget("target-owner"),
				DryRun:           true,
			},
			expectedError: false,
			expectedResult: func(resp PushResponse) bool {
				return resp.Success && !resp.Created && resp.ProjectID == "dry-run-project-id"
			},
		},
		{
			name: "push_failure_during_git_operation",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitRepo *SharedMockGitRepository, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				// Mock GPSUPSTREAM remote setup
				gitRepo.On("Name").Return("fail-repo")
				gitRepo.On("Path").Return("/tmp/fail-repo")
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/fail-repo.git"},
				}, nil).Once()
				gitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(errors.New("not found")).Once()
				gitRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", "https://github.com/source/fail-repo.git").Return(nil).Once()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/fail-repo.git"},
					{Name: "GPSUPSTREAM", URL: "https://github.com/source/fail-repo.git"},
				}, nil).Once()

				// Mock repository exists
				provider.On("ProjectExists", mock.Anything, "target-owner", "fail-repo").Return(true, "project-fail", nil).Once()

				// Mock push operation failure
				gitRepo.On("Push", mock.Anything, mock.AnythingOfType("ports.PushOptions")).Return(errors.New("push failed: authentication failed")).Once()

				// Mock logger calls (lenient)
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: PushRequest{
				SourceRepository: createTestRepository("fail-repo"),
				SourceGitRepo:    nil,
				TargetConfig:     createTestMirrorTarget("target-owner"),
				CreateIfMissing:  true,
				DryRun:           false,
			},
			expectedError: true,
			expectedResult: func(resp PushResponse) bool {
				return !resp.Success && resp.Error != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockProvider := &SharedMockRepositoryProvider{}
			mockGitRepo := &SharedMockGitRepository{}
			mockGitOps := &SharedMockGitOperations{}
			mockLogger := &SharedMockLogger{}

			// Setup test-specific mocks
			tt.setupMocks(mockProvider, mockGitRepo, mockGitOps, mockLogger)

			// Set the git repo in request
			tt.request.SourceGitRepo = mockGitRepo

			// Create use case
			useCase := NewPushToProviderUseCase(mockProvider, mockGitOps, mockLogger)

			// Execute
			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.request)

			// Verify error expectation
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify result expectations
			if tt.expectedResult != nil {
				require.True(t, tt.expectedResult(result), "Result validation failed")
			}

			// Verify all mocks were called as expected
			mockProvider.AssertExpectations(t)
			mockGitRepo.AssertExpectations(t)
			mockGitOps.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestPushToProviderUseCase_setupGPSUpstreamRemote(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*SharedMockGitRepository)
		expectError bool
	}{
		{
			name: "successful_remote_setup",
			setupMock: func(gitRepo *SharedMockGitRepository) {
				gitRepo.On("Name").Return("test-repo").Maybe()
				gitRepo.On("Path").Return("/tmp/test-repo").Maybe()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/test/repo.git"},
				}, nil).Once()
				gitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(errors.New("not found")).Once()
				gitRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", "https://github.com/test/repo.git").Return(nil).Once()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/test/repo.git"},
					{Name: "GPSUPSTREAM", URL: "https://github.com/test/repo.git"},
				}, nil).Once()
			},
			expectError: false,
		},
		{
			name: "missing_origin_remote",
			setupMock: func(gitRepo *SharedMockGitRepository) {
				gitRepo.On("Name").Return("test-repo").Maybe()
				gitRepo.On("Path").Return("/tmp/test-repo").Maybe()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{}, nil).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGitRepo := &SharedMockGitRepository{}
			mockProvider := &SharedMockRepositoryProvider{}
			mockGitOps := &SharedMockGitOperations{}
			mockLogger := &SharedMockLogger{}

			tt.setupMock(mockGitRepo)
			mockLogger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			mockLogger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()

			useCase := NewPushToProviderUseCase(mockProvider, mockGitOps, mockLogger)
			err := useCase.setupGPSUpstreamRemote(context.Background(), mockGitRepo)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockGitRepo.AssertExpectations(t)
		})
	}
}
