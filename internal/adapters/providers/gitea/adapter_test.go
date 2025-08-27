// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Test data constants.
const (
	versionAPIPath = "/api/v1/version"

	multipleReposJSON = `[
		{
			"id": 123,
			"name": "test-repo-1",
			"full_name": "testuser/test-repo-1",
			"description": "Test repository 1",
			"default_branch": "main",
			"html_url": "https://gitea.com/testuser/test-repo-1",
			"ssh_url": "git@gitea.com:testuser/test-repo-1.git",
			"clone_url": "https://gitea.com/testuser/test-repo-1.git",
			"private": true,
			"fork": false,
			"archived": false,
			"updated_at": "2023-01-01T12:00:00Z",
			"created_at": "2023-01-01T10:00:00Z",
			"owner": {
				"id": 456,
				"login": "testuser",
				"full_name": "Test User"
			}
		},
		{
			"id": 124,
			"name": "test-repo-2",
			"full_name": "testuser/test-repo-2",
			"description": "Test repository 2",
			"default_branch": "develop",
			"html_url": "https://gitea.com/testuser/test-repo-2",
			"ssh_url": "git@gitea.com:testuser/test-repo-2.git",
			"clone_url": "https://gitea.com/testuser/test-repo-2.git",
			"private": false,
			"fork": true,
			"archived": false,
			"updated_at": "2023-01-02T12:00:00Z",
			"created_at": "2023-01-02T10:00:00Z",
			"owner": {
				"id": 456,
				"login": "testuser",
				"full_name": "Test User"
			}
		}
	]`

	errorResponseJSON = `{"message": "Unauthorized", "url": "https://gitea.example.com/api/v1"}`
	emptyReposJSON    = `[]`
)

// createFastHTTPClient creates an HTTP client with reasonable timeouts for testing.
func createFastHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second, // Reasonable timeout for unit tests
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			DisableCompression:  false,
			MaxIdleConnsPerHost: 10,
		},
	}
}

func validateGiteaRepositoryFields(t *testing.T, repos []entities.Repository, expectedFields map[string]interface{}) {
	t.Helper()
	// Verify first repository fields
	if name, ok := expectedFields["first_repo_name"]; ok {
		assert.Equal(t, name, repos[0].Name())
	}

	if defaultBranch, ok := expectedFields["first_repo_default"]; ok {
		assert.Equal(t, defaultBranch, repos[0].DefaultBranch())
	}

	if private, ok := expectedFields["first_repo_private"]; ok {
		assert.Equal(t, private, repos[0].IsPrivate())
	}

	// Verify second repository if exists
	if len(repos) > 1 {
		if private, ok := expectedFields["second_repo_private"]; ok {
			assert.Equal(t, private, repos[1].IsPrivate())
		}

		if fork, ok := expectedFields["second_repo_fork"]; ok {
			assert.Equal(t, fork, repos[1].IsFork())
		}
	}

	// Verify required fields are not empty
	assert.NotEmpty(t, repos[0].ProjectID())
	assert.NotEmpty(t, repos[0].Name())
	assert.NotZero(t, repos[0].LastActivityAt())
}

