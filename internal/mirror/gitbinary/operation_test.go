// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

//nolint:wrapcheck
package gitbinary

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test constants for common values.
const (
	testPath = "/test/repo/path"
)

// Common errors for git operations.
var (
	errNetworkTimeout = errors.New("network timeout")
	errLockFailed     = errors.New("unable to create lock file")
	errRemoteUnreach  = errors.New("remote unreachable")
)

type mockExecutor struct {
	mock.Mock
	calls []string
}

func (m *mockExecutor) RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error {
	m.calls = append(m.calls, strings.Join(args, " "))
	callArgs := append([]interface{}{ctx, env, workingDir}, toInterfaces(args)...)

	return m.Called(callArgs...).Error(0)
}

func (m *mockExecutor) RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error) {
	callArgs := append([]interface{}{ctx, workingDir}, toInterfaces(args)...)
	call := m.Called(callArgs...)

	return call.Get(0).([]byte), call.Error(1) //nolint
}

func (m *mockExecutor) getCalls() []string {
	return m.calls
}

func newMockExecutor(t *testing.T) *mockExecutor {
	t.Helper()

	return &mockExecutor{
		calls: make([]string, 0),
	}
}

func TestOperation_Fetch(t *testing.T) {
	tests := []struct {
		name        string
		targetPath  string
		setupMock   func(*mockExecutor)
		expectCalls []string
		expectError error
		verifyFunc  func(*testing.T, *mockExecutor)
	}{
		{
			name:       "successful fetch and pull",
			targetPath: testPath,
			setupMock: func(mockE *mockExecutor) {
				t.Helper()
				// Fetch command
				mockE.On("RunGitCommand", mock.Anything, []string(nil), testPath, "fetch", "--all", "--prune").
					Return(nil)
				// Pull command
				mockE.On("RunGitCommand", mock.Anything, []string(nil), testPath, "pull", "--all").
					Return(nil)
				// Get branches
				mockE.On("RunGitCommandWithOutput", mock.Anything, testPath, "branch", "-r").
					Return([]byte("origin/main\norigin/develop"), nil)
				// Create tracking branches
				mockE.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", "main", "origin/main").
					Return(nil)
				mockE.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", "develop", "origin/develop").
					Return(nil)
			},
			expectCalls: []string{
				"fetch --all --prune",
				"pull --all",
				"branch --track main origin/main",
				"branch --track develop origin/develop",
			},
		},
		{
			name:       "network timeout during fetch",
			targetPath: testPath,
			setupMock: func(m *mockExecutor) {
				m.On("RunGitCommand", mock.Anything, []string(nil), testPath, "fetch", "--all", "--prune").
					Return(errNetworkTimeout)
			},
			expectError: errNetworkTimeout,
			expectCalls: []string{"fetch --all --prune"},
		},
		{
			name:       "lock file error during pull",
			targetPath: testPath,
			setupMock: func(m *mockExecutor) {
				m.On("RunGitCommand", mock.Anything, []string(nil), testPath, "fetch", "--all", "--prune").
					Return(nil)
				m.On("RunGitCommand", mock.Anything, []string(nil), testPath, "pull", "--all").
					Return(errLockFailed)
			},
			expectError: errLockFailed,
			expectCalls: []string{
				"fetch --all --prune",
				"pull --all",
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockE := newMockExecutor(t)
			if tabletest.setupMock != nil {
				tabletest.setupMock(mockE)
			}

			op := NewOperation(mockE)
			err := op.Fetch(context.Background(), tabletest.targetPath)

			if tabletest.expectError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectError)
			} else {
				require.NoError(t, err)
			}

			if tabletest.expectCalls != nil {
				require.Equal(t, tabletest.expectCalls, mockE.getCalls())
			}

			if tabletest.verifyFunc != nil {
				tabletest.verifyFunc(t, mockE)
			}

			mockE.AssertExpectations(t)
		})
	}
}

