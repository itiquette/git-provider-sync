// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package target

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context //nolint
		env        []string
		workingDir string
		args       []string
		wantErr    bool
		setup      func() (string, error)
		verify     func(*testing.T, string)
	}{
		{
			name:       "successful command execution",
			ctx:        context.Background(),
			env:        nil,
			workingDir: "",
			args:       []string{"-la"},
			wantErr:    false,
		},
		{
			name:       "execution with working directory",
			ctx:        context.Background(),
			env:        nil,
			workingDir: t.TempDir(),
			args:       []string{"-la"},
			wantErr:    false,
		},
		{
			name:       "execution with environment variables",
			ctx:        context.Background(),
			env:        []string{"LS_COLORS=auto"},
			workingDir: "",
			args:       []string{"-la"},
			wantErr:    false,
		},
		{
			name: "context timeout",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				time.Sleep(time.Microsecond)
				defer cancel()

				return ctx
			}(),
			env:        nil,
			workingDir: "",
			args:       []string{"-al"}, // Use recursive listing to ensure it takes time
			wantErr:    true,
		},
		{
			name:       "invalid working directory",
			ctx:        context.Background(),
			env:        nil,
			workingDir: "/nonexistent/directory",
			args:       []string{"-la"},
			wantErr:    true,
		},
		{
			name:       "invalid argument",
			ctx:        context.Background(),
			env:        nil,
			workingDir: "",
			args:       []string{"--invalid-flag"},
			wantErr:    true,
		},
		{
			name: "list specific files",
			ctx:  context.Background(),
			env:  nil,
			setup: func() (string, error) {
				dir := t.TempDir()
				err := os.WriteFile(filepath.Join(dir, "testfile"), []byte("test"), 0600)

				return dir, err //nolint
			},
			args:    []string{"-la"},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				t.Helper()
				files, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Contains(t, files[0].Name(), "testfile")
			},
		},
		{
			name: "context canceled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			env:        nil,
			workingDir: "",
			args:       []string{"-la"},
			wantErr:    true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			var dir string

			if tabletest.setup != nil {
				var err error
				dir, err = tabletest.setup()
				require.NoError(t, err)

				if tabletest.workingDir == "" {
					tabletest.workingDir = dir
				}
			}

			executor := newExecService("ls")
			err := executor.RunGitCommand(tabletest.ctx, tabletest.env, tabletest.workingDir, tabletest.args...)

			if tabletest.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tabletest.verify != nil {
				tabletest.verify(t, dir)
			}
		})
	}
}

func TestRunCommandWithOutput(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context //nolint
		workingDir   string
		args         []string
		wantErr      bool
		wantContains string
		setup        func() (string, error)
		verifyOutput func(*testing.T, []byte)
	}{
		{
			name:         "successful command with output",
			ctx:          context.Background(),
			args:         []string{"-la"},
			wantErr:      false,
			wantContains: "total", // Common output in ls -la across Unix systems
		},
		{
			name: "list specific file",
			ctx:  context.Background(),
			setup: func() (string, error) {
				dir := t.TempDir()

				return dir, os.WriteFile(filepath.Join(dir, "testfile"), []byte("test"), 0600)
			},
			args:         []string{"-la"},
			wantErr:      false,
			wantContains: "testfile",
		},
		{
			name:       "invalid directory",
			ctx:        context.Background(),
			workingDir: "/nonexistent/directory",
			args:       []string{"-la"},
			wantErr:    true,
		},
		{
			name: "context canceled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			args:    []string{"-la"},
			wantErr: true,
		},
		{
			name: "verify output format",
			ctx:  context.Background(),
			args: []string{"-l"},
			verifyOutput: func(t *testing.T, output []byte) {
				// Check if output contains typical ls -l format elements
				t.Helper()
				outputStr := string(output)
				require.Contains(t, outputStr, "total")
				require.Regexp(t, `[drwx-]{10}`, outputStr) // Permission bits
			},
			wantErr: false,
		},
		{
			name: "multiple files listing",
			ctx:  context.Background(),
			setup: func() (string, error) {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "file1"), []byte("test1"), 0600); err != nil {
					return "", err //nolint
				}
				if err := os.WriteFile(filepath.Join(dir, "file2"), []byte("test2"), 0600); err != nil {
					return "", err //nolint
				}

				return dir, nil
			},
			args:         []string{"-1"}, // List one file per line
			wantErr:      false,
			wantContains: "file1",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			if tabletest.setup != nil {
				dir, err := tabletest.setup()
				require.NoError(t, err)

				if tabletest.workingDir == "" {
					tabletest.workingDir = dir
				}
			}

			executor := newExecService("ls")
			output, err := executor.RunGitCommandWithOutput(tabletest.ctx, tabletest.workingDir, tabletest.args...)

			if tabletest.wantErr {
				require.Error(t, err)
				require.Nil(t, output)
			} else {
				require.NoError(t, err)
				require.NotNil(t, output)

				if tabletest.wantContains != "" {
					require.Contains(t, string(output), tabletest.wantContains)
				}

				if tabletest.verifyOutput != nil {
					tabletest.verifyOutput(t, output)
				}
			}
		})
	}
}

func TestCommandTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		args    []string
		wantErr error
	}{
		{
			name:    "command completes within timeout",
			timeout: time.Second,
			args:    []string{"-la"},
			wantErr: nil,
		},
		{
			name:    "command times out",
			timeout: time.Nanosecond,
			args:    []string{"-la"},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tabletest.timeout)
			defer cancel()

			executor := newExecService("ls")
			_, err := executor.RunGitCommandWithOutput(ctx, "", tabletest.args...)

			if tabletest.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
