// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"itiquette/git-provider-sync/internal/adapters/shared"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

//nolint:cyclop // Test function with multiple test cases
func TestPushToProviderUseCase_Execute(t *testing.T) {
	t.Parallel()

	// Create isolated temp directory for all test cases
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		setupMocks     func(*SharedMockRepositoryProvider, *SharedMockGitRepository, *SharedMockGitOperations)
		request        PushRequest
		expectedError  bool
		expectedResult func(PushResponse) bool
	}{
		{
			name: "push to existing repository succeeds",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitRepo *SharedMockGitRepository, _ *SharedMockGitOperations) {
				// Mock GPSUPSTREAM remote setup
				gitRepo.On("Name").Return("test-repo")
				gitRepo.On("Path").Return(filepath.Join(tmpDir, "test-repo"))
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/test-repo.git"},
				}, nil).Once()
				gitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(domain.ErrTestNotFound).Once() // Expected
				gitRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", "https://github.com/source/test-repo.git").Return(nil).Once()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/test-repo.git"},
					{Name: "GPSUPSTREAM", URL: "https://github.com/source/test-repo.git"},
				}, nil).Once()

				// Mock repository existence check
				provider.On("ProjectExists", mock.Anything, "target-owner", "test-repo").Return(true, "project-123", nil)

				// Mock remote URL update (CRITICAL: Our fix for GitHub → GitLab sync)
				gitRepo.On("UpdateRemote", mock.Anything, "origin", mock.AnythingOfType("string")).Return(nil).Once()

				// Mock push operation
				gitRepo.On("Push", mock.Anything, mock.AnythingOfType("ports.PushOptions")).Return(nil)

				// Mock logger calls (lenient)
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
			name: "creates and pushes to new repository when missing",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitRepo *SharedMockGitRepository, _ *SharedMockGitOperations) {
				// Mock GPSUPSTREAM remote setup
				gitRepo.On("Name").Return("new-repo")
				gitRepo.On("Path").Return(filepath.Join(tmpDir, "new-repo"))
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/new-repo.git"},
				}, nil).Once()
				gitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(domain.ErrTestNotFound).Once()
				gitRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", "https://github.com/source/new-repo.git").Return(nil).Once()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/new-repo.git"},
					{Name: "GPSUPSTREAM", URL: "https://github.com/source/new-repo.git"},
				}, nil).Once()

				// Mock repository doesn't exist, needs creation
				provider.On("ProjectExists", mock.Anything, "target-owner", "new-repo").Return(false, "", nil)
				provider.On("CreateRepositoryForPush", mock.Anything, mock.AnythingOfType("ports.CreateRepositoryRequest")).Return("new-project-456", nil)

				// Mock remote URL update (CRITICAL: Our fix for GitHub → GitLab sync)
				gitRepo.On("UpdateRemote", mock.Anything, "origin", mock.AnythingOfType("string")).Return(nil).Once()

				// Mock push operation
				gitRepo.On("Push", mock.Anything, mock.AnythingOfType("ports.PushOptions")).Return(nil)

				// Mock logger calls (lenient)
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
			name: "simulates push when dry run is enabled",
			setupMocks: func(_ *SharedMockRepositoryProvider, _ *SharedMockGitRepository, _ *SharedMockGitOperations) {
				// No mocks needed for dry run - it should simulate everything
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
			name: "returns error when push operation fails",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitRepo *SharedMockGitRepository, _ *SharedMockGitOperations) {
				// Mock GPSUPSTREAM remote setup
				gitRepo.On("Name").Return("fail-repo")
				gitRepo.On("Path").Return(filepath.Join(tmpDir, "fail-repo"))
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/fail-repo.git"},
				}, nil).Once()
				gitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(domain.ErrTestNotFound).Once()
				gitRepo.On("AddRemote", mock.Anything, "GPSUPSTREAM", "https://github.com/source/fail-repo.git").Return(nil).Once()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/source/fail-repo.git"},
					{Name: "GPSUPSTREAM", URL: "https://github.com/source/fail-repo.git"},
				}, nil).Once()

				// Mock repository exists
				provider.On("ProjectExists", mock.Anything, "target-owner", "fail-repo").Return(true, "project-fail", nil).Once()

				// Mock remote URL update (CRITICAL: Our fix for GitHub → GitLab sync)
				gitRepo.On("UpdateRemote", mock.Anything, "origin", mock.AnythingOfType("string")).Return(nil).Once()

				// Mock push operation failure
				gitRepo.On("Push", mock.Anything, mock.AnythingOfType("ports.PushOptions")).Return(domain.ErrTestAuthenticationFailed).Once()

				// Mock logger calls (lenient)
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Create mocks
			mockProvider := &SharedMockRepositoryProvider{}
			mockGitRepo := &SharedMockGitRepository{}
			mockGitOps := &SharedMockGitOperations{}

			// Setup test-specific mocks
			test.setupMocks(mockProvider, mockGitRepo, mockGitOps)

			// Set the git repo in request
			test.request.SourceGitRepo = mockGitRepo

			// Create use case
			useCase := NewPushToProviderUseCase(mockProvider, mockGitOps, shared.NewStringUtilsAdapter())

			// Execute
			ctx := context.Background()
			result, err := useCase.Execute(ctx, test.request)

			// Verify error expectation
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify result expectations
			if test.expectedResult != nil {
				require.True(t, test.expectedResult(result), "Result validation failed")
			}

			// Verify all mocks were called as expected
			mockProvider.AssertExpectations(t)
			mockGitRepo.AssertExpectations(t)
			mockGitOps.AssertExpectations(t)
		})
	}
}

