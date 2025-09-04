// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/validation"
)

func TestConnectivityAdapter_ValidateConnectivity_HTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		expectedResult bool
		expectedError  string
	}{
		{
			name: "successful HTTP request",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			expectedResult: true,
		},
		{
			name: "HTTP redirect still succeeds",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusFound)
				}))
			},
			expectedResult: true,
		},
		{
			name: "HTTP client error fails",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedResult: false,
			expectedError:  "HTTP request returned non-2xx/3xx status: 404",
		},
		{
			name: "HTTP server error fails",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectedResult: false,
			expectedError:  "HTTP request returned non-2xx/3xx status: 500",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := NewConnectivityAdapter(5 * time.Second)
			ctx := context.Background()

			val := validation.ConnectivityValidation{
				Type:   validation.ConnectivityTypeHTTP,
				Target: server.URL,
			}

			result := adapter.ValidateConnectivity(ctx, val)

			assert.Equal(t, test.expectedResult, result.Success)
			assert.Greater(t, result.Duration, time.Duration(0))
			assert.Equal(t, val, result.Validation)

			if test.expectedError != "" {
				require.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), test.expectedError)
			} else {
				require.NoError(t, result.Error)
			}
		})
	}
}

func TestConnectivityAdapter_ValidateConnectivity_Provider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		expectedResult bool
		expectedError  string
	}{
		{
			name: "provider API success",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			expectedResult: true,
		},
		{
			name: "provider API not found still succeeds",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedResult: true, // 404 means we can reach provider
		},
		{
			name: "provider API server error fails",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectedResult: false,
			expectedError:  "provider API returned server error: 500",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := NewConnectivityAdapter(5 * time.Second)
			ctx := context.Background()

			val := validation.ConnectivityValidation{
				Type:   validation.ConnectivityTypeProvider,
				Target: server.URL,
			}

			result := adapter.ValidateConnectivity(ctx, val)

			assert.Equal(t, test.expectedResult, result.Success)
			assert.Greater(t, result.Duration, time.Duration(0))

			if test.expectedError != "" {
				require.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), test.expectedError)
			} else {
				require.NoError(t, result.Error)
			}
		})
	}
}

func TestConnectivityAdapter_ValidateConnectivity_Git(t *testing.T) {
	t.Parallel()

	// Git validation is implemented as HTTP validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter := NewConnectivityAdapter(5 * time.Second)
	ctx := context.Background()

	val := validation.ConnectivityValidation{
		Type:   validation.ConnectivityTypeGit,
		Target: server.URL,
	}

	result := adapter.ValidateConnectivity(ctx, val)

	assert.True(t, result.Success)
	require.NoError(t, result.Error)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestConnectivityAdapter_ValidateConnectivity_SSH(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		target         string
		expectedResult bool
		expectedError  string
	}{
		{
			name:           "invalid host fails",
			target:         "nonexistent.invalid.host.example.com",
			expectedResult: false,
			expectedError:  "SSH connectivity test failed",
		},
		{
			name:           "git format invalid host fails",
			target:         "git@nonexistent.invalid.host.example.com",
			expectedResult: false,
			expectedError:  "SSH connectivity test failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewConnectivityAdapter(1 * time.Second) // Short timeout for tests
			ctx := context.Background()

			val := validation.ConnectivityValidation{
				Type:   validation.ConnectivityTypeSSH,
				Target: test.target,
			}

			result := adapter.ValidateConnectivity(ctx, val)

			assert.Equal(t, test.expectedResult, result.Success)
			assert.Greater(t, result.Duration, time.Duration(0))

			if test.expectedError != "" {
				require.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), test.expectedError)
			} else {
				require.NoError(t, result.Error)
			}
		})
	}
}

