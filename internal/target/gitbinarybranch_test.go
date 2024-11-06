// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package target

import (
	"context"
	"errors"
	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name        string
		targetPath  string
		setupMock   func(*mocks.CommandExecutor)
		expectError bool
	}{
		{
			name:       "successful fetch and pull",
			targetPath: "/test/path",
			setupMock: func(moc *mocks.CommandExecutor) {
				// Expect fetch command
				moc.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "fetch", "--all", "--prune").
					Return(nil)

				// Expect pull command
				moc.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "pull", "--all").
					Return(nil)

				// Mock the branch -r command
				moc.EXPECT().
					RunGitCommandWithOutput(context.Background(), "/test/path", "branch", "-r").
					Return([]byte("origin/main\norigin/develop"), nil)

				// Mock tracking branch creation
				moc.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "branch", "--track", "main", "origin/main").
					Return(nil)

				moc.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "branch", "--track", "develop", "origin/develop").
					Return(nil)
			},
			expectError: false,
		},
		{
			name:       "fetch command fails",
			targetPath: "/test/path",
			setupMock: func(m *mocks.CommandExecutor) {
				m.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "fetch", "--all", "--prune").
					Return(errors.New("fetch failed"))
			},
			expectError: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockExecutor := mocks.NewCommandExecutor(t)
			tabletest.setupMock(mockExecutor)

			branch := newGitBranch(mockExecutor)
			err := branch.Fetch(context.Background(), tabletest.targetPath)

			if tabletest.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateTrackingBranches(t *testing.T) {
	tests := []struct {
		name         string
		targetPath   string
		branchOutput string
		setupMock    func(*mocks.CommandExecutor)
		expectError  bool
	}{
		{
			name:         "successful tracking branch creation",
			targetPath:   "/test/path",
			branchOutput: "origin/main\norigin/develop",
			setupMock: func(moc *mocks.CommandExecutor) {
				moc.EXPECT().
					RunGitCommandWithOutput(context.Background(), "/test/path", "branch", "-r").
					Return([]byte("origin/main\norigin/develop"), nil)

				// Expect tracking branch creation for main
				moc.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "branch", "--track", "main", "origin/main").
					Return(nil)

				// Expect tracking branch creation for develop
				moc.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "branch", "--track", "develop", "origin/develop").
					Return(nil)
			},
			expectError: false,
		},
		{
			name:       "branch command fails",
			targetPath: "/test/path",
			setupMock: func(m *mocks.CommandExecutor) {
				m.EXPECT().
					RunGitCommandWithOutput(context.Background(), "/test/path", "branch", "-r").
					Return([]byte{}, errors.New("branch command failed"))
			},
			expectError: true,
		},
		{
			name:         "handle branch with arrow",
			targetPath:   "/test/path",
			branchOutput: "origin/HEAD -> origin/main\norigin/develop",
			setupMock: func(moc *mocks.CommandExecutor) {
				moc.EXPECT().
					RunGitCommandWithOutput(context.Background(), "/test/path", "branch", "-r").
					Return([]byte("origin/HEAD -> origin/main\norigin/develop"), nil)

				// Only expect tracking branch creation for develop
				moc.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "branch", "--track", "develop", "origin/develop").
					Return(nil)
			},
			expectError: false,
		},
	}

	for _, tableTest := range tests {
		t.Run(tableTest.name, func(t *testing.T) {
			mockExecutor := mocks.NewCommandExecutor(t)
			tableTest.setupMock(mockExecutor)

			branch := newGitBranch(mockExecutor)
			err := branch.CreateTrackingBranches(context.Background(), tableTest.targetPath)

			if tableTest.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProcessTrackingBranches(t *testing.T) {
	tests := []struct {
		name        string
		targetPath  string
		input       []byte
		setupMock   func(*mocks.CommandExecutor)
		expectError bool
	}{
		{
			name:       "process multiple branches",
			targetPath: "/test/path",
			input:      []byte("origin/main\norigin/develop\norigin/feature"),
			setupMock: func(m *mocks.CommandExecutor) {
				// Expect tracking branch creation for all branches
				branches := []string{"main", "develop", "feature"}
				for _, branch := range branches {
					m.EXPECT().
						RunGitCommand(context.Background(), []string(nil), "/test/path", "branch", "--track", branch, "origin/"+branch).
						Return(nil)
				}
			},
			expectError: false,
		},
		{
			name:       "handle already existing branch",
			targetPath: "/test/path",
			input:      []byte("origin/main"),
			setupMock: func(m *mocks.CommandExecutor) {
				m.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "branch", "--track", "main", "origin/main").
					Return(errors.New("fatal: branch 'main' already exists"))
			},
			expectError: false,
		},
		{
			name:       "handle branch with arrow",
			targetPath: "/test/path",
			input:      []byte("origin/HEAD -> origin/main\norigin/develop"),
			setupMock: func(m *mocks.CommandExecutor) {
				m.EXPECT().
					RunGitCommand(context.Background(), []string(nil), "/test/path", "branch", "--track", "develop", "origin/develop").
					Return(nil)
			},
			expectError: false,
		},
	}

	for _, tableTest := range tests {
		t.Run(tableTest.name, func(t *testing.T) {
			mockExecutor := mocks.NewCommandExecutor(t)
			tableTest.setupMock(mockExecutor)

			branch := newGitBranch(mockExecutor)
			err := branch.ProcessTrackingBranches(context.Background(), tableTest.targetPath, tableTest.input)

			if tableTest.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
