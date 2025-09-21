// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// mockGitRepository provides a mock implementation of GitRepository for unit testing.
type mockGitRepository struct {
	mock.Mock
}

func (m *mockGitRepository) Name() string {
	args := m.Called()

	return args.String(0)
}

func (m *mockGitRepository) Path() string {
	args := m.Called()

	return args.String(0)
}

func (m *mockGitRepository) IsBare() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *mockGitRepository) ListRemotes(ctx context.Context) ([]ports.RemoteInfo, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, err //nolint:wrapcheck // Test mock
	}

	remotes, _ := args.Get(0).([]ports.RemoteInfo)

	return remotes, nil
}

func (m *mockGitRepository) AddRemote(ctx context.Context, name, url string) error {
	args := m.Called(ctx, name, url)

	return args.Error(0) //nolint:wrapcheck // Test mock
}

func (m *mockGitRepository) RemoveRemote(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0) //nolint:wrapcheck // Test mock
}

func (m *mockGitRepository) FetchRemote(ctx context.Context, name string, options ports.FetchOptions) error {
	args := m.Called(ctx, name, options)

	return args.Error(0) //nolint:wrapcheck // Test mock
}

func (m *mockGitRepository) PushToRemote(ctx context.Context, name string, options ports.PushOptions) error {
	args := m.Called(ctx, name, options)

	return args.Error(0) //nolint:wrapcheck // Test mock
}

func (m *mockGitRepository) GetDefaultBranch(ctx context.Context) (string, error) {
	args := m.Called(ctx)

	return args.String(0), args.Error(1)
}

func (m *mockGitRepository) SetDefaultBranch(ctx context.Context, branch string) error {
	args := m.Called(ctx, branch)

	return args.Error(0) //nolint:wrapcheck // Test mock
}

func (m *mockGitRepository) HasChanges(ctx context.Context) (bool, error) {
	args := m.Called(ctx)

	return args.Bool(0), args.Error(1)
}

func (m *mockGitRepository) GetLastCommitTime(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	timestamp, _ := args.Get(0).(int64)

	return timestamp, args.Error(1) //nolint:wrapcheck // Test mock
}

func (m *mockGitRepository) Close() error {
	args := m.Called()

	return args.Error(0) //nolint:wrapcheck // Test mock
}

