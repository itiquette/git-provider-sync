// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package gitbinary provides git binary repository adapter.
package gitbinary

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	protocolHTTPS = "https"
	protocolSSH   = "ssh"
)

// Adapter provides git binary implementation of GitOperations port.
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

	// Create mirror service with timeout from config
	timeout := a.config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	mirrorSvc, err := NewMirrorServiceWithTimeout(ctx, logger, a.tempDir, timeout)
	if err != nil {
		return fmt.Errorf("failed to create git binary mirror service: %w", err)
	}

	a.mirrorSvc = mirrorSvc
	a.initialized = true

	return nil
}

// Clone implements the ports.GitOperations interface using git binary.
func (a *Adapter) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) { //nolint:ireturn
	if !a.initialized {
		return nil, domain.ErrAdapterNotInitialized
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
func (a *Adapter) Open(_ context.Context, path string) (ports.GitRepository, error) { //nolint:ireturn
	if !a.initialized {
		return nil, domain.ErrAdapterNotInitialized
	}

	// Check if path exists and is a git repository
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrNotGitRepository, path)
	}

	return NewGitRepository(path, a), nil
}

// Init implements the ports.GitOperations interface.
func (a *Adapter) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) { //nolint:ireturn
	if !a.initialized {
		return nil, domain.ErrAdapterNotInitialized
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
func (a *Adapter) Cleanup(_ context.Context, path string) error {
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
func (a *Adapter) SupportsURL(_ string) bool {
	// Git binary supports all git URLs
	return true
}

// GetName implements the ports.GitOperations interface.
func (a *Adapter) GetName() string {
	return "git-binary"
}

// CreateTmpDir implements the ports.GitOperations interface.
func (a *Adapter) CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	ctxWithTmp, err := entities.CreateTmpDir(ctx, dir, prefix)
	if err != nil {
		return ctx, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return ctxWithTmp, nil
}

// GetTmpDirPath implements the ports.GitOperations interface.
func (a *Adapter) GetTmpDirPath(ctx context.Context) (string, error) {
	path, err := entities.GetTmpDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get temporary directory path: %w", err)
	}

	return path, nil
}

// DeleteTmpDir implements the ports.GitOperations interface.
func (a *Adapter) DeleteTmpDir(ctx context.Context) error {
	if err := entities.DeleteTmpDir(ctx); err != nil {
		return fmt.Errorf("failed to delete temporary directory: %w", err)
	}

	return nil
}

// Helper methods

// convertAuthOptions converts ports.AuthOptions to gitbinary.AuthConfig.
func (a *Adapter) convertAuthOptions(auth ports.AuthOptions) AuthConfig {
	authConfig := AuthConfig{
		Token: auth.Token,
	}

	switch auth.Type {
	case ports.AuthTypeNone:
		// No authentication required
		authConfig.Protocol = protocolHTTPS
	case ports.AuthTypeBasic:
		authConfig.Protocol = protocolHTTPS
	case ports.AuthTypeToken:
		authConfig.Protocol = protocolHTTPS
	case ports.AuthTypeSSH:
		// Generic SSH type
		authConfig.Protocol = protocolSSH
	case ports.AuthTypeSSHKey, ports.AuthTypeSSHAgent:
		authConfig.Protocol = protocolSSH
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

var _ ports.GitOperations = (*Adapter)(nil)
