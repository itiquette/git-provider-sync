// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	httpGET  = "GET"
	httpPOST = "POST"
)

// String pointer factory.
func stringPtr(s string) *string {
	return &s
}

// Creates test server for successful repository creation.
func createRepositoryForPushServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/orgs/testowner"):
			// Mock organization check
			writer.WriteHeader(http.StatusNotFound)
		case req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/user/repos"):
			// Mock repository creation
			writer.WriteHeader(http.StatusCreated)

			response := `{
				"id": 123,
				"name": "test-repo",
				"clone_url": "https://github.com/testowner/test-repo.git"
			}`
			_, _ = writer.Write([]byte(response))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Creates test server for user repository creation.
func createUserRepositoryServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/user/repos") {
			// Mock user repository creation
			writer.WriteHeader(http.StatusCreated)

			response := `{
				"id": 456,
				"name": "user-repo",
				"clone_url": "https://github.com/testuser/user-repo.git"
			}`
			_, _ = writer.Write([]byte(response))
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Creates test server for successful protection.
func createSuccessfulProtectionServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/repositories/123"):
			// Mock repository lookup by ID
			writer.WriteHeader(http.StatusOK)

			response := `{"id": 123, "name": "test-repo", "owner": {"login": "testowner"}}`
			_, _ = writer.Write([]byte(response))
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/protection"):
			writer.WriteHeader(http.StatusOK)

			response := `{"required_status_checks": {"strict": true}}`
			_, _ = writer.Write([]byte(response))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Creates test server for failed protection.
func createFailedProtectionServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/repositories/999") {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Creates test server for successful unprotection.
func createSuccessfulUnprotectionServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/repositories/123"):
			// Mock repository lookup by ID
			writer.WriteHeader(http.StatusOK)

			response := `{"id": 123, "name": "test-repo", "owner": {"login": "testowner"}}`
			_, _ = writer.Write([]byte(response))
		case req.Method == http.MethodDelete && strings.Contains(req.URL.Path, "/protection"):
			writer.WriteHeader(http.StatusNoContent)
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Creates test server for failed unprotection.
func createFailedUnprotectionServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/repositories/999") {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Test helper functions for mock servers.
func createSuccessfulRepoCreationServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == httpGET && strings.Contains(request.URL.Path, "/orgs/testowner"):
			// Mock organization check - return 404 to indicate it's not an org
			writer.WriteHeader(http.StatusNotFound)

			return
		case request.Method == httpPOST && strings.Contains(request.URL.Path, "/user/repos"):
			// Parse request body
			var reqBody github.Repository

			err := json.NewDecoder(request.Body).Decode(&reqBody)
			if err != nil {
				t.Error(err)

				return
			}

			// Verify request data
			assert.Equal(t, "new-repo", *reqBody.Name)
			assert.Equal(t, "Test repository", *reqBody.Description)
			assert.False(t, *reqBody.Private)

			// Mock response
			repo := github.Repository{
				ID:        github.Ptr(int64(123)),
				Name:      reqBody.Name,
				FullName:  github.Ptr("testowner/new-repo"),
				HTMLURL:   github.Ptr("https://github.com/testowner/new-repo"),
				CloneURL:  github.Ptr("https://github.com/testowner/new-repo.git"),
				SSHURL:    github.Ptr("git@github.com:testowner/new-repo.git"),
				Private:   reqBody.Private,
				CreatedAt: &github.Timestamp{Time: time.Now()},
				UpdatedAt: &github.Timestamp{Time: time.Now()},
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(writer).Encode(repo)
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

func createRepoExistsErrorServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == httpGET && strings.Contains(request.URL.Path, "/orgs/testowner"):
			// Mock organization check - return 404 to indicate it's not an org
			writer.WriteHeader(http.StatusNotFound)

			return
		case request.Method == httpPOST && strings.Contains(request.URL.Path, "/user/repos"):
			writer.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = writer.Write([]byte(`{"message": "Repository creation failed.", "errors": [{"field": "name", "code": "already_exists"}]}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestListRepositories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		config        ports.ProviderConfig
		expectedRepos int
		expectedError bool
		errorContains string
	}{
		{
			name: "successful repository listing",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					// Verify request parameters
					assert.Equal(t, httpGET, req.Method)
					assert.Contains(t, req.URL.Path, "/users/testowner/repos")

					// Mock GitHub API response
					repos := []github.Repository{
						{
							ID:       github.Ptr(int64(1)),
							Name:     github.Ptr("repo1"),
							FullName: github.Ptr("testowner/repo1"),
							HTMLURL:  github.Ptr("https://github.com/testowner/repo1"),
							CloneURL: github.Ptr("https://github.com/testowner/repo1.git"),
							SSHURL:   github.Ptr("git@github.com:testowner/repo1.git"),
							Private:  github.Ptr(false),
							Fork:     github.Ptr(false),
							Archived: github.Ptr(false),
							UpdatedAt: &github.Timestamp{
								Time: time.Now().Add(-24 * time.Hour),
							},
						},
						{
							ID:       github.Ptr(int64(2)),
							Name:     github.Ptr("repo2"),
							FullName: github.Ptr("testowner/repo2"),
							HTMLURL:  github.Ptr("https://github.com/testowner/repo2"),
							CloneURL: github.Ptr("https://github.com/testowner/repo2.git"),
							SSHURL:   github.Ptr("git@github.com:testowner/repo2.git"),
							Private:  github.Ptr(true),
							Fork:     github.Ptr(false),
							Archived: github.Ptr(false),
							UpdatedAt: &github.Timestamp{
								Time: time.Now().Add(-48 * time.Hour),
							},
						},
					}

					writer.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(writer).Encode(repos)
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
				AuthConfig: ports.AuthenticationConfig{
					Token: "test-token",
				},
			},
			expectedRepos: 2,
			expectedError: false,
		},
		{
			name: "API error response",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusUnauthorized)
					_, _ = writer.Write([]byte(`{"message": "Bad credentials"}`))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
				AuthConfig: ports.AuthenticationConfig{
					Token: "invalid-token",
				},
			},
			expectedRepos: 0,
			expectedError: true,
			errorContains: "401",
		},
		{
			name: "empty repository list",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(writer).Encode([]github.Repository{})
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
				AuthConfig: ports.AuthenticationConfig{
					Token: "test-token",
				},
			},
			expectedRepos: 0,
			expectedError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := testCase.setupServer()
			defer server.Close()

			// Create adapter with mock server
			config := Config{
				Token:   testCase.config.AuthConfig.Token,
				BaseURL: server.URL + "/",
			}
			adapter := NewWithConfig(context.Background(), config)

			// Execute test
			repos, err := adapter.ListRepositories(context.Background(), testCase.config)

			// Verify results
			if testCase.expectedError {
				require.Error(t, err)

				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, repos, testCase.expectedRepos)

				// Verify repository structure
				for _, repo := range repos {
					assert.NotEmpty(t, repo.Name())
					assert.NotEmpty(t, repo.HTTPSURL())
					assert.NotEmpty(t, repo.SSHURL())
					assert.NotNil(t, repo.LastActivityAt())
				}
			}
		})
	}
}

func TestRepositoryExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		request        ports.RepositoryExistsRequest
		expectedExists bool
		expectedURL    string
		expectedError  bool
	}{
		{
			name: "repository exists",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					if strings.Contains(req.URL.Path, "/repos/testowner/existing-repo") {
						repo := github.Repository{
							ID:       github.Ptr(int64(1)),
							Name:     github.Ptr("existing-repo"),
							FullName: github.Ptr("testowner/existing-repo"),
							HTMLURL:  github.Ptr("https://github.com/testowner/existing-repo"),
							CloneURL: github.Ptr("https://github.com/testowner/existing-repo.git"),
						}
						writer.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(writer).Encode(repo)
					} else {
						writer.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			request: ports.RepositoryExistsRequest{
				Owner: "testowner",
				Name:  "existing-repo",
			},
			expectedExists: true,
			expectedURL:    "1",
			expectedError:  false,
		},
		{
			name: "repository does not exist",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
				}))
			},
			request: ports.RepositoryExistsRequest{
				Owner: "testowner",
				Name:  "nonexistent-repo",
			},
			expectedExists: false,
			expectedURL:    "",
			expectedError:  false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := testCase.setupServer()
			defer server.Close()

			// Create adapter with mock server
			config := Config{
				Token:   "test-token",
				BaseURL: server.URL + "/",
			}
			adapter := NewWithConfig(context.Background(), config)

			// Execute test
			exists, url, err := adapter.RepositoryExists(context.Background(), testCase.request)

			// Verify results
			if testCase.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedExists, exists)
				assert.Equal(t, testCase.expectedURL, url)
			}
		})
	}
}

func TestCreateRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		config        ports.ProviderConfig
		options       ports.CreateRepositoryOptions
		expectedError bool
		errorContains string
	}{
		{
			name: "successful repository creation",
			setupServer: func() *httptest.Server {
				return createSuccessfulRepoCreationServer(t)
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
				AuthConfig: ports.AuthenticationConfig{
					Token: "test-token",
				},
			},
			options: ports.CreateRepositoryOptions{
				Name:        "new-repo",
				Description: "Test repository",
				Visibility:  "public",
			},
			expectedError: false,
		},
		{
			name:        "repository already exists",
			setupServer: createRepoExistsErrorServer,
			config: ports.ProviderConfig{
				Owner: "testowner",
				AuthConfig: ports.AuthenticationConfig{
					Token: "test-token",
				},
			},
			options: ports.CreateRepositoryOptions{
				Name:        "existing-repo",
				Description: "Test repository",
				Visibility:  "public",
			},
			expectedError: true,
			errorContains: "422",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := testCase.setupServer()
			defer server.Close()

			// Create adapter with mock server
			config := Config{
				Token:   testCase.config.AuthConfig.Token,
				BaseURL: server.URL + "/",
			}
			adapter := NewWithConfig(context.Background(), config)

			// Execute test
			repo, err := adapter.CreateRepository(context.Background(), testCase.config, testCase.options)

			// Verify results
			if testCase.expectedError {
				require.Error(t, err)

				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, repo)
				assert.Equal(t, testCase.options.Name, repo.Name())
				assert.NotEmpty(t, repo.HTTPSURL())
				assert.NotEmpty(t, repo.SSHURL())
			}
		})
	}
}

func TestValidateRepositoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		repoName      string
		expectedError bool
	}{
		{
			name:          "valid repository name",
			repoName:      "valid-repo-name",
			expectedError: false,
		},
		{
			name:          "empty repository name",
			repoName:      "",
			expectedError: true,
		},

		{
			name:          "repository name with spaces",
			repoName:      "repo name",
			expectedError: true,
		},
		{
			name:          "repository name too long",
			repoName:      strings.Repeat("a", 101),
			expectedError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			adapter := New(context.Background(), "")
			err := adapter.ValidateRepositoryName(testCase.repoName)

			if testCase.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTransformRepositoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		options      ports.NameTransformOptions
		expectedName string
	}{
		{
			name:         "no transformation needed",
			input:        "valid-repo-name",
			options:      ports.NameTransformOptions{},
			expectedName: "valid-repo-name",
		},
		{
			name:         "alphanumeric transformation",
			input:        "repo with spaces!@#",
			options:      ports.NameTransformOptions{AlphaNumericOnly: true},
			expectedName: "repo-with-spaces---",
		},
		{
			name:         "lowercase transformation",
			input:        "RepoName",
			options:      ports.NameTransformOptions{ToLowercase: true},
			expectedName: "reponame",
		},
		{
			name:         "combined transformations",
			input:        "My Repo Name!",
			options:      ports.NameTransformOptions{AlphaNumericOnly: true, ToLowercase: true},
			expectedName: "my-repo-name-",
		},
		{
			name:         "empty input",
			input:        "",
			options:      ports.NameTransformOptions{},
			expectedName: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			adapter := New(context.Background(), "")
			result := adapter.TransformRepositoryName(testCase.input, testCase.options)
			assert.Equal(t, testCase.expectedName, result)
		})
	}
}

func TestGetRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		config        ports.ProviderConfig
		repoName      string
		expectedError bool
		errorContains string
	}{
		{
			name: "successful repository retrieval",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, httpGET, req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo")

					repo := github.Repository{
						ID:        github.Ptr(int64(1)),
						Name:      github.Ptr("test-repo"),
						FullName:  github.Ptr("testowner/test-repo"),
						HTMLURL:   github.Ptr("https://github.com/testowner/test-repo"),
						CloneURL:  github.Ptr("https://github.com/testowner/test-repo.git"),
						SSHURL:    github.Ptr("git@github.com:testowner/test-repo.git"),
						Private:   github.Ptr(false),
						UpdatedAt: &github.Timestamp{Time: time.Now()},
					}

					writer.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(writer).Encode(repo)
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
				AuthConfig: ports.AuthenticationConfig{
					Token: "test-token",
				},
			},
			repoName:      "test-repo",
			expectedError: false,
		},
		{
			name: "repository not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
				AuthConfig: ports.AuthenticationConfig{
					Token: "test-token",
				},
			},
			repoName:      "nonexistent-repo",
			expectedError: true,
			errorContains: "404",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := testCase.setupServer()
			defer server.Close()

			// Create adapter with mock server
			config := Config{
				Token:   testCase.config.AuthConfig.Token,
				BaseURL: server.URL + "/",
			}
			adapter := NewWithConfig(context.Background(), config)

			// Execute test
			repo, err := adapter.GetRepository(context.Background(), testCase.config, testCase.repoName)

			// Verify results
			if testCase.expectedError {
				require.Error(t, err)

				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, repo)
				assert.Equal(t, testCase.repoName, repo.Name())
				assert.NotEmpty(t, repo.HTTPSURL())
				assert.NotEmpty(t, repo.SSHURL())
			}
		})
	}
}

// GitHub adapter workflow tests with mocked HTTP responses.
func TestGitHubAdapterWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		owner          string
		token          string
		expectedRepos  []string
		expectOrgCheck bool
	}{
		{
			name:           "user repository workflow",
			owner:          "testuser",
			token:          "github_pat_token123",
			expectedRepos:  []string{"user-repo-1", "user-repo-2"},
			expectOrgCheck: false,
		},
		{
			name:           "organization repository workflow",
			owner:          "testorg",
			token:          "github_pat_org_token",
			expectedRepos:  []string{"org-repo-1", "org-repo-2", "org-repo-3"},
			expectOrgCheck: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create mock server with controlled responses
			server := createMockGitHubServer(t, test.owner, test.expectedRepos, test.expectOrgCheck)
			t.Cleanup(server.Close)

			adapter := createGitHubAdapterWithMockServer(t, server, test.token)
			providerConfig := ports.ProviderConfig{
				Owner: test.owner,
				AuthConfig: ports.AuthenticationConfig{
					Token: test.token,
				},
			}

			// Act & Assert: Test complete repository workflow
			t.Run("list repositories", func(t *testing.T) {
				t.Parallel()

				repos, err := adapter.ListRepositories(context.Background(), providerConfig)
				require.NoError(t, err)
				assert.Len(t, repos, len(test.expectedRepos))

				for i, expectedRepo := range test.expectedRepos {
					assert.Equal(t, expectedRepo, repos[i].Name())
				}
			})

			t.Run("get specific repository", func(t *testing.T) {
				t.Parallel()

				if len(test.expectedRepos) > 0 {
					repo, err := adapter.GetRepository(context.Background(), providerConfig, test.expectedRepos[0])
					require.NoError(t, err)
					assert.Equal(t, test.expectedRepos[0], repo.Name())
				}
			})

			t.Run("check repository exists", func(t *testing.T) {
				t.Parallel()

				if len(test.expectedRepos) > 0 {
					exists, url, err := adapter.RepositoryExists(context.Background(), ports.RepositoryExistsRequest{
						Owner: test.owner,
						Name:  test.expectedRepos[0],
					})
					require.NoError(t, err)
					assert.True(t, exists)
					assert.NotEmpty(t, url)
				}
			})

			t.Run("create new repository", func(t *testing.T) {
				t.Parallel()

				newRepo, err := adapter.CreateRepository(context.Background(), providerConfig, ports.CreateRepositoryOptions{
					Name:        "new-created-repo",
					Description: "Created via API",
					Visibility:  "public",
				})
				require.NoError(t, err)
				assert.Equal(t, "new-created-repo", newRepo.Name())
				assert.Contains(t, newRepo.HTTPSURL(), "new-created-repo")
			})
		})
	}
}

// Helper functions for mocking - moved outside main test for reusability

//nolint:cyclop // Test helper function needs to handle multiple HTTP endpoints
func createMockGitHubServer(t *testing.T, owner string, repos []string, isOrg bool) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == httpGET && strings.Contains(req.URL.Path, "/users/"+owner+"/repos"):
			mockRepos := make([]github.Repository, len(repos))
			for i, repoName := range repos {
				mockRepos[i] = github.Repository{
					ID:        github.Ptr(int64(i + 1)),
					Name:      github.Ptr(repoName),
					FullName:  github.Ptr(owner + "/" + repoName),
					HTMLURL:   github.Ptr("https://github.com/" + owner + "/" + repoName),
					CloneURL:  github.Ptr("https://github.com/" + owner + "/" + repoName + ".git"),
					SSHURL:    github.Ptr("git@github.com:" + owner + "/" + repoName + ".git"),
					Private:   github.Ptr(false),
					UpdatedAt: &github.Timestamp{Time: time.Now()},
				}
			}

			respondWithJSON(t, writer, mockRepos)

		case req.Method == httpGET && strings.Contains(req.URL.Path, "/repos/"+owner+"/"):
			repoName := extractRepoNameFromPath(req.URL.Path, owner)
			mockRepo := github.Repository{
				ID:        github.Ptr(int64(1)),
				Name:      github.Ptr(repoName),
				FullName:  github.Ptr(owner + "/" + repoName),
				HTMLURL:   github.Ptr("https://github.com/" + owner + "/" + repoName),
				CloneURL:  github.Ptr("https://github.com/" + owner + "/" + repoName + ".git"),
				SSHURL:    github.Ptr("git@github.com:" + owner + "/" + repoName + ".git"),
				Private:   github.Ptr(false),
				UpdatedAt: &github.Timestamp{Time: time.Now()},
			}
			respondWithJSON(t, writer, mockRepo)

		case req.Method == httpGET && strings.Contains(req.URL.Path, "/orgs/"+owner):
			if isOrg {
				writer.WriteHeader(http.StatusOK)
			} else {
				writer.WriteHeader(http.StatusNotFound)
			}

		case req.Method == httpPOST && strings.Contains(req.URL.Path, "/user/repos"):
			var reqBody github.Repository
			if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
				t.Error(err)

				return
			}

			mockRepo := github.Repository{
				ID:       github.Ptr(int64(999)),
				Name:     reqBody.Name,
				FullName: github.Ptr(owner + "/" + *reqBody.Name),
				HTMLURL:  github.Ptr("https://github.com/" + owner + "/" + *reqBody.Name),
				CloneURL: github.Ptr("https://github.com/" + owner + "/" + *reqBody.Name + ".git"),
				SSHURL:   github.Ptr("git@github.com:" + owner + "/" + *reqBody.Name + ".git"),
				Private:  reqBody.Private,
			}

			writer.WriteHeader(http.StatusCreated)
			respondWithJSON(t, writer, mockRepo)

		case req.Method == httpPOST && strings.Contains(req.URL.Path, "/orgs/"+owner+"/repos"):
			var reqBody github.Repository
			if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
				t.Error(err)

				return
			}

			mockRepo := github.Repository{
				ID:       github.Ptr(int64(888)),
				Name:     reqBody.Name,
				FullName: github.Ptr(owner + "/" + *reqBody.Name),
				HTMLURL:  github.Ptr("https://github.com/" + owner + "/" + *reqBody.Name),
				CloneURL: github.Ptr("https://github.com/" + owner + "/" + *reqBody.Name + ".git"),
				SSHURL:   github.Ptr("git@github.com:" + owner + "/" + *reqBody.Name + ".git"),
				Private:  reqBody.Private,
			}

			writer.WriteHeader(http.StatusCreated)
			respondWithJSON(t, writer, mockRepo)

		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
}

func createGitHubAdapterWithMockServer(t *testing.T, server *httptest.Server, token string) *Adapter {
	t.Helper()

	config := Config{
		Token:   token,
		BaseURL: server.URL + "/",
	}

	return NewWithConfig(context.Background(), config)
}

func respondWithJSON(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(data))
}

func extractRepoNameFromPath(path, owner string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == owner && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return "unknown-repo"
}

// Test UpdateRepository method (0% coverage).
func TestUpdateRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		config        ports.ProviderConfig
		repoName      string
		options       ports.UpdateRepositoryOptions
		expectedError bool
		errorContains string
	}{
		{
			name: "successful repository update",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "PATCH", req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo")

					writer.WriteHeader(http.StatusOK)
					response := `{
						"id": 1,
						"name": "test-repo",
						"description": "Updated description",
						"private": true
					}`
					_, _ = writer.Write([]byte(response))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName: "test-repo",
			options: ports.UpdateRepositoryOptions{
				Description: stringPtr("Updated description"),
				Visibility:  stringPtr("private"),
			},
			expectedError: false,
		},
		{
			name: "repository not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "nonexistent-repo",
			options:       ports.UpdateRepositoryOptions{},
			expectedError: true,
			errorContains: "failed to update repository",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			err := adapter.UpdateRepository(context.Background(), test.config, test.repoName, test.options)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test DeleteRepository method (0% coverage).
func TestDeleteRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		config        ports.ProviderConfig
		repoName      string
		expectedError bool
		errorContains string
	}{
		{
			name: "successful repository deletion",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "DELETE", req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo")
					writer.WriteHeader(http.StatusNoContent)
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "test-repo",
			expectedError: false,
		},
		{
			name: "repository not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "nonexistent-repo",
			expectedError: true,
			errorContains: "failed to delete repository",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			err := adapter.DeleteRepository(context.Background(), test.config, test.repoName)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test GetProviderInfo method (0% coverage).

// Test CreateRepositoryForPush method (0% coverage).
func TestCreateRepositoryForPush(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		request       ports.CreateRepositoryRequest
		expectedID    string
		expectedError bool
		errorContains string
	}{
		{
			name:        "successful repository creation for push",
			setupServer: createRepositoryForPushServer,
			request: ports.CreateRepositoryRequest{
				Name:        "test-repo",
				Description: "Test repository",
				Private:     true,
			},
			expectedID:    "123",
			expectedError: false,
		},
		{
			name:        "user repository creation",
			setupServer: createUserRepositoryServer,
			request: ports.CreateRepositoryRequest{
				Name:        "user-repo",
				Description: "User repository",
				Private:     false,
			},
			expectedID:    "456",
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			repoID, err := adapter.CreateRepositoryForPush(context.Background(), test.request)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedID, repoID)
			}
		})
	}
}

// Test ProjectExists method (0% coverage).
func TestProjectExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		owner          string
		repo           string
		expectedExists bool
		expectedID     string
		expectedError  bool
	}{
		{
			name: "project exists",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "GET", req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo")

					writer.WriteHeader(http.StatusOK)
					response := `{"id": 123, "name": "test-repo"}`
					_, _ = writer.Write([]byte(response))
				}))
			},
			owner:          "testowner",
			repo:           "test-repo",
			expectedExists: true,
			expectedID:     "123",
			expectedError:  false,
		},
		{
			name: "project does not exist",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
				}))
			},
			owner:          "testowner",
			repo:           "nonexistent-repo",
			expectedExists: false,
			expectedID:     "",
			expectedError:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			exists, projectID, err := adapter.ProjectExists(context.Background(), test.owner, test.repo)

			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedExists, exists)
				assert.Equal(t, test.expectedID, projectID)
			}
		})
	}
}

// Sets up test adapter with custom server URL.
func setupTestAdapter(t *testing.T, serverURL string) *Adapter {
	t.Helper()

	config := Config{
		Token:   "test-token",
		BaseURL: serverURL + "/",
	}

	return NewWithConfig(context.Background(), config)
}

// Test GetBranchProtection method (0% coverage).
func TestGetBranchProtection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		config        ports.ProviderConfig
		repoName      string
		branch        string
		expectedError bool
		errorContains string
	}{
		{
			name: "successful branch protection retrieval",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "GET", req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo/branches/main/protection")

					writer.WriteHeader(http.StatusOK)
					response := `{
						"required_status_checks": {
							"strict": true,
							"contexts": ["ci"]
						},
						"enforce_admins": {
							"enabled": true
						},
						"required_pull_request_reviews": {
							"required_approving_review_count": 2
						}
					}`
					_, _ = writer.Write([]byte(response))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "test-repo",
			branch:        "main",
			expectedError: false,
		},
		{
			name: "branch protection not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Branch not protected"}`))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "test-repo",
			branch:        "unprotected",
			expectedError: true,
			errorContains: "failed to get branch protection",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			protection, err := adapter.GetBranchProtection(context.Background(), test.config, test.repoName, test.branch)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, protection)
			}
		})
	}
}