func TestOperation_CreateTrackingBranches(t *testing.T) {
	tests := []struct {
		name        string
		targetPath  string
		setupMock   func(*mockExecutor)
		expectCalls []string
		expectError error
	}{
		{
			name:       "create tracking branches for multiple remotes",
			targetPath: testPath,
			setupMock: func(m *mockExecutor) {
				m.On("RunGitCommandWithOutput", mock.Anything, testPath, "branch", "-r").
					Return([]byte("origin/main\norigin/develop\nupstream/main"), nil)
				m.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", mock.Anything, mock.Anything).
					Return(nil)
				m.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectCalls: []string{
				"branch --track main origin/main",
				"branch --track develop origin/develop",
				"branch --track upstream/main upstream/main",
			},
		},
		{
			name:       "handle detached HEAD reference",
			targetPath: testPath,
			setupMock: func(m *mockExecutor) {
				m.On("RunGitCommandWithOutput", mock.Anything, testPath, "branch", "-r").
					Return([]byte("origin/HEAD -> origin/main\norigin/develop"), nil)
				m.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", "develop", "origin/develop").
					Return(nil)
			},
			expectCalls: []string{
				"branch --track develop origin/develop",
			},
		},
		{
			name:       "handle no remote branches",
			targetPath: testPath,
			setupMock: func(m *mockExecutor) {
				m.On("RunGitCommandWithOutput", mock.Anything, testPath, "branch", "-r").
					Return([]byte(""), nil)
				m.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectCalls: []string{"branch --track  "},
		},
		{
			name:       "handle remote unreachable",
			targetPath: testPath,
			setupMock: func(m *mockExecutor) {
				m.On("RunGitCommandWithOutput", mock.Anything, testPath, "branch", "-r").
					Return([]byte{}, errRemoteUnreach)
			},
			expectError: ErrGetRemoteBranches,
			expectCalls: []string{},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockE := newMockExecutor(t)
			tabletest.setupMock(mockE)

			op := NewOperation(mockE)
			err := op.CreateTrackingBranches(context.Background(), tabletest.targetPath)

			if tabletest.expectError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tabletest.expectCalls, mockE.getCalls())
			mockE.AssertExpectations(t)
		})
	}
}

func TestOperation_ProcessTrackingBranches(t *testing.T) {
	tests := []struct {
		name        string
		targetPath  string
		input       []byte
		setupMock   func(*mockExecutor)
		expectCalls []string
	}{
		{
			name:       "process feature branches with special characters",
			targetPath: testPath,
			input:      []byte("origin/feature/ABC-123\norigin/feature/DEF-456\norigin/bugfix/GHI-789"),
			setupMock: func(mockE *mockExecutor) {
				branches := []struct{ local, remote string }{
					{"feature/ABC-123", "origin/feature/ABC-123"},
					{"feature/DEF-456", "origin/feature/DEF-456"},
					{"bugfix/GHI-789", "origin/bugfix/GHI-789"},
				}
				for _, b := range branches {
					mockE.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", b.local, b.remote).
						Return(nil)
				}
			},
			expectCalls: []string{
				"branch --track feature/ABC-123 origin/feature/ABC-123",
				"branch --track feature/DEF-456 origin/feature/DEF-456",
				"branch --track bugfix/GHI-789 origin/bugfix/GHI-789",
			},
		},
		{
			name:       "handle existing branches gracefully",
			targetPath: testPath,
			input:      []byte("origin/main\norigin/develop"),
			setupMock: func(mockE *mockExecutor) {
				mockE.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", "main", "origin/main").
					Return(errors.New("fatal: branch 'main' already exists"))
				mockE.On("RunGitCommand", mock.Anything, []string(nil), testPath, "branch", "--track", "develop", "origin/develop").
					Return(nil)
			},
			expectCalls: []string{
				"branch --track main origin/main",
				"branch --track develop origin/develop",
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockE := newMockExecutor(t)
			if tabletest.setupMock != nil {
				tabletest.setupMock(mockE)
			}

			op := NewOperation(mockE)
			err := op.ProcessTrackingBranches(context.Background(), tabletest.targetPath, tabletest.input)
			require.NoError(t, err)

			if tabletest.expectCalls != nil {
				require.Equal(t, tabletest.expectCalls, mockE.getCalls())
			}

			mockE.AssertExpectations(t)
		})
	}
}

// Helper function to convert string slice to interface slice.
func toInterfaces(ss []string) []interface{} {
	is := make([]interface{}, len(ss))
	for i, s := range ss {
		is[i] = s
	}

	return is
}