func TestConnectivityAdapter_ValidateConnectivity_UnsupportedType(t *testing.T) {
	t.Parallel()

	adapter := NewConnectivityAdapter(5 * time.Second)
	ctx := context.Background()

	val := validation.ConnectivityValidation{
		Type:   validation.ConnectivityType("unsupported"),
		Target: "test",
	}

	result := adapter.ValidateConnectivity(ctx, val)

	assert.False(t, result.Success)
	require.Error(t, result.Error)
	require.ErrorIs(t, result.Error, domain.ErrUnsupportedConnectivityType)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestConnectivityAdapter_ValidateConnectivity_Timeout(t *testing.T) {
	t.Parallel()

	adapter := NewConnectivityAdapter(5 * time.Second)
	ctx := context.Background()

	// Test with custom timeout
	val := validation.ConnectivityValidation{
		Type:    validation.ConnectivityTypeHTTP,
		Target:  "http://invalid.url.test",
		Timeout: 100 * time.Millisecond, // 100ms timeout for quick test failure
	}

	result := adapter.ValidateConnectivity(ctx, val)

	assert.False(t, result.Success)
	require.Error(t, result.Error)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.Less(t, result.Duration, 1*time.Second) // Should timeout quickly
}

func TestConnectivityAdapter_ValidateConnectivity_ContextCancellation(t *testing.T) {
	t.Parallel()

	adapter := NewConnectivityAdapter(5 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	val := validation.ConnectivityValidation{
		Type:   validation.ConnectivityTypeHTTP,
		Target: "http://example.com",
	}

	result := adapter.ValidateConnectivity(ctx, val)

	assert.False(t, result.Success)
	require.Error(t, result.Error)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestConnectivityAdapter_validateHTTP_InvalidURL(t *testing.T) {
	t.Parallel()

	adapter := NewConnectivityAdapter(5 * time.Second)
	ctx := context.Background()

	// Test with invalid URL
	success, err := adapter.validateHTTP(ctx, "not-a-valid-url")

	assert.False(t, success)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP connectivity check failed")
}

func TestConnectivityAdapter_validateProvider_InvalidURL(t *testing.T) {
	t.Parallel()

	adapter := NewConnectivityAdapter(5 * time.Second)
	ctx := context.Background()

	// Test with invalid URL
	success, err := adapter.validateProvider(ctx, "not-a-valid-url")

	assert.False(t, success)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider API connectivity test failed")
}

func TestConnectivityAdapter_validateSSH_HostFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		target       string
		expectedHost string
	}{
		{
			name:         "git format",
			target:       "git@nonexistent.invalid.host.test.local",
			expectedHost: "nonexistent.invalid.host.test.local:22",
		},
		{
			name:         "plain host format",
			target:       "nonexistent.invalid.host.test.local",
			expectedHost: "nonexistent.invalid.host.test.local:22",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewConnectivityAdapter(100 * time.Millisecond) // 100ms timeout for quick test failure
			ctx := context.Background()

			// We expect this to fail due to invalid host, but we can check the error message
			success, err := adapter.validateSSH(ctx, test.target)

			assert.False(t, success)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedHost)
		})
	}
}

func TestFileSystemAdapter_ValidateFileSystem_Directory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupPath       func() string
		needsWritable   bool
		expectedSuccess bool
		expectedExists  bool
		expectedError   string
	}{
		{
			name: "valid readable directory",
			setupPath: func() string {
				return t.TempDir()
			},
			needsWritable:   false,
			expectedSuccess: true,
			expectedExists:  true,
		},
		{
			name: "valid writable directory",
			setupPath: func() string {
				return t.TempDir()
			},
			needsWritable:   true,
			expectedSuccess: true,
			expectedExists:  true,
		},
		{
			name: "nonexistent directory",
			setupPath: func() string {
				return "/nonexistent/directory/path"
			},
			needsWritable:   false,
			expectedSuccess: false,
			expectedExists:  false,
			expectedError:   "failed to stat path",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewFileSystemAdapter()
			ctx := context.Background()

			path := test.setupPath()
			val := validation.FileSystemValidation{
				Type:     validation.FileSystemTypeDirectory,
				Path:     path,
				Writable: test.needsWritable,
			}

			result := adapter.ValidateFileSystem(ctx, val)

			assert.Equal(t, test.expectedSuccess, result.Success)
			assert.Equal(t, test.expectedExists, result.Exists)
			assert.Equal(t, val, result.Validation)
			assert.NotNil(t, result.Details)

			if test.expectedError != "" {
				require.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), test.expectedError)
			} else {
				require.NoError(t, result.Error)

				if test.expectedExists {
					assert.True(t, result.Readable)

					if test.needsWritable {
						assert.True(t, result.Writable)
					}
				}
			}

			// Verify details
			assert.Equal(t, path, result.Details["path"])
			assert.Equal(t, result.Exists, result.Details["exists"])
			assert.Equal(t, result.Readable, result.Details["readable"])
			assert.Equal(t, result.Writable, result.Details["writable"])
		})
	}
}

