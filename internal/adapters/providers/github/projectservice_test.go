// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	testRepoPath   = "/repos/testowner/testrepo"
	testTopicsPath = "/repos/testowner/testrepo/topics"
)

// mockLoggerForTest is a simple mock logger for testing.
type mockLoggerForTest struct{}

func (m *mockLoggerForTest) Debug(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLoggerForTest) Info(_ context.Context, _ string, _ map[string]any)  {}
func (m *mockLoggerForTest) Warn(_ context.Context, _ string, _ map[string]any)  {}
func (m *mockLoggerForTest) Error(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLoggerForTest) Trace(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLoggerForTest) IsLevelEnabled(_ ports.LogLevel) bool                { return false }
func (m *mockLoggerForTest) Fatal(_ context.Context, _ string, _ map[string]any) {}

func TestProjectService_UpdateProject(t *testing.T) { //nolint:gocognit,gocyclo,cyclop,maintidx // Table-driven test with many cases
	t.Parallel()

	tests := []struct {
		name          string
		owner         string
		repoName      string
		updates       ports.UpdateRepositoryOptions
		setupServer   func(t *testing.T) *httptest.Server
		expectedError bool
		errorContains string
	}{
		{
			name:     "update description successfully",
			owner:    "testowner",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Description: stringPtr("Updated repository description"),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					switch {
					case r.Method == http.MethodPatch && r.URL.Path == "/repos/testowner/testrepo":
						var req github.Repository
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}
						assert.Equal(t, "Updated repository description", *req.Description)

						w.WriteHeader(http.StatusOK)
						resp := github.Repository{
							ID:          github.Ptr(int64(123)),
							Name:        github.Ptr("testrepo"),
							Description: req.Description,
						}
						if err := json.NewEncoder(w).Encode(resp); err != nil {
							t.Errorf("Failed to encode response: %v", err)
						}
					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedError: false,
		},
		{
			name:     "update visibility to private",
			owner:    "testowner",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Visibility: stringPtr("private"),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					if r.Method == http.MethodPatch && r.URL.Path == testRepoPath {
						var req github.Repository
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}
						assert.True(t, *req.Private)

						w.WriteHeader(http.StatusOK)
						resp := github.Repository{
							ID:      github.Ptr(int64(123)),
							Name:    github.Ptr("testrepo"),
							Private: github.Ptr(true),
						}
						if err := json.NewEncoder(w).Encode(resp); err != nil {
							t.Errorf("Failed to encode response: %v", err)
						}
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedError: false,
		},
		{
			name:     "update default branch",
			owner:    "testowner",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				DefaultBranch: stringPtr("develop"),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					if r.Method == http.MethodPatch && r.URL.Path == testRepoPath {
						var req github.Repository
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}
						assert.Equal(t, "develop", *req.DefaultBranch)

						w.WriteHeader(http.StatusOK)
						resp := github.Repository{
							ID:            github.Ptr(int64(123)),
							Name:          github.Ptr("testrepo"),
							DefaultBranch: req.DefaultBranch,
						}
						if err := json.NewEncoder(w).Encode(resp); err != nil {
							t.Errorf("Failed to encode response: %v", err)
						}
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedError: false,
		},
		{
			name:     "update topics",
			owner:    "testowner",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Topics: []string{"go", "sync", "git"},
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					if r.Method == http.MethodPut && r.URL.Path == testTopicsPath {
						var req map[string][]string
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}
						assert.Equal(t, []string{"go", "sync", "git"}, req["names"])

						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(`{"names": ["go", "sync", "git"]}`))
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedError: false,
		},
		{
			name:     "update multiple fields",
			owner:    "testowner",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Description:   stringPtr("New description"),
				Visibility:    stringPtr("private"),
				DefaultBranch: stringPtr("main"),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					if r.Method == http.MethodPatch && r.URL.Path == testRepoPath {
						var req github.Repository
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}

						// Verify all fields are set
						assert.Equal(t, "New description", *req.Description)
						assert.True(t, *req.Private)
						assert.Equal(t, "main", *req.DefaultBranch)

						w.WriteHeader(http.StatusOK)
						resp := github.Repository{
							ID:            github.Ptr(int64(123)),
							Name:          github.Ptr("testrepo"),
							Description:   req.Description,
							Private:       req.Private,
							DefaultBranch: req.DefaultBranch,
						}
						if err := json.NewEncoder(w).Encode(resp); err != nil {
							t.Errorf("Failed to encode response: %v", err)
						}
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedError: false,
		},
		{
			name:     "handle repository not found",
			owner:    "testowner",
			repoName: "nonexistent",
			updates: ports.UpdateRepositoryOptions{
				Description: stringPtr("Will fail"),
			},
			setupServer: func(_ *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"message": "Not Found", "documentation_url": "https://docs.github.com/rest"}`))
				}))
			},
			expectedError: true,
			errorContains: "failed to update GitHub project",
		},
		{
			name:     "handle API rate limit",
			owner:    "testowner",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Description: stringPtr("Rate limited"),
			},
			setupServer: func(_ *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("X-RateLimit-Limit", "60")
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
				}))
			},
			expectedError: true,
			errorContains: "rate limit",
		},
		{
			name:     "empty updates should not make API call",
			owner:    "testowner",
			repoName: "testrepo",
			updates:  ports.UpdateRepositoryOptions{},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()
				callCount := 0

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					callCount++
					// This should not be called for empty updates
					t.Errorf("API should not be called for empty updates, but was called %d times", callCount)
					w.WriteHeader(http.StatusOK)
				}))
			},
			expectedError: false,
		},
		{
			name:     "update with topics and description",
			owner:    "testowner",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Description: stringPtr("Repository with topics"),
				Topics:      []string{"golang", "testing"},
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()
				callCount := 0

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					callCount++
					switch {
					case r.Method == http.MethodPatch && r.URL.Path == "/repos/testowner/testrepo":
						// First call: update description
						var req github.Repository
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}
						assert.Equal(t, "Repository with topics", *req.Description)

						w.WriteHeader(http.StatusOK)
						resp := github.Repository{
							ID:          github.Ptr(int64(123)),
							Name:        github.Ptr("testrepo"),
							Description: req.Description,
						}
						if err := json.NewEncoder(w).Encode(resp); err != nil {
							t.Errorf("Failed to encode response: %v", err)
						}
					case r.Method == http.MethodPut && r.URL.Path == testTopicsPath:
						// Second call: update topics
						var req map[string][]string
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}
						assert.Equal(t, []string{"golang", "testing"}, req["names"])

						w.WriteHeader(http.StatusOK)
						_, err = w.Write([]byte(`{"names": ["golang", "testing"]}`))
						if err != nil {
							t.Errorf("Failed to write response: %v", err)
						}
					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := testCase.setupServer(t)
			defer server.Close()

			// Create project service with test server
			config := Config{
				Token:   "test-token",
				BaseURL: server.URL + "/",
			}
			// Create GitHub client with test configuration
			githubClient := github.NewClient(nil)
			githubClient.BaseURL, _ = githubClient.BaseURL.Parse(server.URL + "/")

			if config.Token != "" {
				githubClient = githubClient.WithAuthToken(config.Token)
			}

			// Create a mock logger
			mockLogger := &mockLoggerForTest{}
			projectService := NewProjectService(githubClient, mockLogger)

			// Execute update
			err := projectService.UpdateProject(context.Background(), testCase.owner, testCase.repoName, testCase.updates)

			// Verify results
			if testCase.expectedError {
				require.Error(t, err)

				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test that UpdateProject correctly builds the GitHub API request.
func TestProjectService_UpdateProject_RequestBuilding(t *testing.T) { //nolint:cyclop // Test with multiple assertions
	t.Parallel()

	requestReceived := false

	var capturedRepoRequest *github.Repository

	var capturedTopics []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
		requestReceived = true

		if r.Method == http.MethodPatch && r.URL.Path == "/repos/owner/repo" { //nolint:nestif // Test assertions
			var req github.Repository

			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode request: %v", err)
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			capturedRepoRequest = &req

			// Respond with success
			w.WriteHeader(http.StatusOK)

			resp := github.Repository{
				ID:   github.Ptr(int64(123)),
				Name: github.Ptr("repo"),
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		} else if r.Method == http.MethodPut && r.URL.Path == "/repos/owner/repo/topics" {
			var req map[string][]string

			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode request: %v", err)
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			capturedTopics = req["names"]

			w.WriteHeader(http.StatusOK)

			_, err = w.Write([]byte(`{"names": ["topic1", "topic2"]}`))
			if err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		}
	}))
	defer server.Close()

	config := Config{
		Token:   "test-token",
		BaseURL: server.URL + "/",
	}

	// Create GitHub client with test configuration
	githubClient := github.NewClient(nil)
	githubClient.BaseURL, _ = githubClient.BaseURL.Parse(server.URL + "/")

	if config.Token != "" {
		githubClient = githubClient.WithAuthToken(config.Token)
	}

	mockLogger := &mockLoggerForTest{}
	projectService := NewProjectService(githubClient, mockLogger)

	// Test with all possible update fields
	updates := ports.UpdateRepositoryOptions{
		Description:   stringPtr("Test description"),
		Visibility:    stringPtr("private"),
		DefaultBranch: stringPtr("develop"),
		Topics:        []string{"topic1", "topic2"},
	}

	err := projectService.UpdateProject(context.Background(), "owner", "repo", updates)
	require.NoError(t, err)
	assert.True(t, requestReceived, "Request should have been made")

	// Verify the request was built correctly
	if capturedRepoRequest != nil {
		assert.Equal(t, "Test description", *capturedRepoRequest.Description)
		assert.True(t, *capturedRepoRequest.Private)
		assert.Equal(t, "develop", *capturedRepoRequest.DefaultBranch)
	}

	if len(capturedTopics) > 0 {
		assert.Equal(t, []string{"topic1", "topic2"}, capturedTopics)
	}
}
