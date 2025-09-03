// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/generated/mocks/mockhexagonal"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
)

// mustBuildRepository is a test helper to create a repository for testing.
func mustBuildRepository(t *testing.T, url string) entities.Repository {
	t.Helper()

	builder := entities.NewRepositoryBuilder()
	builder, err := builder.WithName("test-repo")
	require.NoError(t, err)
	builder, err = builder.WithHTTPSURL(url)
	require.NoError(t, err)
	repo, err := builder.Build()
	require.NoError(t, err)

	return repo
}

func TestBranchProtectionUseCase_ExecuteProtection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		request       sync.ProtectionRequest
		setupMocks    func(*mockhexagonal.RepositoryProvider, *mockhexagonal.Logger)
		expectedResp  sync.ProtectionResponse
		expectedErr   bool
		errorContains string
	}{
		{
			name: "enable protection successfully",
			request: sync.ProtectionRequest{
				ProviderConfig: ports.ProviderConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
				},
				Repository: mustBuildRepository(t, "https://github.com/testuser/test-repo.git"),
				Branch:     "main",
				Protection: ports.BranchProtection{
					Protected:     true,
					EnforceAdmins: true,
					RequiredPullRequestReviews: ports.RequiredPullRequestReviews{
						RequiredApprovingReviewCount: 2,
						DismissStaleReviews:          true,
					},
				},
				Operation: sync.ProtectionOperationEnable,
			},
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
				mockLogger.On("Debug", mock.Anything, "Enabling branch protection", mock.Anything).Once()
				mockProvider.On("SupportsFeature", ports.FeatureBranchProtection).Return(true).Once()
				mockProvider.On("SetBranchProtection", mock.Anything, mock.Anything, "test-repo", "main", mock.Anything).Return(nil).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection enabled successfully", mock.Anything).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()
			},
			expectedResp: sync.ProtectionResponse{
				Success:    true,
				Repository: "test-repo",
				Branch:     "main",
				Operation:  sync.ProtectionOperationEnable,
				Protected:  true,
			},
			expectedErr: false,
		},
		{
			name: "disable protection successfully",
			request: sync.ProtectionRequest{
				ProviderConfig: ports.ProviderConfig{
					ProviderType: "gitlab",
					Domain:       "gitlab.com",
					Owner:        "testuser",
				},
				Repository: mustBuildRepository(t, "https://gitlab.com/testuser/test-repo.git"),
				Branch:     "main",
				Operation:  sync.ProtectionOperationDisable,
			},
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
				mockLogger.On("Debug", mock.Anything, "Disabling branch protection", mock.Anything).Once()
				mockProvider.On("SupportsFeature", ports.FeatureBranchProtection).Return(true).Once()
				mockProvider.On("RemoveBranchProtection", mock.Anything, mock.Anything, "test-repo", "main").Return(nil).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection disabled successfully", mock.Anything).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()
			},
			expectedResp: sync.ProtectionResponse{
				Success:    true,
				Repository: "test-repo",
				Branch:     "main",
				Operation:  sync.ProtectionOperationDisable,
				Protected:  false,
			},
			expectedErr: false,
		},
		{
			name: "update protection successfully",
			request: sync.ProtectionRequest{
				ProviderConfig: ports.ProviderConfig{
					ProviderType: "gitea",
					Domain:       "gitea.com",
					Owner:        "testuser",
				},
				Repository: mustBuildRepository(t, "https://gitea.com/testuser/test-repo.git"),
				Branch:     "develop",
				Protection: ports.BranchProtection{
					Protected:             true,
					RequiredLinearHistory: true,
					AllowForcePushes:      false,
				},
				Operation: sync.ProtectionOperationUpdate,
			},
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
				mockLogger.On("Debug", mock.Anything, "Updating branch protection", mock.Anything).Once()
				mockProvider.On("SupportsFeature", ports.FeatureBranchProtection).Return(true).Once()
				mockProvider.On("SetBranchProtection", mock.Anything, mock.Anything, "test-repo", "develop", mock.Anything).Return(nil).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection updated successfully", mock.Anything).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()
			},
			expectedResp: sync.ProtectionResponse{
				Success:    true,
				Repository: "test-repo",
				Branch:     "develop",
				Operation:  sync.ProtectionOperationUpdate,
				Protected:  true,
			},
			expectedErr: false,
		},
		{
			name: "provider does not support branch protection",
			request: sync.ProtectionRequest{
				ProviderConfig: ports.ProviderConfig{
					ProviderType: "unsupported",
					Domain:       "example.com",
					Owner:        "testuser",
				},
				Repository: mustBuildRepository(t, "https://example.com/testuser/test-repo.git"),
				Branch:     "main",
				Operation:  sync.ProtectionOperationEnable,
			},
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
				mockLogger.On("Debug", mock.Anything, "Enabling branch protection", mock.Anything).Once()
				mockProvider.On("SupportsFeature", ports.FeatureBranchProtection).Return(false).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()
			},
			expectedResp: sync.ProtectionResponse{
				Repository: "test-repo",
				Branch:     "main",
				Operation:  sync.ProtectionOperationEnable,
				Success:    false,
				Protected:  false,
			},
			expectedErr:   true,
			errorContains: "provider does not support branch protection",
		},
		{
			name: "unknown protection operation",
			request: sync.ProtectionRequest{
				ProviderConfig: ports.ProviderConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
				},
				Repository: mustBuildRepository(t, "https://github.com/testuser/test-repo.git"),
				Branch:     "main",
				Operation:  "invalid",
			},
			setupMocks: func(_ *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()
			},
			expectedResp: sync.ProtectionResponse{
				Repository: "test-repo",
				Branch:     "main",
				Operation:  "invalid",
				Success:    false,
				Protected:  false,
			},
			expectedErr:   true,
			errorContains: "unknown protection operation",
		},
		{
			name: "enable protection fails with API error",
			request: sync.ProtectionRequest{
				ProviderConfig: ports.ProviderConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
				},
				Repository: mustBuildRepository(t, "https://github.com/testuser/test-repo.git"),
				Branch:     "main",
				Protection: ports.BranchProtection{
					Protected: true,
				},
				Operation: sync.ProtectionOperationEnable,
			},
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
				mockLogger.On("Debug", mock.Anything, "Enabling branch protection", mock.Anything).Once()
				mockProvider.On("SupportsFeature", ports.FeatureBranchProtection).Return(true).Once()
				mockProvider.On("SetBranchProtection", mock.Anything, mock.Anything, "test-repo", "main", mock.Anything).
					Return(errors.New("API rate limit exceeded")).Once()
				mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()
			},
			expectedResp: sync.ProtectionResponse{
				Repository: "test-repo",
				Branch:     "main",
				Operation:  sync.ProtectionOperationEnable,
				Success:    false,
				Protected:  false,
			},
			expectedErr:   true,
			errorContains: "failed to enable protection",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockProvider := mockhexagonal.NewRepositoryProvider(t)
			mockLogger := mockhexagonal.NewLogger(t)

			// Configure mocks
			testCase.setupMocks(mockProvider, mockLogger)

			// Create use case
			useCase := sync.NewBranchProtectionUseCase(mockProvider, mockLogger)

			// Execute
			ctx := context.Background()
			resp, err := useCase.ExecuteProtection(ctx, testCase.request)

			// Assert
			if testCase.expectedErr {
				require.Error(t, err)

				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, testCase.expectedResp.Success, resp.Success)
			assert.Equal(t, testCase.expectedResp.Repository, resp.Repository)
			assert.Equal(t, testCase.expectedResp.Branch, resp.Branch)
			assert.Equal(t, testCase.expectedResp.Operation, resp.Operation)
			assert.Equal(t, testCase.expectedResp.Protected, resp.Protected)

			// Verify all expectations were met
			mockProvider.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestBranchProtectionUseCase_GetProtectionStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		providerConfig ports.ProviderConfig
		repoName       string
		branch         string
		setupMocks     func(*mockhexagonal.RepositoryProvider, *mockhexagonal.Logger)
		expectedProt   ports.BranchProtection
		expectedErr    bool
		errorContains  string
	}{
		{
			name: "get protection status successfully",
			providerConfig: ports.ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com",
				Owner:        "testuser",
			},
			repoName: "test-repo",
			branch:   "main",
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Debug", mock.Anything, "Getting branch protection status", mock.Anything).Once()
				expectedProtection := ports.BranchProtection{
					Protected:     true,
					EnforceAdmins: true,
					RequiredPullRequestReviews: ports.RequiredPullRequestReviews{
						RequiredApprovingReviewCount: 2,
					},
				}
				mockProvider.On("GetBranchProtection", mock.Anything, mock.Anything, "test-repo", "main").
					Return(expectedProtection, nil).Once()
			},
			expectedProt: ports.BranchProtection{
				Protected:     true,
				EnforceAdmins: true,
				RequiredPullRequestReviews: ports.RequiredPullRequestReviews{
					RequiredApprovingReviewCount: 2,
				},
			},
			expectedErr: false,
		},
		{
			name: "get protection status for unprotected branch",
			providerConfig: ports.ProviderConfig{
				ProviderType: "gitlab",
				Domain:       "gitlab.com",
				Owner:        "testuser",
			},
			repoName: "test-repo",
			branch:   "develop",
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Debug", mock.Anything, "Getting branch protection status", mock.Anything).Once()
				mockProvider.On("GetBranchProtection", mock.Anything, mock.Anything, "test-repo", "develop").
					Return(ports.BranchProtection{Protected: false}, nil).Once()
			},
			expectedProt: ports.BranchProtection{Protected: false},
			expectedErr:  false,
		},
		{
			name: "get protection status fails with error",
			providerConfig: ports.ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com",
				Owner:        "testuser",
			},
			repoName: "test-repo",
			branch:   "main",
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Debug", mock.Anything, "Getting branch protection status", mock.Anything).Once()
				mockProvider.On("GetBranchProtection", mock.Anything, mock.Anything, "test-repo", "main").
					Return(ports.BranchProtection{}, errors.New("branch not found")).Once()
			},
			expectedProt:  ports.BranchProtection{},
			expectedErr:   true,
			errorContains: "failed to get branch protection",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockProvider := mockhexagonal.NewRepositoryProvider(t)
			mockLogger := mockhexagonal.NewLogger(t)

			// Configure mocks
			testCase.setupMocks(mockProvider, mockLogger)

			// Create use case
			useCase := sync.NewBranchProtectionUseCase(mockProvider, mockLogger)

			// Execute
			ctx := context.Background()
			protection, err := useCase.GetProtectionStatus(ctx, testCase.providerConfig, testCase.repoName, testCase.branch)

			// Assert
			if testCase.expectedErr {
				require.Error(t, err)

				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedProt, protection)
			}

			// Verify all expectations were met
			mockProvider.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestBranchProtectionUseCase_ListProtectedBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		providerConfig   ports.ProviderConfig
		repoName         string
		setupMocks       func(*mockhexagonal.RepositoryProvider, *mockhexagonal.Logger)
		expectedBranches []string
		expectedErr      bool
		errorContains    string
	}{
		{
			name: "list protected branches successfully",
			providerConfig: ports.ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com",
				Owner:        "testuser",
			},
			repoName: "test-repo",
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Debug", mock.Anything, "Listing protected branches", mock.Anything).Once()
				mockProvider.On("ListProtectedBranches", mock.Anything, mock.Anything, "test-repo").
					Return([]string{"main", "develop", "release"}, nil).Once()
			},
			expectedBranches: []string{"main", "develop", "release"},
			expectedErr:      false,
		},
		{
			name: "list protected branches returns empty list",
			providerConfig: ports.ProviderConfig{
				ProviderType: "gitlab",
				Domain:       "gitlab.com",
				Owner:        "testuser",
			},
			repoName: "test-repo",
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Debug", mock.Anything, "Listing protected branches", mock.Anything).Once()
				mockProvider.On("ListProtectedBranches", mock.Anything, mock.Anything, "test-repo").
					Return([]string{}, nil).Once()
			},
			expectedBranches: []string{},
			expectedErr:      false,
		},
		{
			name: "list protected branches fails with error",
			providerConfig: ports.ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com",
				Owner:        "testuser",
			},
			repoName: "test-repo",
			setupMocks: func(mockProvider *mockhexagonal.RepositoryProvider, mockLogger *mockhexagonal.Logger) {
				mockLogger.On("Debug", mock.Anything, "Listing protected branches", mock.Anything).Once()
				mockProvider.On("ListProtectedBranches", mock.Anything, mock.Anything, "test-repo").
					Return(nil, errors.New("repository not found")).Once()
			},
			expectedBranches: nil,
			expectedErr:      true,
			errorContains:    "failed to list protected branches",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockProvider := mockhexagonal.NewRepositoryProvider(t)
			mockLogger := mockhexagonal.NewLogger(t)

			// Configure mocks
			testCase.setupMocks(mockProvider, mockLogger)

			// Create use case
			useCase := sync.NewBranchProtectionUseCase(mockProvider, mockLogger)

			// Execute
			ctx := context.Background()
			branches, err := useCase.ListProtectedBranches(ctx, testCase.providerConfig, testCase.repoName)

			// Assert
			if testCase.expectedErr {
				require.Error(t, err)

				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedBranches, branches)
			}

			// Verify all expectations were met
			mockProvider.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestBranchProtectionUseCase_CrossProviderCompatibility(t *testing.T) {
	t.Parallel()

	// Test that protection settings work across different providers
	providers := []string{"github", "gitlab", "gitea"}

	for _, provider := range providers {
		t.Run(provider+"_protection_cycle", func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mockProvider := mockhexagonal.NewRepositoryProvider(t)
			mockLogger := mockhexagonal.NewLogger(t)

			providerConfig := ports.ProviderConfig{
				ProviderType: provider,
				Domain:       provider + ".com",
				Owner:        "testuser",
			}

			protection := ports.BranchProtection{
				Protected:     true,
				EnforceAdmins: true,
				RequiredPullRequestReviews: ports.RequiredPullRequestReviews{
					RequiredApprovingReviewCount: 1,
				},
			}

			// Setup expectations for full protection lifecycle
			// 1. Enable protection
			mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
			mockLogger.On("Debug", mock.Anything, "Enabling branch protection", mock.Anything).Once()
			mockProvider.On("SupportsFeature", ports.FeatureBranchProtection).Return(true).Once()
			mockProvider.On("SetBranchProtection", mock.Anything, mock.Anything, "test-repo", "main", mock.Anything).Return(nil).Once()
			mockLogger.On("Info", mock.Anything, "Branch protection enabled successfully", mock.Anything).Once()
			mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()

			// 2. Get status
			mockLogger.On("Debug", mock.Anything, "Getting branch protection status", mock.Anything).Once()
			mockProvider.On("GetBranchProtection", mock.Anything, mock.Anything, "test-repo", "main").
				Return(protection, nil).Once()

			// 3. List protected branches
			mockLogger.On("Debug", mock.Anything, "Listing protected branches", mock.Anything).Once()
			mockProvider.On("ListProtectedBranches", mock.Anything, mock.Anything, "test-repo").
				Return([]string{"main"}, nil).Once()

			// 4. Disable protection
			mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
			mockLogger.On("Debug", mock.Anything, "Disabling branch protection", mock.Anything).Once()
			mockProvider.On("SupportsFeature", ports.FeatureBranchProtection).Return(true).Once()
			mockProvider.On("RemoveBranchProtection", mock.Anything, mock.Anything, "test-repo", "main").Return(nil).Once()
			mockLogger.On("Info", mock.Anything, "Branch protection disabled successfully", mock.Anything).Once()
			mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()

			// Create use case
			useCase := sync.NewBranchProtectionUseCase(mockProvider, mockLogger)
			ctx := context.Background()

			// 1. Enable protection
			enableReq := sync.ProtectionRequest{
				ProviderConfig: providerConfig,
				Repository:     mustBuildRepository(t, "https://"+provider+".com/testuser/test-repo.git"),
				Branch:         "main",
				Protection:     protection,
				Operation:      sync.ProtectionOperationEnable,
			}
			resp, err := useCase.ExecuteProtection(ctx, enableReq)
			require.NoError(t, err)
			assert.True(t, resp.Success)
			assert.True(t, resp.Protected)

			// 2. Get protection status
			status, err := useCase.GetProtectionStatus(ctx, providerConfig, "test-repo", "main")
			require.NoError(t, err)
			assert.True(t, status.Protected)

			// 3. List protected branches
			branches, err := useCase.ListProtectedBranches(ctx, providerConfig, "test-repo")
			require.NoError(t, err)
			assert.Contains(t, branches, "main")

			// 4. Disable protection
			disableReq := sync.ProtectionRequest{
				ProviderConfig: providerConfig,
				Repository:     mustBuildRepository(t, "https://"+provider+".com/testuser/test-repo.git"),
				Branch:         "main",
				Operation:      sync.ProtectionOperationDisable,
			}
			resp, err = useCase.ExecuteProtection(ctx, disableReq)
			require.NoError(t, err)
			assert.True(t, resp.Success)
			assert.False(t, resp.Protected)

			// Verify all expectations were met
			mockProvider.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestBranchProtectionUseCase_ProviderSpecificFeatures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		provider    string
		protection  ports.BranchProtection
		description string
	}{
		{
			name:     "github_advanced_features",
			provider: "github",
			protection: ports.BranchProtection{
				Protected:                      true,
				RequiredLinearHistory:          true,
				RequiredConversationResolution: true,
				RequiredStatusChecks: ports.RequiredStatusChecks{
					Strict:   true,
					Contexts: []string{"ci/build", "ci/test"},
				},
			},
			description: "GitHub-specific advanced protection features",
		},
		{
			name:     "gitlab_merge_restrictions",
			provider: "gitlab",
			protection: ports.BranchProtection{
				Protected:        true,
				AllowForcePushes: false,
				AllowDeletions:   false,
				RequiredPullRequestReviews: ports.RequiredPullRequestReviews{
					RequiredApprovingReviewCount: 3,
					RequireCodeOwnerReviews:      true,
				},
			},
			description: "GitLab-specific merge request restrictions",
		},
		{
			name:     "gitea_basic_protection",
			provider: "gitea",
			protection: ports.BranchProtection{
				Protected:     true,
				EnforceAdmins: false,
				Restrictions: ports.BranchRestrictions{
					Users: []string{"admin", "maintainer"},
				},
			},
			description: "Gitea basic protection with user restrictions",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockProvider := mockhexagonal.NewRepositoryProvider(t)
			mockLogger := mockhexagonal.NewLogger(t)

			// Configure provider config
			providerConfig := ports.ProviderConfig{
				ProviderType: testCase.provider,
				Domain:       testCase.provider + ".com",
				Owner:        "testuser",
			}

			// Setup expectations
			mockLogger.On("Info", mock.Anything, "Executing branch protection operation", mock.Anything).Once()
			mockLogger.On("Debug", mock.Anything, "Enabling branch protection", mock.Anything).Once()
			mockProvider.On("SupportsFeature", ports.FeatureBranchProtection).Return(true).Once()
			mockProvider.On("SetBranchProtection", mock.Anything, mock.Anything, "test-repo", "main", testCase.protection).Return(nil).Once()
			mockLogger.On("Info", mock.Anything, "Branch protection enabled successfully", mock.Anything).Once()
			mockLogger.On("Info", mock.Anything, "Branch protection operation completed", mock.Anything).Once()

			// Create use case
			useCase := sync.NewBranchProtectionUseCase(mockProvider, mockLogger)

			// Execute
			ctx := context.Background()
			req := sync.ProtectionRequest{
				ProviderConfig: providerConfig,
				Repository:     mustBuildRepository(t, "https://"+testCase.provider+".com/testuser/test-repo.git"),
				Branch:         "main",
				Protection:     testCase.protection,
				Operation:      sync.ProtectionOperationEnable,
			}

			resp, err := useCase.ExecuteProtection(ctx, req)

			// Assert
			require.NoError(t, err)
			assert.True(t, resp.Success)
			assert.True(t, resp.Protected)
			assert.Equal(t, "main", resp.Branch)

			// Verify all expectations were met
			mockProvider.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
