// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Adapter provides git binary implementation of GitOperations port.
// This restores the git binary functionality from main branch in hexagonal architecture.
type Adapter struct {
	config      ports.GitConfig
	mirrorSvc   *MirrorService
	tempDir     string
	initialized bool
}

// New creates a new git binary adapter.
func New(config ports.GitConfig) *Adapter {
	return &Adapter{
		config:      config,
		initialized: false,
	}
}

// Initialize initializes the git binary adapter with proper dependencies.
func (a *Adapter) Initialize(ctx context.Context, logger ports.Logger) error {
	if a.initialized {
		return nil
	}

	// Create temporary directory
	a.tempDir = "/tmp/git-provider-sync-binary"

	// Create mirror service
	mirrorSvc, err := NewMirrorService(logger, a.tempDir)
	if err != nil {
		return fmt.Errorf("failed to create git binary mirror service: %w", err)
	}

	a.mirrorSvc = mirrorSvc
	a.initialized = true

	return nil
}

// Clone implements the ports.GitOperations interface using git binary.
func (a *Adapter) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) {
	if !a.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	// Ensure path exists
	if err := os.MkdirAll(filepath.Dir(options.Path), 0750); err != nil {
		return nil, fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Convert ports.CloneOptions to gitbinary.MirrorConfig
	config := MirrorConfig{
		SourceURL:    options.URL,
		Name:         filepath.Base(options.Path),
		AuthConfig:   a.convertAuthOptions(options.Auth),
		MirrorType:   a.determineMirrorType(options),
		ShallowDepth: options.Depth,
		DryRun:       false, // Not in CloneOptions
		SourceType:   a.detectSourceType(options.URL),
	}

	// Perform clone using mirror service
	result, err := a.mirrorSvc.Clone(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("git binary clone failed: %w", err)
	}

	// Create GitRepository wrapper
	return NewGitRepository(result.LocalPath, a), nil
}

// Open implements the ports.GitOperations interface.
func (a *Adapter) Open(ctx context.Context, path string) (ports.GitRepository, error) {
	if !a.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	// Check if path exists and is a git repository
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", path)
	}

	return NewGitRepository(path, a), nil
}

// Init implements the ports.GitOperations interface.
func (a *Adapter) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) {
	if !a.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(path, 0750); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Build git init command
	args := []string{"init"}
	if options.Bare {
		args = append(args, "--bare")
	}
	if options.DefaultBranch != "" {
		args = append(args, "--initial-branch", options.DefaultBranch)
	}
	if options.Template != "" {
		args = append(args, "--template", options.Template)
	}
	args = append(args, path)

	// Execute git init
	if err := a.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, ".", args...); err != nil {
		return nil, fmt.Errorf("failed to init repository: %w", err)
	}

	return NewGitRepository(path, a), nil
}

// Cleanup implements the ports.GitOperations interface.
func (a *Adapter) Cleanup(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}

	// Remove the repository directory
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to cleanup repository at %s: %w", path, err)
	}

	return nil
}

// SupportsURL implements the ports.GitOperations interface.
func (a *Adapter) SupportsURL(url string) bool {
	// Git binary supports all git URLs
	return true
}

// GetName implements the ports.GitOperations interface.
func (a *Adapter) GetName() string {
	return "git-binary"
}

// Helper methods

// convertAuthOptions converts ports.AuthOptions to gitbinary.AuthConfig.
func (a *Adapter) convertAuthOptions(auth ports.AuthOptions) AuthConfig {
	authConfig := AuthConfig{
		Token: auth.Token,
	}

	switch auth.Type {
	case ports.AuthTypeBasic:
		authConfig.Protocol = "https"
	case ports.AuthTypeToken:
		authConfig.Protocol = "https"
	case ports.AuthTypeSSHKey, ports.AuthTypeSSHAgent:
		authConfig.Protocol = "ssh"
		if len(auth.SSHKey) > 0 {
			authConfig.SSHCommand = "ssh -i /tmp/ssh_key"
		}
	}

	return authConfig
}

// determineMirrorType determines the mirror type from clone options.
func (a *Adapter) determineMirrorType(options ports.CloneOptions) string {
	if options.Mirror {
		return "mirror"
	}
	if options.Bare {
		return "bare"
	}
	if options.Depth > 0 {
		return "shallow"
	}

	return "full"
}

// detectSourceType detects the source type from URL.
func (a *Adapter) detectSourceType(url string) string {
	if strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://") {
		return "ssh"
	}

	return "https"
}

// Ensure Adapter implements ports.GitOperations interface
var _ ports.GitOperations = (*Adapter)(nil)