func TestFileSystemAdapter_ValidateFileSystem_File(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupPath       func() string
		needsWritable   bool
		expectedSuccess bool
		expectedExists  bool
		expectedError   string
	}{
		{
			name: "valid readable file",
			setupPath: func() string {
				tempDir := t.TempDir()
				testFile := tempDir + "/test.txt"
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0600))

				return testFile
			},
			needsWritable:   false,
			expectedSuccess: true,
			expectedExists:  true,
		},
		{
			name: "valid writable file",
			setupPath: func() string {
				tempDir := t.TempDir()
				testFile := tempDir + "/test.txt"
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0600))

				return testFile
			},
			needsWritable:   true,
			expectedSuccess: true,
			expectedExists:  true,
		},
		{
			name: "directory when expecting file",
			setupPath: func() string {
				return t.TempDir()
			},
			needsWritable:   false,
			expectedSuccess: false,
			expectedExists:  true,
			expectedError:   "path exists but is a directory",
		},
		{
			name: "nonexistent file",
			setupPath: func() string {
				return "/nonexistent/file.txt"
			},
			needsWritable:   false,
			expectedSuccess: false,
			expectedExists:  false,
			expectedError:   "failed to stat path",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewFileSystemAdapter()
			ctx := context.Background()

			path := test.setupPath()
			val := validation.FileSystemValidation{
				Type:     validation.FileSystemTypeFile,
				Path:     path,
				Writable: test.needsWritable,
			}

			result := adapter.ValidateFileSystem(ctx, val)

			assert.Equal(t, test.expectedSuccess, result.Success)
			assert.Equal(t, test.expectedExists, result.Exists)

			if test.expectedError != "" {
				require.Error(t, result.Error)
				assert.Contains(t, result.Error.Error(), test.expectedError)
			} else {
				require.NoError(t, result.Error)

				if test.expectedExists {
					assert.True(t, result.Readable)

					if test.needsWritable {
						assert.True(t, result.Writable)
					}
				}
			}
		})
	}
}

func TestFileSystemAdapter_ValidateFileSystem_Archive(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	archiveFile := tempDir + "/test.tar.gz"
	require.NoError(t, os.WriteFile(archiveFile, []byte("fake archive"), 0600))

	adapter := NewFileSystemAdapter()
	ctx := context.Background()

	val := validation.FileSystemValidation{
		Type: validation.FileSystemTypeArchive,
		Path: archiveFile,
	}

	result := adapter.ValidateFileSystem(ctx, val)

	assert.True(t, result.Success)
	assert.True(t, result.Exists)
	assert.True(t, result.Readable)
	require.NoError(t, result.Error)
}

func TestFileSystemAdapter_ValidateFileSystem_UnsupportedType(t *testing.T) {
	t.Parallel()

	adapter := NewFileSystemAdapter()
	ctx := context.Background()

	val := validation.FileSystemValidation{
		Type: validation.FileSystemType("unsupported"),
		Path: "/some/path",
	}

	result := adapter.ValidateFileSystem(ctx, val)

	assert.False(t, result.Success)
	require.Error(t, result.Error)
	require.ErrorIs(t, result.Error, domain.ErrUnsupportedFileSystemType)
}

// Benchmark tests for performance monitoring.
func BenchmarkConnectivityAdapter_ValidateHTTP(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter := NewConnectivityAdapter(5 * time.Second)
	ctx := context.Background()

	val := validation.ConnectivityValidation{
		Type:   validation.ConnectivityTypeHTTP,
		Target: server.URL,
	}

	b.ResetTimer()

	for range b.N {
		result := adapter.ValidateConnectivity(ctx, val)
		require.True(b, result.Success)
	}
}

func BenchmarkFileSystemAdapter_ValidateDirectory(b *testing.B) {
	tempDir := b.TempDir()
	adapter := NewFileSystemAdapter()
	ctx := context.Background()

	val := validation.FileSystemValidation{
		Type: validation.FileSystemTypeDirectory,
		Path: tempDir,
	}

	b.ResetTimer()

	for range b.N {
		result := adapter.ValidateFileSystem(ctx, val)
		require.True(b, result.Success)
	}
}