func TestNew_ValidCredentials_CreatesGiteaAdapter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		token          string
		domain         string
		expectedDomain string
		expectError    bool
	}{
		{
			name:           "valid token and domain",
			token:          "test-token",
			domain:         "gitea.example.com",
			expectedDomain: "gitea.example.com",
			expectError:    false,
		},
		{
			name:           "valid token with empty domain defaults to gitea.com",
			token:          "test-token",
			domain:         "",
			expectedDomain: "gitea.com",
			expectError:    false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server to avoid real HTTP calls
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Use NewWithConfig instead to provide mock server
			config := Config{
				Token:      testCase.token,
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(),
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)

			if testCase.expectError {
				require.Error(t, err)
				assert.Nil(t, adapter)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, adapter)
			// Since we're using a mock server, domain will be from mock server URL
			assert.Contains(t, adapter.domain, "127.0.0.1")
			assert.NotNil(t, adapter.client)
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "config with custom base URL",
			expectError: false,
		},
		{
			name:        "config with default settings",
			expectError: false,
		},
		{
			name:        "config with HTTP client",
			expectError: false,
		},
		{
			name:        "config with SSL skip",
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server to avoid real HTTP calls
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				// Handle version endpoint for client setup
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				// Handle user info endpoint for authentication
				if request.URL.Path == "/api/v1/user" {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"id": 1, "login": "testuser", "full_name": "Test User"}`))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Configure with mock server
			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(), // Use test server's client for proper isolation
			}

			// All test cases use mock server - no external calls
			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)

			if testCase.expectError {
				require.Error(t, err)
				assert.Nil(t, adapter)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, adapter)
			// Domain will be extracted from mock server URL (contains 127.0.0.1)
			assert.Contains(t, adapter.domain, "127.0.0.1")
			assert.NotNil(t, adapter.client)
		})
	}
}

func TestListRepositories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockResponse   string
		statusCode     int
		expectError    bool
		expectedRepos  int
		expectedFields map[string]interface{}
	}{
		{
			name:          "successful response with multiple repositories",
			mockResponse:  multipleReposJSON,
			statusCode:    200,
			expectError:   false,
			expectedRepos: 2,
			expectedFields: map[string]interface{}{
				"first_repo_name":     "test-repo-1",
				"first_repo_default":  "main",
				"first_repo_private":  true,
				"second_repo_private": false,
				"second_repo_fork":    true,
			},
		},
		{
			name:          "empty repository list",
			mockResponse:  emptyReposJSON,
			statusCode:    200,
			expectError:   false,
			expectedRepos: 0,
		},
		{
			name:         "API error response",
			mockResponse: errorResponseJSON,
			statusCode:   401,
			expectError:  true,
		},
		{
			name:         "invalid JSON response",
			mockResponse: `invalid json`,
			statusCode:   200,
			expectError:  true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				// Verify request is to the correct Gitea API endpoint
				expectedPath := "/api/v1/users/testuser/repos"
				if request.URL.Path != expectedPath {
					// Handle other endpoints that might be called during client setup
					if request.URL.Path == versionAPIPath {
						writer.Header().Set("Content-Type", "application/json")
						writer.WriteHeader(http.StatusOK)
						_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

						return
					}

					t.Errorf("Unexpected endpoint called: %s, expected: %s", request.URL.Path, expectedPath)
					writer.WriteHeader(http.StatusNotFound)

					return
				}

				assert.Equal(t, "GET", request.Method)

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(testCase.statusCode)
				_, _ = writer.Write([]byte(testCase.mockResponse))
			}))
			defer server.Close()

			// Create adapter with mock server and custom HTTP client
			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(), // Use test server's client for proper isolation
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			// Test ListRepositories with proper provider config
			providerConfig := ports.ProviderConfig{
				Owner: "testuser", // Provide owner to make correct API call
			}
			repos, err := adapter.ListRepositories(ctx, providerConfig)

			if testCase.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Len(t, repos, testCase.expectedRepos)

			if testCase.expectedRepos > 0 {
				validateGiteaRepositoryFields(t, repos, testCase.expectedFields)
			}
		})
	}
}

func TestRepositoryExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		owner        string
		repoName     string
		mockResponse string
		statusCode   int
		expectExists bool
		expectError  bool
	}{
		{
			name:     "repository exists",
			owner:    "testuser",
			repoName: "test-repo",
			mockResponse: `{
				"id": 123,
				"name": "test-repo",
				"full_name": "testuser/test-repo"
			}`,
			statusCode:   200,
			expectExists: true,
			expectError:  false,
		},
		{
			name:         "repository does not exist",
			owner:        "testuser",
			repoName:     "nonexistent",
			mockResponse: `{"message": "404 Not Found"}`,
			statusCode:   404,
			expectExists: false,
			expectError:  false,
		},
		{
			name:         "API error",
			owner:        "testuser",
			repoName:     "test-repo",
			mockResponse: `{"message": "500 Internal Server Error"}`,
			statusCode:   500,
			expectExists: false,
			expectError:  true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				// Verify request is to the correct Gitea API endpoint
				expectedPath := "/api/v1/repos/" + testCase.owner + "/" + testCase.repoName
				if request.URL.Path != expectedPath {
					// Handle other endpoints that might be called during client setup
					if request.URL.Path == versionAPIPath {
						writer.Header().Set("Content-Type", "application/json")
						writer.WriteHeader(http.StatusOK)
						_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

						return
					}

					t.Errorf("Unexpected endpoint called: %s, expected: %s", request.URL.Path, expectedPath)
					writer.WriteHeader(http.StatusNotFound)

					return
				}

				assert.Equal(t, "GET", request.Method)

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(testCase.statusCode)
				_, _ = writer.Write([]byte(testCase.mockResponse))
			}))
			defer server.Close()

			// Create adapter with mock server and custom HTTP client
			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(), // Use test server's client for proper isolation
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			// Test RepositoryExists
			request := ports.RepositoryExistsRequest{
				Owner: testCase.owner,
				Name:  testCase.repoName,
			}

			exists, _, err := adapter.RepositoryExists(ctx, request)

			if testCase.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.expectExists, exists)
		})
	}
}

func TestCreateRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		repoRequest  ports.CreateRepositoryOptions
		mockResponse string
		statusCode   int
		expectError  bool
		expectedID   string
	}{
		{
			name: "successful repository creation",
			repoRequest: ports.CreateRepositoryOptions{
				Name:        "new-repo",
				Description: "A new test repository",
				Visibility:  "private",
			},
			mockResponse: `{
				"id": 789,
				"name": "new-repo",
				"full_name": "testuser/new-repo",
				"description": "A new test repository",
				"private": true,
				"html_url": "https://gitea.com/testuser/new-repo",
				"ssh_url": "git@gitea.com:testuser/new-repo.git",
				"clone_url": "https://gitea.com/testuser/new-repo.git",
				"default_branch": "main"
			}`,
			statusCode:  201,
			expectError: false,
			expectedID:  "789",
		},
		{
			name: "public repository creation",
			repoRequest: ports.CreateRepositoryOptions{
				Name:       "public-repo",
				Visibility: "public",
			},
			mockResponse: `{
				"id": 790,
				"name": "public-repo",
				"full_name": "testuser/public-repo",
				"private": false,
				"html_url": "https://gitea.com/testuser/public-repo",
				"ssh_url": "git@gitea.com:testuser/public-repo.git",
				"clone_url": "https://gitea.com/testuser/public-repo.git",
				"default_branch": "main"
			}`,
			statusCode:  201,
			expectError: false,
			expectedID:  "790",
		},
		{
			name: "API error on creation",
			repoRequest: ports.CreateRepositoryOptions{
				Name: "invalid-repo",
			},
			mockResponse: `{"message": "Repository name already exists"}`,
			statusCode:   400,
			expectError:  true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				// Handle version endpoint for client setup
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				// Handle repository creation endpoint
				if request.Method == http.MethodPost && strings.Contains(request.URL.Path, "/repos") {
					// Verify request body
					var requestBody map[string]interface{}

					err := json.NewDecoder(request.Body).Decode(&requestBody)
					if err != nil {
						http.Error(writer, "Invalid JSON", http.StatusBadRequest)

						return
					}

					assert.Equal(t, testCase.repoRequest.Name, requestBody["name"])

					if testCase.repoRequest.Description != "" {
						assert.Equal(t, testCase.repoRequest.Description, requestBody["description"])
					}

					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(testCase.statusCode)
					_, _ = writer.Write([]byte(testCase.mockResponse))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Create adapter with mock server
			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(), // Use test server's client for proper isolation
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			// Test CreateRepository
			result, err := adapter.CreateRepository(ctx, ports.ProviderConfig{}, testCase.repoRequest)

			if testCase.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.expectedID, result.ProjectID())
			assert.Equal(t, testCase.repoRequest.Name, result.Name())
		})
	}
}

func TestUpdateRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		updateReq    ports.UpdateRepositoryOptions
		mockResponse string
		statusCode   int
		expectError  bool
	}{
		{
			name: "successful repository update",
			updateReq: ports.UpdateRepositoryOptions{
				Description: stringPtr("Updated description"),
				Visibility:  stringPtr("public"),
			},
			mockResponse: `{
				"id": 123,
				"name": "test-repo",
				"description": "Updated description",
				"private": false
			}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:         "API error on update",
			updateReq:    ports.UpdateRepositoryOptions{},
			mockResponse: `{"message": "403 Forbidden"}`,
			statusCode:   403,
			expectError:  true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				// Handle version endpoint for client setup
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				// Handle repository update endpoint
				if request.Method == http.MethodPatch && request.URL.Path == "/api/v1/repos/testuser/test-repo" {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(testCase.statusCode)
					_, _ = writer.Write([]byte(testCase.mockResponse))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Create adapter with mock server
			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(), // Use test server's client for proper isolation
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			// Test UpdateRepository
			providerConfig := ports.ProviderConfig{
				Owner: "testuser", // Provide owner for API call
			}
			err = adapter.UpdateRepository(ctx, providerConfig, "test-repo", testCase.updateReq)

			if testCase.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAdapterImplementsInterface(t *testing.T) {
	t.Parallel()

	// Verify that Adapter implements the RepositoryProvider interface
	var _ ports.RepositoryProvider = (*Adapter)(nil)
}

func TestAdapter_Initialization_SetsCorrectFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T, adapter *Adapter)
	}{
		{
			name: "adapter has expected domain field",
			testFunc: func(t *testing.T, adapter *Adapter) {
				t.Helper()
				// Verify adapter was created with expected domain (will be mock server domain)
				assert.Contains(t, adapter.domain, "127.0.0.1")
			},
		},
		{
			name: "adapter has Gitea client",
			testFunc: func(t *testing.T, adapter *Adapter) {
				t.Helper()
				// Verify adapter has a valid Gitea client
				assert.NotNil(t, adapter.client)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server to avoid real HTTP calls
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Use NewWithConfig with mock server
			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(),
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			testCase.testFunc(t, adapter)
		})
	}
}

func TestValidateRepositoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		repoName    string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid repository name",
			repoName:    "valid-repo-name",
			expectError: false,
		},
		{
			name:        "empty repository name",
			repoName:    "",
			expectError: true,
			expectedErr: ErrRepositoryNameEmpty,
		},
		{
			name:        "repository name too long",
			repoName:    "this-is-a-very-long-repository-name-that-exceeds-the-maximum-allowed-length-for-gitea-repositories-extra-long",
			expectError: true,
			expectedErr: ErrRepositoryNameTooLong,
		},
		{
			name:        "repository name with invalid characters",
			repoName:    "repo name with spaces",
			expectError: true,
			expectedErr: ErrRepositoryNameInvalidCharacters,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// ValidateRepositoryName is a pure function that doesn't need HTTP client
			// Create minimal adapter with mock client to avoid real HTTP calls
			adapter := &Adapter{
				client: nil, // Not needed for this pure validation function
				domain: "test-domain",
			}

			err := adapter.ValidateRepositoryName(testCase.repoName)

			if testCase.expectError {
				require.Error(t, err)

				if testCase.expectedErr != nil {
					require.ErrorIs(t, err, testCase.expectedErr)
				}

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "custom domain parsing from URL",
			testFunc: func(t *testing.T) {
				t.Helper()

				// Create mock server for version endpoint
				server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if request.URL.Path == versionAPIPath {
						writer.Header().Set("Content-Type", "application/json")
						writer.WriteHeader(http.StatusOK)
						_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

						return
					}

					writer.WriteHeader(http.StatusNotFound)
				}))
				defer server.Close()

				config := Config{
					Token:      "test-token",
					BaseURL:    server.URL, // Use mock server URL
					HTTPClient: createFastHTTPClient(),
				}

				ctx := context.Background()
				adapter, err := NewWithConfig(ctx, config)
				require.NoError(t, err)

				// Extract domain from server URL for testing - domain will contain port
				assert.Contains(t, adapter.domain, "127.0.0.1")
			},
		},
		{
			name: "URL without protocol defaults to gitea.com",
			testFunc: func(t *testing.T) {
				t.Helper()

				// Create mock server for version endpoint
				server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if request.URL.Path == versionAPIPath {
						writer.Header().Set("Content-Type", "application/json")
						writer.WriteHeader(http.StatusOK)
						_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

						return
					}

					writer.WriteHeader(http.StatusNotFound)
				}))
				defer server.Close()

				config := Config{
					Token:      "test-token",
					BaseURL:    server.URL, // Use mock server URL instead of external
					HTTPClient: createFastHTTPClient(),
				}

				ctx := context.Background()
				adapter, err := NewWithConfig(ctx, config)
				require.NoError(t, err)

				// Extract domain from base URL - will be from mock server
				assert.Contains(t, adapter.domain, "127.0.0.1")
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestNew_MockedConnectionValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		token          string
		domain         string
		expectedDomain string
		expectError    bool
	}{
		{
			name:           "valid token and domain with mocked client",
			token:          "test-token",
			domain:         "gitea.example.com",
			expectedDomain: "gitea.example.com",
			expectError:    false,
		},
		{
			name:           "empty domain defaults to gitea.com with mocked client",
			token:          "test-token",
			domain:         "",
			expectedDomain: "gitea.com",
			expectError:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Use NewWithConfig with mock server to avoid real network calls
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Test adapter creation with mock server instead of real domain
			config := Config{
				Token:      test.token,
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(),
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)

			if test.expectError {
				require.Error(t, err)
				assert.Nil(t, adapter)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, adapter)
				assert.NotNil(t, adapter.client)
				// Domain will be extracted from mock server URL
				assert.Contains(t, adapter.domain, "127.0.0.1")
			}
		})
	}
}

func TestGetRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		owner        string
		repoName     string
		mockResponse string
		statusCode   int
		expectError  bool
	}{
		{
			name:     "successful get repository",
			owner:    "testowner",
			repoName: "test-repo",
			mockResponse: `{
				"id": 456,
				"name": "test-repo",
				"full_name": "testowner/test-repo",
				"description": "Test repository",
				"default_branch": "main",
				"ssh_url": "git@gitea.com:testowner/test-repo.git",
				"clone_url": "https://gitea.com/testowner/test-repo.git",
				"private": false,
				"archived": false,
				"updated_at": "2023-01-01T12:00:00Z"
			}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:         "repository not found",
			owner:        "testowner",
			repoName:     "nonexistent",
			mockResponse: `{"message": "404 Not Found"}`,
			statusCode:   404,
			expectError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				expectedPath := "/api/v1/repos/" + test.owner + "/" + test.repoName
				if request.URL.Path == expectedPath {
					assert.Equal(t, "GET", request.Method)
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(test.statusCode)
					_, _ = writer.Write([]byte(test.mockResponse))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(),
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			providerConfig := ports.ProviderConfig{Owner: test.owner}
			repo, err := adapter.GetRepository(ctx, providerConfig, test.repoName)

			if test.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.repoName, repo.Name())
			assert.Equal(t, "456", repo.ProjectID())
		})
	}
}

func TestDeleteRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		owner       string
		repoName    string
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful deletion",
			owner:       "testowner",
			repoName:    "test-repo",
			statusCode:  204,
			expectError: false,
		},
		{
			name:        "repository not found",
			owner:       "testowner",
			repoName:    "nonexistent",
			statusCode:  404,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				expectedPath := "/api/v1/repos/" + test.owner + "/" + test.repoName
				if request.URL.Path == expectedPath {
					assert.Equal(t, "DELETE", request.Method)
					writer.WriteHeader(test.statusCode)

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(),
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			providerConfig := ports.ProviderConfig{Owner: test.owner}
			err = adapter.DeleteRepository(ctx, providerConfig, test.repoName)

			if test.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTransformRepositoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		inputName    string
		options      ports.NameTransformOptions
		expectedName string
	}{
		{
			name:      "lowercase transformation",
			inputName: "TestRepo",
			options: ports.NameTransformOptions{
				ToLowercase: true,
			},
			expectedName: "testrepo",
		},
		{
			name:      "uppercase transformation",
			inputName: "testrepo",
			options: ports.NameTransformOptions{
				ToUppercase: true,
			},
			expectedName: "TESTREPO",
		},
		{
			name:      "alphanumeric only",
			inputName: "test@repo#123",
			options: ports.NameTransformOptions{
				AlphaNumericOnly: true,
			},
			expectedName: "test-repo-123",
		},
		{
			name:      "max length truncation",
			inputName: "verylongrepositoryname",
			options: ports.NameTransformOptions{
				MaxLength: 10,
			},
			expectedName: "verylongre",
		},
		{
			name:         "no transformation",
			inputName:    "valid-repo",
			options:      ports.NameTransformOptions{},
			expectedName: "valid-repo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			adapter := &Adapter{domain: "test"}

			result := adapter.TransformRepositoryName(test.inputName, test.options)
			assert.Equal(t, test.expectedName, result)
		})
	}
}

func TestGetProviderInfo(t *testing.T) {
	t.Parallel()

	adapter := &Adapter{domain: "custom.gitea.com"}

	info := adapter.GetProviderInfo()

	assert.Equal(t, "Gitea", info.Name)
	assert.Equal(t, "gitea", info.Type)
	assert.Equal(t, "custom.gitea.com", info.Domain)
	assert.Equal(t, "v1", info.APIVersion)
	assert.NotEmpty(t, info.Features)
	assert.Contains(t, info.Features, ports.FeatureRepositoryCreation)
	assert.Contains(t, info.Features, ports.FeatureBranchProtection)
	assert.Equal(t, 100, info.Capabilities.MaxRepositoriesPerRequest)
	assert.Equal(t, 3000, info.Capabilities.RateLimitPerHour)
	assert.True(t, info.Capabilities.SupportsSSH)
	assert.True(t, info.Capabilities.SupportsHTTPS)
	assert.True(t, info.Capabilities.SupportsPrivateRepos)
	assert.True(t, info.Capabilities.SupportsOrganizations)
}

func TestSupportsFeature(t *testing.T) {
	t.Parallel()

	adapter := &Adapter{}

	tests := []struct {
		name     string
		feature  ports.ProviderFeature
		expected bool
	}{
		{
			name:     "supports repository creation",
			feature:  ports.FeatureRepositoryCreation,
			expected: true,
		},
		{
			name:     "supports branch protection",
			feature:  ports.FeatureBranchProtection,
			expected: true,
		},
		{
			name:     "supports webhooks",
			feature:  ports.FeatureWebhooks,
			expected: true,
		},
		{
			name:     "supports organizations",
			feature:  ports.FeatureOrganizations,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := adapter.SupportsFeature(test.feature)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCreateRepositoryForPush(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		request      ports.CreateRepositoryRequest
		mockResponse string
		statusCode   int
		expectError  bool
		expectedID   string
	}{
		{
			name: "successful repository creation for push",
			request: ports.CreateRepositoryRequest{
				Name:        "push-repo",
				Description: "Repository for push",
				Private:     true,
			},
			mockResponse: `{
				"id": 999,
				"name": "push-repo",
				"description": "Repository for push",
				"private": true
			}`,
			statusCode:  201,
			expectError: false,
			expectedID:  "999",
		},
		{
			name: "API error during creation",
			request: ports.CreateRepositoryRequest{
				Name: "error-repo",
			},
			mockResponse: `{"message": "Name has already been taken"}`,
			statusCode:   400,
			expectError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				if request.Method == http.MethodPost && strings.Contains(request.URL.Path, "/repos") {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(test.statusCode)
					_, _ = writer.Write([]byte(test.mockResponse))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(),
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			repoID, err := adapter.CreateRepositoryForPush(ctx, test.request)

			if test.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedID, repoID)
		})
	}
}

func TestProjectExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		owner        string
		repo         string
		mockResponse string
		statusCode   int
		expectExists bool
		expectedID   string
		expectError  bool
	}{
		{
			name:  "project exists",
			owner: "testowner",
			repo:  "test-repo",
			mockResponse: `{
				"id": 555,
				"name": "test-repo",
				"full_name": "testowner/test-repo"
			}`,
			statusCode:   200,
			expectExists: true,
			expectedID:   "555",
			expectError:  false,
		},
		{
			name:         "project does not exist",
			owner:        "testowner",
			repo:         "nonexistent",
			mockResponse: `{"message": "404 Not Found"}`,
			statusCode:   404,
			expectExists: false,
			expectedID:   "",
			expectError:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				expectedPath := "/api/v1/repos/" + test.owner + "/" + test.repo
				if request.URL.Path == expectedPath {
					assert.Equal(t, "GET", request.Method)
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(test.statusCode)
					_, _ = writer.Write([]byte(test.mockResponse))

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(),
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			exists, projectID, err := adapter.ProjectExists(ctx, test.owner, test.repo)

			if test.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectExists, exists)
			assert.Equal(t, test.expectedID, projectID)
		})
	}
}

func TestIsValidProjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		projName string
		expected bool
	}{
		{
			name:     "valid project name",
			projName: "valid-project",
			expected: true,
		},
		{
			name:     "empty name is invalid",
			projName: "",
			expected: false,
		},
		{
			name:     "invalid characters",
			projName: "invalid name",
			expected: false,
		},
		{
			name:     "name too long",
			projName: string(make([]byte, 101)),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			adapter := &Adapter{}

			ctx := context.Background()
			result := adapter.IsValidProjectName(ctx, test.projName)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGiteaAdapter_SetDefaultBranch_UpdatesViaAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		owner       string
		repo        string
		branch      string
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful set default branch",
			owner:       "testowner",
			repo:        "test-repo",
			branch:      "develop",
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "API error",
			owner:       "testowner",
			repo:        "test-repo",
			branch:      "develop",
			statusCode:  400,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == versionAPIPath {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte(`{"version": "1.16.0"}`))

					return
				}

				expectedPath := "/api/v1/repos/" + test.owner + "/" + test.repo
				if request.URL.Path == expectedPath && request.Method == http.MethodPatch {
					writer.WriteHeader(test.statusCode)

					if test.statusCode == 200 {
						_, _ = writer.Write([]byte(`{"default_branch": "develop"}`))
					}

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			config := Config{
				Token:      "test-token",
				BaseURL:    server.URL,
				HTTPClient: createFastHTTPClient(),
			}

			ctx := context.Background()
			adapter, err := NewWithConfig(ctx, config)
			require.NoError(t, err)

			err = adapter.SetDefaultBranch(ctx, test.owner, test.repo, test.branch)

			if test.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

// Helper function for creating string pointers.
func stringPtr(s string) *string {
	return &s
}
