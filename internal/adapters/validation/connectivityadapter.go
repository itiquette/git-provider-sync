// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides adapters for validation operations.
package validation

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/validation"
)

// ConnectivityAdapter implements connectivity validation using HTTP and network operations.
type ConnectivityAdapter struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewConnectivityAdapter creates a new connectivity validation adapter.
func NewConnectivityAdapter(timeout time.Duration) *ConnectivityAdapter {
	return &ConnectivityAdapter{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// ValidateConnectivity executes a connectivity validation.
func (c *ConnectivityAdapter) ValidateConnectivity(ctx context.Context, val validation.ConnectivityValidation) validation.ConnectivityResult {
	start := time.Now()

	result := validation.ConnectivityResult{
		Validation: val,
		Success:    false,
		Details:    make(map[string]any),
		Duration:   0,
	}

	// Set timeout from validation or use default
	timeout := val.Timeout
	if timeout == 0 {
		timeout = c.timeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch val.Type {
	case validation.ConnectivityTypeHTTP:
		result.Success, result.Error = c.validateHTTP(ctx, val.Target)
	case validation.ConnectivityTypeProvider:
		result.Success, result.Error = c.validateProvider(ctx, val.Target)
	case validation.ConnectivityTypeGit:
		result.Success, result.Error = c.validateGit(ctx, val.Target)
	case validation.ConnectivityTypeSSH:
		result.Success, result.Error = c.validateSSH(ctx, val.Target)
	default:
		result.Error = fmt.Errorf("%w: %s", domain.ErrUnsupportedConnectivityType, val.Type)
	}

	result.Duration = time.Since(start)

	return result
}

// validateHTTP tests HTTP connectivity to a URL.
func (c *ConnectivityAdapter) validateHTTP(ctx context.Context, target string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, target, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request for connectivity check to %s: %w", target, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("HTTP connectivity check failed for %s: %w", target, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	// Accept any 2xx or 3xx response as success
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, nil
	}

	return false, fmt.Errorf("%w: %d", domain.ErrHTTPRequestFailed, resp.StatusCode)
}

// validateProvider tests provider API connectivity.
func (c *ConnectivityAdapter) validateProvider(ctx context.Context, target string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create provider API request for %s: %w", target, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("provider API connectivity test failed for %s: %w", target, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	// Accept any response that's not a connection error
	// Even 404 means we can reach the provider
	if resp.StatusCode < 500 {
		return true, nil
	}

	return false, fmt.Errorf("%w: %d", domain.ErrProviderAPIServerError, resp.StatusCode)
}

// validateGit tests Git connectivity (simplified - just HTTP check).
func (c *ConnectivityAdapter) validateGit(ctx context.Context, target string) (bool, error) {
	// For Git validation, we'll do a HEAD request to the Git URL
	// This is a simplified check - a full implementation might try git ls-remote
	return c.validateHTTP(ctx, target)
}

// validateSSH tests SSH connectivity to a host.
func (c *ConnectivityAdapter) validateSSH(ctx context.Context, target string) (bool, error) {
	// Extract host from git@host format and add port
	var host string
	if len(target) > 4 && target[:4] == "git@" {
		host = target[4:] + ":22"
	} else {
		host = target + ":22"
	}

	// Test TCP connection to SSH port
	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return false, fmt.Errorf("SSH connectivity test failed for %s (trying %s): %w", target, host, err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	return true, nil
}

// FileSystemAdapter implements file system validation.
type FileSystemAdapter struct{}

// NewFileSystemAdapter creates a new file system validation adapter.
func NewFileSystemAdapter() *FileSystemAdapter {
	return &FileSystemAdapter{}
}

// ValidateFileSystem executes a file system validation.
func (f *FileSystemAdapter) ValidateFileSystem(_ context.Context, val validation.FileSystemValidation) validation.FileSystemResult {
	result := validation.FileSystemResult{
		Validation: val,
		Success:    false,
		Details:    make(map[string]any),
	}

	switch val.Type {
	case validation.FileSystemTypeDirectory:
		result.Success, result.Exists, result.Readable, result.Writable, result.Error = f.validateDirectory(val.Path, val.Writable)
	case validation.FileSystemTypeFile:
		result.Success, result.Exists, result.Readable, result.Writable, result.Error = f.validateFile(val.Path, val.Writable)
	case validation.FileSystemTypeArchive:
		result.Success, result.Exists, result.Readable, result.Writable, result.Error = f.validateArchive(val.Path)
	default:
		result.Error = fmt.Errorf("%w: %s", domain.ErrUnsupportedFileSystemType, val.Type)
	}

	// Add details
	result.Details["path"] = val.Path
	result.Details["exists"] = result.Exists
	result.Details["readable"] = result.Readable
	result.Details["writable"] = result.Writable

	return result
}

// validateDirectory validates directory access.
func (f *FileSystemAdapter) validateDirectory(path string, needsWritable bool) (bool, bool, bool, bool, error) {
	info, err := statPath(path)
	if err != nil {
		return false, false, false, false, err
	}

	if !info.IsDir() {
		return false, true, false, false, domain.ErrPathNotDirectory
	}

	readable := isReadable(path)
	writable := isWritable(path)

	success := readable && (!needsWritable || writable)

	return success, true, readable, writable, nil
}

// validateFile validates file access.
func (f *FileSystemAdapter) validateFile(path string, needsWritable bool) (bool, bool, bool, bool, error) {
	info, err := statPath(path)
	if err != nil {
		return false, false, false, false, err
	}

	if info.IsDir() {
		return false, true, false, false, domain.ErrPathIsDirectory
	}

	readable := isReadable(path)
	writable := isWritable(path)

	success := readable && (!needsWritable || writable)

	return success, true, readable, writable, nil
}

// validateArchive validates archive file access.
func (f *FileSystemAdapter) validateArchive(path string) (bool, bool, bool, bool, error) {
	// For archive validation, check if file exists and is readable
	return f.validateFile(path, false)
}