// TestGitRepositoryOperations_AddAndListRemotes tests remote management without real git.
func TestGitRepository_RemoteManagement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialRemotes []ports.RemoteInfo
		addRemote      struct {
			name string
			url  string
			err  error
		}
		expectedRemotes []ports.RemoteInfo
		wantErr         error
	}{
		{
			name:           "successfully_adds_remote_to_empty_repository",
			initialRemotes: []ports.RemoteInfo{},
			addRemote: struct {
				name string
				url  string
				err  error
			}{
				name: "origin",
				url:  "https://github.com/test/repo.git",
				err:  nil,
			},
			expectedRemotes: []ports.RemoteInfo{
				{Name: "origin", URL: "https://github.com/test/repo.git"},
			},
			wantErr: nil,
		},
		{
			name: "successfully_adds_second_remote",
			initialRemotes: []ports.RemoteInfo{
				{Name: "origin", URL: "https://github.com/test/repo.git"},
			},
			addRemote: struct {
				name string
				url  string
				err  error
			}{
				name: "backup",
				url:  "https://gitlab.com/backup/repo.git",
				err:  nil,
			},
			expectedRemotes: []ports.RemoteInfo{
				{Name: "origin", URL: "https://github.com/test/repo.git"},
				{Name: "backup", URL: "https://gitlab.com/backup/repo.git"},
			},
			wantErr: nil,
		},
		{
			name: "handles_duplicate_remote_name_error",
			initialRemotes: []ports.RemoteInfo{
				{Name: "origin", URL: "https://github.com/test/repo.git"},
			},
			addRemote: struct {
				name string
				url  string
				err  error
			}{
				name: "origin",
				url:  "https://gitlab.com/other/repo.git",
				err:  errors.New("remote 'origin' already exists"),
			},
			expectedRemotes: []ports.RemoteInfo{
				{Name: "origin", URL: "https://github.com/test/repo.git"},
			},
			wantErr: errors.New("remote 'origin' already exists"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockRepo := new(mockGitRepository)

			// Setup initial state
			mockRepo.On("ListRemotes", ctx).Return(testCase.initialRemotes, nil).Once()

			// Setup add remote expectation if we're adding
			if testCase.addRemote.name != "" {
				mockRepo.On("AddRemote", ctx, testCase.addRemote.name, testCase.addRemote.url).Return(testCase.addRemote.err).Once()
			}

			// Setup final list remotes call
			if testCase.addRemote.err == nil && testCase.addRemote.name != "" {
				mockRepo.On("ListRemotes", ctx).Return(testCase.expectedRemotes, nil).Once()
			}

			// Test initial state
			remotes, err := mockRepo.ListRemotes(ctx)
			require.NoError(t, err)
			require.Equal(t, testCase.initialRemotes, remotes)

			// Test adding remote
			if testCase.addRemote.name != "" {
				err = mockRepo.AddRemote(ctx, testCase.addRemote.name, testCase.addRemote.url)
				if testCase.wantErr != nil {
					require.Error(t, err)
					require.Equal(t, testCase.wantErr.Error(), err.Error())
				} else {
					require.NoError(t, err)

					// Verify final state
					remotes, err = mockRepo.ListRemotes(ctx)
					require.NoError(t, err)
					require.Equal(t, testCase.expectedRemotes, remotes)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGitRepositoryProperties_ReturnsCorrectMetadata tests repository metadata without real git.
func TestGitRepository_MetadataAccuracy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repoName string
		repoPath string
		isBare   bool
		wantName string
		wantPath string
		wantBare bool
	}{
		{
			name:     "regular_repository_with_working_directory",
			repoName: "test-repo",
			repoPath: "/tmp/repos/test-repo",
			isBare:   false,
			wantName: "test-repo",
			wantPath: "/tmp/repos/test-repo",
			wantBare: false,
		},
		{
			name:     "bare_repository_without_working_directory",
			repoName: "bare-repo",
			repoPath: "/tmp/repos/bare-repo.git",
			isBare:   true,
			wantName: "bare-repo",
			wantPath: "/tmp/repos/bare-repo.git",
			wantBare: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(mockGitRepository)

			// Setup expectations
			mockRepo.On("Name").Return(testCase.repoName)
			mockRepo.On("Path").Return(testCase.repoPath)
			mockRepo.On("IsBare").Return(testCase.isBare)

			// Test properties
			require.Equal(t, testCase.wantName, mockRepo.Name())
			require.Equal(t, testCase.wantPath, mockRepo.Path())
			require.Equal(t, testCase.wantBare, mockRepo.IsBare())

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGitRepositoryBranches_HandlesDefaultBranchOperations tests branch operations without real git.
func TestGitRepository_DefaultBranchOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		currentBranch    string
		getBranchErr     error
		newBranch        string
		setBranchErr     error
		expectedGetCalls int
		expectedSetCalls int
	}{
		{
			name:             "gets_current_default_branch",
			currentBranch:    "main",
			getBranchErr:     nil,
			newBranch:        "",
			setBranchErr:     nil,
			expectedGetCalls: 1,
			expectedSetCalls: 0,
		},
		{
			name:             "sets_new_default_branch",
			currentBranch:    "main",
			getBranchErr:     nil,
			newBranch:        "develop",
			setBranchErr:     nil,
			expectedGetCalls: 1,
			expectedSetCalls: 1,
		},
		{
			name:             "handles_error_when_getting_branch",
			currentBranch:    "",
			getBranchErr:     errors.New("no default branch found"),
			newBranch:        "",
			setBranchErr:     nil,
			expectedGetCalls: 1,
			expectedSetCalls: 0,
		},
		{
			name:             "handles_error_when_setting_branch",
			currentBranch:    "main",
			getBranchErr:     nil,
			newBranch:        "invalid-branch",
			setBranchErr:     errors.New("branch 'invalid-branch' does not exist"),
			expectedGetCalls: 1,
			expectedSetCalls: 1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockRepo := new(mockGitRepository)

			// Setup get branch expectation
			if testCase.expectedGetCalls > 0 {
				mockRepo.On("GetDefaultBranch", ctx).Return(testCase.currentBranch, testCase.getBranchErr).Times(testCase.expectedGetCalls)
			}

			// Setup set branch expectation
			if testCase.expectedSetCalls > 0 {
				mockRepo.On("SetDefaultBranch", ctx, testCase.newBranch).Return(testCase.setBranchErr).Times(testCase.expectedSetCalls)
			}

			// Test getting default branch
			branch, err := mockRepo.GetDefaultBranch(ctx)
			if testCase.getBranchErr != nil {
				require.Error(t, err)
				require.Equal(t, testCase.getBranchErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.currentBranch, branch)
			}

			// Test setting default branch if needed
			if testCase.newBranch != "" {
				err = mockRepo.SetDefaultBranch(ctx, testCase.newBranch)
				if testCase.setBranchErr != nil {
					require.Error(t, err)
					require.Equal(t, testCase.setBranchErr.Error(), err.Error())
				} else {
					require.NoError(t, err)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
