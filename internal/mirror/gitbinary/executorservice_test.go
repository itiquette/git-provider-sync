// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2
package gitbinary

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExecutorService_RunGitCommand(t *testing.T) {
	tests := []struct {
		name       string
		binary     string
		env        []string
		workingDir string
		args       []string
		setup      func(t *testing.T) string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "successful command execution",
			binary:     "true",
			workingDir: os.TempDir(),
			wantErr:    false,
		},
		{
			name:       "failed command execution",
			binary:     "false",
			workingDir: os.TempDir(),
			wantErr:    true,
			errMsg:     "exit status 1",
		},
		{
			name:       "execution with custom environment",
			workingDir: os.TempDir(),
			binary:     "true",
			env:        []string{"CUSTOM_VAR=test", "LANG=C"},
		},
		{
			name:       "execution with empty environment",
			workingDir: os.TempDir(),
			binary:     "true",
			env:        []string{},
		},
		{
			name:       "execution with nil environment",
			workingDir: os.TempDir(),
			binary:     "true",
			env:        nil,
		},
		{
			name:       "execution with invalid environment variable",
			workingDir: os.TempDir(),
			binary:     "true",
			env:        []string{"INVALID=VAR=VALUE"},
		},
		{
			name:       "execution with multiple environment variables",
			workingDir: os.TempDir(),
			binary:     "true",
			env:        []string{"VAR1=value1", "VAR2=value2", "VAR3=value3"},
		},
		{
			name:       "empty binary path",
			workingDir: os.TempDir(),
			binary:     "",
			wantErr:    true,
		},
		{
			name:    "empty workdir",
			binary:  "true",
			wantErr: true,
		},
		{
			name:   "very long working directory path",
			binary: "true",
			setup: func(t *testing.T) string {
				t.Helper()
				base := t.TempDir()
				longPath := base
				for i := 0; i < 10; i++ {
					longPath = filepath.Join(longPath, fmt.Sprintf("subdir%d", i))
				}
				require.NoError(t, os.MkdirAll(longPath, 0755))

				return longPath
			},
		},
		{
			name:   "unicode in working directory",
			binary: "true",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := filepath.Join(t.TempDir(), "τεστ")
				require.NoError(t, os.MkdirAll(dir, 0755))

				return dir
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			var workingDir string
			if tabletest.setup != nil {
				workingDir = tabletest.setup(t)
			}

			if tabletest.workingDir == "" {
				tabletest.workingDir = workingDir
			}

			executor := NewExecutorService(tabletest.binary)
			err := executor.RunGitCommand(context.Background(), tabletest.env, tabletest.workingDir, tabletest.args...)

			if tabletest.wantErr {
				require.Error(t, err)

				if tabletest.errMsg != "" {
					require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tabletest.errMsg))
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExecutorService_RunGitCommandWithOutput(t *testing.T) {
	tests := []struct {
		name       string
		binary     string
		workingDir string
		args       []string
		setup      func(t *testing.T) string
		wantErr    bool
		wantEmpty  bool
		timeout    time.Duration
	}{
		{
			name:      "successful command execution",
			binary:    "true",
			wantEmpty: true,
		},
		{
			name:      "failed command execution",
			binary:    "false",
			wantErr:   true,
			wantEmpty: true,
		},
		{
			name:    "multiple sequential timeouts",
			binary:  "true",
			timeout: 1 * time.Nanosecond,
			wantErr: true,
		},
		{
			name:   "binary with no execute permission",
			binary: filepath.Join(t.TempDir(), "non-executable"),
			setup: func(t *testing.T) string {
				t.Helper()
				path := filepath.Join(t.TempDir(), "non-executable")
				require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\ntrue"), 0600))

				return ""
			},
			wantErr: true,
		},
		{
			name:    "very long command path",
			binary:  filepath.Join(strings.Repeat("a", 1000), "true"),
			wantErr: true,
		},
	}

	for _, tabltest := range tests {
		t.Run(tabltest.name, func(t *testing.T) {
			if tabltest.setup != nil {
				tabltest.workingDir = tabltest.setup(t)
			}

			ctx := context.Background()

			if tabltest.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tabltest.timeout)
				defer cancel()
			}

			executor := NewExecutorService(tabltest.binary)
			output, err := executor.RunGitCommandWithOutput(ctx, tabltest.workingDir, tabltest.args...)

			if tabltest.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tabltest.wantEmpty {
				require.Empty(t, output)
			}
		})
	}
}

func TestExecutorService_ResourceCleanup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping resource cleanup test on Windows")
	}

	tests := []struct {
		name     string
		setup    func(t *testing.T) (*executorService, context.Context)
		validate func(t *testing.T, ctx context.Context)
	}{
		{
			name: "cleanup after normal execution",
			setup: func(_ *testing.T) (*executorService, context.Context) {
				return NewExecutorService("true"), context.Background()
			},
		},
		{
			name: "cleanup after timeout",
			setup: func(_ *testing.T) (*executorService, context.Context) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancel()

				return NewExecutorService("true"), ctx
			},
		},
		{
			name: "cleanup after cancellation",
			setup: func(_ *testing.T) (*executorService, context.Context) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return NewExecutorService("true"), ctx
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			executor, ctx := tabletest.setup(t)
			_, _ = executor.RunGitCommandWithOutput(ctx, "", "")

			if tabletest.validate != nil {
				tabletest.validate(t, ctx)
			}
		})
	}
}
