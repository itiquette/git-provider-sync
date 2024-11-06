// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package target_test

import (
	"context"
	"testing"

	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"
	"itiquette/git-provider-sync/internal/target"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuth implements transport.AuthMethod.
type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) Name() string {
	args := m.Called()

	return args.String(0)
}

func (m *MockAuth) String() string {
	args := m.Called()

	return args.String(0)
}

func TestOpen(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		path    string
		setupFn func(t *testing.T, mock *mocks.GitLibOperation)
		wantErr error
	}{
		{
			name: "successful open",
			path: "/valid/repo/path",
			setupFn: func(_ *testing.T, mock *mocks.GitLibOperation) {
				mock.On("Open", ctx, "/valid/repo/path").
					Return(&git.Repository{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "open error",
			path: "/invalid/path",
			setupFn: func(_ *testing.T, mock *mocks.GitLibOperation) {
				mock.On("Open", ctx, "/invalid/path").
					Return(nil, target.ErrOpenRepository)
			},
			wantErr: target.ErrOpenRepository,
		},
	}

	for _, tableTest := range tests {
		t.Run(tableTest.name, func(t *testing.T) {
			gitlibOp := mocks.NewGitLibOperation(t)
			tableTest.setupFn(t, gitlibOp)

			repo, err := gitlibOp.Open(ctx, tableTest.path)
			if tableTest.wantErr != nil {
				require.ErrorIs(t, err, tableTest.wantErr)
				require.Nil(t, repo)
			} else {
				require.NoError(t, err)
				require.NotNil(t, repo)
			}
		})
	}
}

func TestGetWorktree(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		setupFn func(t *testing.T, gitlibOp *mocks.GitLibOperation)
		wantErr error
	}{
		{
			name: "successful worktree get",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("GetWorktree", ctx, mock.Anything).
					Return(&git.Worktree{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "worktree error",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.EXPECT().GetWorktree(ctx, mock.Anything).
					Return(nil, target.ErrWorktree)
			},
			wantErr: target.ErrWorktree,
		},
	}

	for _, tableTest := range tests {
		t.Run(tableTest.name, func(t *testing.T) {
			gitlibOp := mocks.NewGitLibOperation(t)
			tableTest.setupFn(t, gitlibOp)

			if tableTest.wantErr != nil {
				wt, err := gitlibOp.GetWorktree(ctx, &git.Repository{})
				require.ErrorIs(err, tableTest.wantErr)
				require.Nil(wt)
			} else {
				wt, err := gitlibOp.GetWorktree(ctx, &git.Repository{})
				require.NoError(err)
				require.NotNil(wt)
			}
		})
	}
}

func TestWorktreeStatus(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		setupFn func(t *testing.T, gitlibOp *mocks.GitLibOperation)
		wantErr error
	}{
		{
			name: "clean worktree",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("WorktreeStatus", ctx, mock.AnythingOfType("*git.Worktree")).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "unclean worktree",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("WorktreeStatus", ctx, mock.AnythingOfType("*git.Worktree")).
					Return(target.ErrUncleanWorkspace)
			},
			wantErr: target.ErrUncleanWorkspace,
		},
		{
			name: "status error",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("WorktreeStatus", ctx, mock.AnythingOfType("*git.Worktree")).
					Return(target.ErrWorktree)
			},
			wantErr: target.ErrWorktree,
		},
	}

	for _, tableTest := range tests {
		t.Run(tableTest.name, func(t *testing.T) {
			gitlibOp := mocks.NewGitLibOperation(t)
			tableTest.setupFn(t, gitlibOp)

			err := gitlibOp.WorktreeStatus(ctx, &git.Worktree{})
			if tableTest.wantErr != nil {
				require.ErrorIs(err, tableTest.wantErr)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestFetchBranches(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	mockAuth := &MockAuth{}
	mockAuth.On("Name").Return("mock_auth")
	mockAuth.On("String").Return("mock_auth")

	tests := []struct {
		name     string
		repoName string
		setupFn  func(t *testing.T, mock *mocks.GitLibOperation)
		wantErr  error
	}{
		{
			name:     "successful fetch",
			repoName: "test-repo",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("FetchBranches", ctx, mock.AnythingOfType("*git.Repository"), mockAuth, "test-repo").
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:     "already up to date",
			repoName: "test-repo",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("FetchBranches", ctx, mock.AnythingOfType("*git.Repository"), mockAuth, "test-repo").
					Return(git.NoErrAlreadyUpToDate)
			},
			wantErr: git.NoErrAlreadyUpToDate,
		},
		{
			name:     "fetch error",
			repoName: "test-repo",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("FetchBranches", ctx, mock.AnythingOfType("*git.Repository"), mockAuth, "test-repo").
					Return(target.ErrFetchBranches)
			},
			wantErr: target.ErrFetchBranches,
		},
	}

	for _, tableTest := range tests {
		t.Run(tableTest.name, func(t *testing.T) {
			mock := mocks.NewGitLibOperation(t)
			tableTest.setupFn(t, mock)

			err := mock.FetchBranches(ctx, &git.Repository{}, mockAuth, tableTest.repoName)
			if tableTest.wantErr != nil {
				require.ErrorIs(err, tableTest.wantErr)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestSetDefaultBranch(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		branchName string
		setupFn    func(t *testing.T, mock *mocks.GitLibOperation)
		wantErr    error
	}{
		{
			name:       "successful branch set",
			branchName: "main",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("SetDefaultBranch", ctx, mock.AnythingOfType("*git.Repository"), "main").
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:       "worktree error",
			branchName: "main",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("SetDefaultBranch", ctx, mock.AnythingOfType("*git.Repository"), "main").
					Return(target.ErrWorktree)
			},
			wantErr: target.ErrWorktree,
		},
		{
			name:       "checkout error",
			branchName: "main",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("SetDefaultBranch", ctx, mock.AnythingOfType("*git.Repository"), "main").
					Return(target.ErrBranchCheckout)
			},
			wantErr: target.ErrBranchCheckout,
		},
		{
			name:       "head set error",
			branchName: "main",
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("SetDefaultBranch", ctx, mock.AnythingOfType("*git.Repository"), "main").
					Return(target.ErrHeadSet)
			},
			wantErr: target.ErrHeadSet,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			gitlibOp := mocks.NewGitLibOperation(t)
			tabletest.setupFn(t, gitlibOp)

			err := gitlibOp.SetDefaultBranch(ctx, &git.Repository{}, tabletest.branchName)
			if tabletest.wantErr != nil {
				require.ErrorIs(err, tabletest.wantErr)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestCreateRemote(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		remote  string
		urls    []string
		setupFn func(t *testing.T, gitLibOp *mocks.GitLibOperation)
		wantErr error
	}{
		{
			name:   "successful remote creation",
			remote: "origin",
			urls:   []string{"https://gathub.com/test/repo.git"},
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("CreateRemote", ctx, mock.AnythingOfType("*git.Repository"), "origin",
					[]string{"https://gathub.com/test/repo.git"}).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "remote creation error",
			remote: "origin",
			urls:   []string{"https://gathub.com/test/repo.git"},
			setupFn: func(_ *testing.T, gitlibOp *mocks.GitLibOperation) {
				gitlibOp.On("CreateRemote", ctx, mock.AnythingOfType("*git.Repository"), "origin",
					[]string{"https://gathub.com/test/repo.git"}).
					Return(target.ErrRemoteCreation)
			},
			wantErr: target.ErrRemoteCreation,
		},
	}

	for _, tableTest := range tests {
		t.Run(tableTest.name, func(t *testing.T) {
			gitlibOp := mocks.NewGitLibOperation(t)
			tableTest.setupFn(t, gitlibOp)

			err := gitlibOp.CreateRemote(ctx, &git.Repository{}, tableTest.remote, tableTest.urls)
			if tableTest.wantErr != nil {
				require.ErrorIs(err, tableTest.wantErr)
			} else {
				require.NoError(err)
			}
		})
	}
}