// Test SetBranchProtection method (0% coverage).
func TestSetBranchProtection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		config        ports.ProviderConfig
		repoName      string
		branch        string
		protection    ports.BranchProtection
		expectedError bool
		errorContains string
	}{
		{
			name: "successful branch protection setting",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "PUT", req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo/branches/main/protection")

					writer.WriteHeader(http.StatusOK)
					response := `{
						"required_status_checks": {
							"strict": true,
							"contexts": ["ci"]
						}
					}`
					_, _ = writer.Write([]byte(response))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName: "test-repo",
			branch:   "main",
			protection: ports.BranchProtection{
				RequiredStatusChecks: ports.RequiredStatusChecks{
					Strict:   true,
					Contexts: []string{"ci"},
				},
			},
			expectedError: false,
		},
		{
			name: "repository not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "nonexistent-repo",
			branch:        "main",
			protection:    ports.BranchProtection{},
			expectedError: true,
			errorContains: "failed to set branch protection",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			err := adapter.SetBranchProtection(context.Background(), test.config, test.repoName, test.branch, test.protection)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test RemoveBranchProtection method (0% coverage).
func TestRemoveBranchProtection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		config        ports.ProviderConfig
		repoName      string
		branch        string
		expectedError bool
		errorContains string
	}{
		{
			name: "successful branch protection removal",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "DELETE", req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo/branches/main/protection")
					writer.WriteHeader(http.StatusNoContent)
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "test-repo",
			branch:        "main",
			expectedError: false,
		},
		{
			name: "branch protection not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Branch not protected"}`))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "test-repo",
			branch:        "unprotected",
			expectedError: true,
			errorContains: "failed to remove branch protection",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			err := adapter.RemoveBranchProtection(context.Background(), test.config, test.repoName, test.branch)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test ListProtectedBranches method (0% coverage).
func TestListProtectedBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupServer      func() *httptest.Server
		config           ports.ProviderConfig
		repoName         string
		expectedBranches []string
		expectedError    bool
		errorContains    string
	}{
		{
			name: "successful protected branches listing",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "GET", req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo/branches")

					writer.WriteHeader(http.StatusOK)
					response := `[
						{
							"name": "main",
							"protected": true,
							"protection": {
								"enabled": true
							}
						},
						{
							"name": "develop",
							"protected": true,
							"protection": {
								"enabled": true
							}
						}
					]`
					_, _ = writer.Write([]byte(response))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:         "test-repo",
			expectedBranches: []string{"main", "develop"},
			expectedError:    false,
		},
		{
			name: "repository not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
				}))
			},
			config: ports.ProviderConfig{
				Owner: "testowner",
			},
			repoName:      "nonexistent-repo",
			expectedError: true,
			errorContains: "failed to list protected branches",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			branches, err := adapter.ListProtectedBranches(context.Background(), test.config, test.repoName)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedBranches, branches)
			}
		})
	}
}

// Test Protect method (0% coverage).
func TestProtect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		owner         string
		defaultBranch string
		projectID     string
		expectedError bool
		errorContains string
	}{
		{
			name:          "successful protection",
			setupServer:   createSuccessfulProtectionServer,
			owner:         "testowner",
			defaultBranch: "main",
			projectID:     "123",
			expectedError: false,
		},
		{
			name:          "protection failure",
			setupServer:   createFailedProtectionServer,
			owner:         "testowner",
			defaultBranch: "main",
			projectID:     "999",
			expectedError: true,
			errorContains: "failed to get repository by ID",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			err := adapter.Protect(context.Background(), test.owner, test.defaultBranch, test.projectID)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test Unprotect method (0% coverage).
func TestUnprotect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		defaultBranch string
		projectID     string
		expectedError bool
		errorContains string
	}{
		{
			name:          "successful unprotection",
			setupServer:   createSuccessfulUnprotectionServer,
			defaultBranch: "main",
			projectID:     "123",
			expectedError: false,
		},
		{
			name:          "unprotection failure",
			setupServer:   createFailedUnprotectionServer,
			defaultBranch: "main",
			projectID:     "999",
			expectedError: true,
			errorContains: "failed to get repository by ID",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			err := adapter.Unprotect(context.Background(), test.defaultBranch, test.projectID)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test SetDefaultBranch method (0% coverage).
func TestGitHubAdapter_SetDefaultBranch_UpdatesViaAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		owner         string
		repoName      string
		branch        string
		expectedError bool
		errorContains string
	}{
		{
			name: "successful set default branch",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "PATCH", req.Method)
					assert.Contains(t, req.URL.Path, "/repos/testowner/test-repo")

					// Verify request body
					var reqBody github.Repository
					_ = json.NewDecoder(req.Body).Decode(&reqBody)
					assert.Equal(t, "develop", *reqBody.DefaultBranch)

					writer.WriteHeader(http.StatusOK)
					response := `{
						"id": 123,
						"name": "test-repo",
						"default_branch": "develop"
					}`
					_, _ = writer.Write([]byte(response))
				}))
			},
			owner:         "testowner",
			repoName:      "test-repo",
			branch:        "develop",
			expectedError: false,
		},
		{
			name: "repository not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
					_, _ = writer.Write([]byte(`{"message": "Not Found"}`))
				}))
			},
			owner:         "testowner",
			repoName:      "nonexistent-repo",
			branch:        "main",
			expectedError: true,
			errorContains: "failed to set default branch",
		},
		{
			name: "API error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusForbidden)
					_, _ = writer.Write([]byte(`{"message": "Insufficient permissions"}`))
				}))
			},
			owner:         "testowner",
			repoName:      "test-repo",
			branch:        "main",
			expectedError: true,
			errorContains: "failed to set default branch",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := test.setupServer()
			defer server.Close()

			adapter := setupTestAdapter(t, server.URL)

			err := adapter.SetDefaultBranch(context.Background(), test.owner, test.repoName, test.branch)

			if test.expectedError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test IsValidProjectName method (0% coverage).
func TestIsValidProjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		projName string
		expected bool
	}{
		{
			name:     "valid project name with alphanumeric and underscore",
			projName: "valid_repo123",
			expected: true,
		},
		{
			name:     "empty name is invalid",
			projName: "",
			expected: false,
		},
		{
			name:     "name with spaces is invalid",
			projName: "invalid name",
			expected: false,
		},
		{
			name:     "name starting with period is valid",
			projName: ".gitignore-repo",
			expected: true,
		},
		{
			name:     "name too long is invalid",
			projName: strings.Repeat("a", 101),
			expected: false,
		},
		{
			name:     "name with special characters is invalid",
			projName: "repo@#$",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			adapter := New(context.Background(), "test-token")

			result := adapter.IsValidProjectName(context.Background(), test.projName)
			assert.Equal(t, test.expected, result)
		})
	}
}
