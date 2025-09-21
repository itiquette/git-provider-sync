// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	protocolHTTPS = "https"
	protocolSSH   = "ssh"

	// DefaultTimeout is the default timeout for git operations.
	DefaultTimeout = 5 * time.Minute

	// TempDirPrefix is the prefix for temporary directories created by the adapter.
	TempDirPrefix = "git-provider-sync-binary-"
)

// Adapter provides git binary implementation of GitOperations port.
// Now uses functional options for proper dependency injection - no two-phase init.
type Adapter struct {
	config     ports.GitConfig
	mirrorSvc  *MirrorService
	tempDir    string
	fileSystem ports.FileSystem
	logger     ports.Logger
}

// Option is a functional option for configuring the Adapter.
type Option func(*Adapter) error

// WithFileSystem sets a custom filesystem implementation.
func WithFileSystem(fs ports.FileSystem) Option {
	return func(a *Adapter) error {
		a.fileSystem = fs

		return nil
	}
}

// WithTempDir sets a specific temp directory for operations.
func WithTempDir(tempDir string) Option {
	return func(a *Adapter) error {
		a.tempDir = tempDir

		return nil
	}
}

// WithMirrorService sets a custom mirror service (useful for testing).
func WithMirrorService(svc *MirrorService) Option {
	return func(a *Adapter) error {
		a.mirrorSvc = svc

		return nil
	}
}

// NewWithDependencies creates a new git binary adapter with all dependencies injected.
// Uses functional options pattern for clean dependency injection.
// No more two-phase initialization - fully initialized on construction.
func NewWithDependencies(ctx context.Context, config ports.GitConfig, logger ports.Logger, opts ...Option) (*Adapter, error) {
	adapter := &Adapter{
		config:     config,
		logger:     logger,
		fileSystem: filesystem.NewOSFileSystem(), // Default filesystem
	}

	// Apply functional options
	for _, opt := range opts {
		if err := opt(adapter); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create temporary directory if not provided
	if adapter.tempDir == "" {
		tempDir, err := os.MkdirTemp("", TempDirPrefix+"*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		adapter.tempDir = tempDir
	}

	// Create mirror service if not provided (allows injection for testing)
	//nolint:nestif // Initialization logic requires nested checks
	if adapter.mirrorSvc == nil {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}

		mirrorSvc, err := NewMirrorServiceWithTimeout(ctx, logger, adapter.tempDir, timeout)
		if err != nil {
			originalErr := err
			// Clean up temp dir if we created it
			if adapter.tempDir != "" {
				// Best effort cleanup - we're already returning an error
				if cleanupErr := os.RemoveAll(adapter.tempDir); cleanupErr != nil {
					// Combine errors explicitly
					return nil, fmt.Errorf("failed to create git binary mirror service: %w (cleanup also failed: %w)", originalErr, cleanupErr)
				}
			}

			return nil, fmt.Errorf("failed to create git binary mirror service: %w", originalErr)
		}

		adapter.mirrorSvc = mirrorSvc
	}

	return adapter, nil
}

// Clone implements the ports.GitOperations interface using git binary.
func (a *Adapter) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) { //nolint:ireturn
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
	// Check if path exists and is a git repository
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrNotGitRepository, path)
	}

	return NewGitRepository(path, a), nil
}

// Init implements the ports.GitOperations interface.
func (a *Adapter) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) { //nolint:ireturn
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Initialize git repository
	args := []string{"init"}
	if options.Bare {
		args = append(args, "--bare")
	}

	args = append(args, path)

	if err := a.mirrorSvc.executorSvc.RunGitCommand(ctx, nil, "", args...); err != nil {
		return nil, fmt.Errorf("git init failed: %w", err)
	}

	// Set initial branch if specified
	if options.DefaultBranch != "" {
		repo := NewGitRepository(path, a)

		err := a.mirrorSvc.executorSvc.RunGitCommand(ctx, nil, path,
			"symbolic-ref", "HEAD", "refs/heads/"+options.DefaultBranch)
		if err != nil {
			return nil, fmt.Errorf("failed to set default branch: %w", err)
		}

		return repo, nil
	}

	return NewGitRepository(path, a), nil
}

// Cleanup cleans up any temporary resources.
func (a *Adapter) Cleanup(_ context.Context, path string) error {
	// If a specific path is provided, remove it
	if path != "" {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove path %s: %w", path, err)
		}

		return nil
	}

	// Otherwise, clean up the temp dir if it was created by us
	if a.tempDir != "" && strings.Contains(a.tempDir, TempDirPrefix) {
		if err := os.RemoveAll(a.tempDir); err != nil {
			return fmt.Errorf("failed to remove temp dir %s: %w", a.tempDir, err)
		}

		return nil
	}

	return nil
}