func TestPushToProviderUseCase_setupGPSUpstreamRemote(t *testing.T) {
	t.Parallel()

	// Create isolated temp directory for all test cases
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setupMock   func(*SharedMockGitRepository)
		expectError bool
	}{
		{
			name: "successful_remote_setup",
			setupMock: func(gitRepo *SharedMockGitRepository) {
				gitRepo.On("Name").Return("test-repo").Maybe()
				gitRepo.On("Path").Return(filepath.Join(tmpDir, "test-repo")).Maybe()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{
					{Name: "origin", URL: "https://github.com/test/repo.git"},
				}, nil).Once()
				gitRepo.On("RemoveRemote", mock.Anything, "GPSUPSTREAM").Return(domain.ErrTestNotFound).Once()
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
				gitRepo.On("Path").Return(filepath.Join(tmpDir, "test-repo")).Maybe()
				gitRepo.On("ListRemotes", mock.Anything).Return([]ports.RemoteInfo{}, nil).Once()
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockGitRepo := &SharedMockGitRepository{}
			mockProvider := &SharedMockRepositoryProvider{}
			mockGitOps := &SharedMockGitOperations{}

			test.setupMock(mockGitRepo)

			useCase := NewPushToProviderUseCase(mockProvider, mockGitOps, shared.NewStringUtilsAdapter())
			err := useCase.setupGPSUpstreamRemote(context.Background(), mockGitRepo)

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockGitRepo.AssertExpectations(t)
		})
	}
}

func TestPushToProviderUseCase_createAuthOptions(t *testing.T) {
	t.Parallel()

	useCase := &PushToProviderUseCase{}
	ctx := context.Background()

	tests := []struct {
		name         string
		authConfig   entities.AuthConfig
		expectedType ports.AuthType
		expectedAuth ports.AuthOptions
	}{
		{
			name:         "token authentication",
			authConfig:   entities.NewAuthConfigWithToken("test-token", ""),
			expectedType: ports.AuthTypeToken,
			expectedAuth: ports.AuthOptions{
				Type:     ports.AuthTypeToken,
				Token:    "test-token",
				Username: "git",
			},
		},
		{
			name:         "SSH key path authentication",
			authConfig:   entities.NewAuthConfigWithSSH("/path/to/key", ""),
			expectedType: ports.AuthTypeSSHKey,
			expectedAuth: ports.AuthOptions{
				Type:       ports.AuthTypeSSHKey,
				SSHKeyPath: "/path/to/key",
			},
		},
		{
			name:         "SSH key content authentication",
			authConfig:   entities.NewAuthConfigWithSSHKey("ssh-key-content", ""),
			expectedType: ports.AuthTypeSSHKey,
			expectedAuth: ports.AuthOptions{
				Type:   ports.AuthTypeSSHKey,
				SSHKey: []byte("ssh-key-content"),
			},
		},
		{
			name:         "no authentication",
			authConfig:   entities.NewAuthenticationConfig(entities.AuthTypeNone, "", "", "", ""),
			expectedType: ports.AuthTypeNone,
			expectedAuth: ports.AuthOptions{
				Type: ports.AuthTypeNone,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := useCase.createAuthOptions(ctx, testCase.authConfig)

			require.Equal(t, testCase.expectedAuth.Type, result.Type)
			require.Equal(t, testCase.expectedAuth.Token, result.Token)
			require.Equal(t, testCase.expectedAuth.Username, result.Username)
			require.Equal(t, testCase.expectedAuth.SSHKeyPath, result.SSHKeyPath)
			require.Equal(t, testCase.expectedAuth.SSHKey, result.SSHKey)
		})
	}
}