// convertAuthOptions converts ports auth options to gitbinary auth config.
func (a *Adapter) convertAuthOptions(auth ports.AuthOptions) AuthConfig {
	authConfig := AuthConfig{
		Token: auth.Token,
	}

	// Determine protocol and configure SSH if needed
	if a.isSSHAuth(auth.Type) {
		authConfig.Protocol = protocolSSH
		authConfig.SSHCommand = a.buildSSHCommand(auth)

		// Configure URL rewriting for SSH with token (GitHub-specific)
		if auth.Type == ports.AuthTypeSSH && auth.Token != "" {
			authConfig.SSHURLRewriteFrom = protocolHTTPS + "://github.com/"
			authConfig.SSHURLRewriteTo = "git@github.com:"
		}
	} else {
		authConfig.Protocol = protocolHTTPS
	}

	return authConfig
}

// isSSHAuth checks if the auth type is SSH-based.
func (a *Adapter) isSSHAuth(authType ports.AuthType) bool {
	return authType == ports.AuthTypeSSH ||
		authType == ports.AuthTypeSSHAgent ||
		authType == ports.AuthTypeSSHKey
}

// buildSSHCommand builds the SSH command based on auth options.
func (a *Adapter) buildSSHCommand(auth ports.AuthOptions) string {
	switch auth.Type {
	case ports.AuthTypeSSHKey:
		return a.buildSSHKeyCommand(auth)
	case ports.AuthTypeSSH:
		if auth.SSHKeyPath != "" {
			return fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", auth.SSHKeyPath)
		}

		return ""
	case ports.AuthTypeNone, ports.AuthTypeBasic, ports.AuthTypeToken, ports.AuthTypeSSHAgent:
		// These auth types don't use SSH commands
		return ""
	default:
		// Unknown auth type
		return ""
	}
}

// buildSSHKeyCommand builds SSH command for SSH key authentication.
func (a *Adapter) buildSSHKeyCommand(auth ports.AuthOptions) string {
	// If we have key bytes, write to temp file
	if len(auth.SSHKey) > 0 {
		keyFile := filepath.Join(a.tempDir, "ssh_key")
		if err := os.WriteFile(keyFile, auth.SSHKey, 0600); err == nil {
			return fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", keyFile)
		}
		// Fall through to use path if write fails
	}

	// Use key path if available
	if auth.SSHKeyPath != "" {
		return fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", auth.SSHKeyPath)
	}

	return ""
}

// determineMirrorType determines the mirror type based on clone options.
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

// detectSourceType attempts to detect the source type from URL.
func (a *Adapter) detectSourceType(url string) string {
	url = strings.ToLower(url)

	switch {
	case strings.Contains(url, "github.com"):
		return "github"
	case strings.Contains(url, "gitlab.com"):
		return "gitlab"
	case strings.Contains(url, "gitea.com"), strings.Contains(url, "gitea.io"):
		return "gitea"
	case strings.Contains(url, "bitbucket.org"):
		return "bitbucket"
	default:
		return "generic"
	}
}

// GetMirrorService returns the mirror service for direct access if needed.
func (a *Adapter) GetMirrorService() *MirrorService {
	return a.mirrorSvc
}

// GetConfig returns the git configuration.
func (a *Adapter) GetConfig() ports.GitConfig {
	return a.config
}

// contextKey is a type for context keys to avoid collisions.
type contextKey string

const (
	// tmpDirKey is the context key for temporary directory path.
	// TODO: This is an anti-pattern - context should not be used for data storage.
	// The interface should be refactored to return the tmpDir as a value.
	tmpDirKey contextKey = "tmpDir"
)

// CreateTmpDir creates a temporary directory for git operations.
// TODO: Refactor interface to not use context for data storage.
func (a *Adapter) CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	tmpDir, err := os.MkdirTemp(dir, prefix)
	if err != nil {
		return ctx, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Store the temp directory path in context for cleanup later
	// This is an anti-pattern but required by the current interface
	ctx = context.WithValue(ctx, tmpDirKey, tmpDir)

	return ctx, nil
}

// GetTmpDirPath retrieves the temp directory path from context.
// TODO: Refactor to not use context for data storage.
func (a *Adapter) GetTmpDirPath(ctx context.Context) (string, error) {
	tmpDir, ok := ctx.Value(tmpDirKey).(string)
	if !ok {
		return "", errors.New("no temp directory found in context")
	}

	return tmpDir, nil
}

// DeleteTmpDir deletes the temporary directory stored in context.
// TODO: Refactor to not use context for data storage.
func (a *Adapter) DeleteTmpDir(ctx context.Context) error {
	tmpDir, err := a.GetTmpDirPath(ctx)
	if err != nil {
		//nolint:nilerr // No temp directory to delete is not an error
		return nil // No temp directory to delete
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to remove temp dir %s: %w", tmpDir, err)
	}

	return nil
}

// SupportsURL checks if the adapter supports the given URL.
func (a *Adapter) SupportsURL(url string) bool {
	// Git binary supports all git URLs
	return strings.HasPrefix(url, "git://") ||
		strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "ssh://") ||
		strings.Contains(url, "@") // Support git@github.com style URLs
}

// GetName returns the name of the git implementation.
func (a *Adapter) GetName() string {
	return "gitbinary"
}
