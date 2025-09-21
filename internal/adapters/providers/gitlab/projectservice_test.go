// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/api/client-go"

	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	testProjectPath          = "/api/v4/projects/testuser%2Ftestrepo"
	testProjectPathAlternate = "/api/v4/projects/testuser/testrepo"
	testProjectIDPath        = "/api/v4/projects/123"
)

// Mock logger for testing.
type mockLoggerForTest struct{}

func (m *mockLoggerForTest) Debug(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLoggerForTest) Info(_ context.Context, _ string, _ map[string]any)  {}
func (m *mockLoggerForTest) Warn(_ context.Context, _ string, _ map[string]any)  {}
func (m *mockLoggerForTest) Error(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLoggerForTest) Trace(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLoggerForTest) IsLevelEnabled(_ ports.LogLevel) bool                { return false }
func (m *mockLoggerForTest) Fatal(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLoggerForTest) GetLevel() ports.LogLevel                            { return ports.LogLevelInfo }

// mockGitLabProject returns a mock GitLab project response.
func mockGitLabProject() map[string]interface{} {
	return map[string]interface{}{
		"id":             123,
		"name":           "testrepo",
		"path":           "testrepo",
		"description":    "Test description",
		"visibility":     "public",
		"default_branch": "main",
	}
}

// createUpdateTestServer creates a test server for update operations.
func createUpdateTestServer(t *testing.T, expectedField string, expectedValue interface{}, responseFields map[string]interface{}) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
		switch {
		case r.Method == http.MethodGet && (r.URL.Path == testProjectPath || r.URL.Path == testProjectPathAlternate):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(mockGitLabProject())
		case r.Method == http.MethodPut && r.URL.Path == testProjectIDPath:
			var req map[string]interface{}

			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			assert.Equal(t, expectedValue, req[expectedField])

			w.WriteHeader(http.StatusOK)

			response := map[string]interface{}{
				"id":   123,
				"name": "testrepo",
			}
			for k, v := range responseFields {
				response[k] = v
			}

			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestProjectService_UpdateProject(t *testing.T) { //nolint:gocognit,cyclop,maintidx // Table-driven test with many cases
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
			owner:    "testuser",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Description: stringPtr("Updated GitLab repository description"),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					switch {
					case r.Method == http.MethodGet && (r.URL.Path == testProjectPath || r.URL.Path == testProjectPathAlternate):
						// Return project info
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"id":          123,
							"name":        "testrepo",
							"path":        "testrepo",
							"description": "Old description",
						})
					case r.Method == http.MethodPut && r.URL.Path == testProjectIDPath:
						// Update project
						var req map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request body: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}
						assert.Equal(t, "Updated GitLab repository description", req["description"])

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"id":          123,
							"name":        "testrepo",
							"description": req["description"],
						})
					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedError: false,
		},
		{
			name:     "update visibility to private",
			owner:    "testuser",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Visibility: stringPtr("private"),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return createUpdateTestServer(t, "visibility", "private", map[string]interface{}{
					"visibility": "private",
				})
			},
			expectedError: false,
		},
		{
			name:     "update default branch",
			owner:    "testuser",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				DefaultBranch: stringPtr("develop"),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return createUpdateTestServer(t, "default_branch", "develop", map[string]interface{}{
					"default_branch": "develop",
				})
			},
			expectedError: false,
		},
		{
			name:     "update topics",
			owner:    "testuser",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Topics: []string{"golang", "gitlab", "sync"},
			},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					switch {
					case r.Method == http.MethodGet && (r.URL.Path == testProjectPath || r.URL.Path == testProjectPathAlternate):
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockGitLabProject())
					case r.Method == http.MethodPut && r.URL.Path == testProjectIDPath:
						var req map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&req)
						if err != nil {
							t.Errorf("Failed to decode request: %v", err)
							w.WriteHeader(http.StatusBadRequest)

							return
						}

						// GitLab uses "topics" field for tags
						topics, ok := req["topics"].([]interface{})
						if !ok {
							t.Errorf("Failed to get topics from request")
							w.WriteHeader(http.StatusBadRequest)

							return
						}
						assert.Len(t, topics, 3)

						w.WriteHeader(http.StatusOK)
						err = json.NewEncoder(w).Encode(map[string]interface{}{
							"id":     123,
							"name":   "testrepo",
							"topics": []string{"golang", "gitlab", "sync"},
						})
						if err != nil {
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
			name:     "handle project not found",
			owner:    "testuser",
			repoName: "nonexistent",
			updates: ports.UpdateRepositoryOptions{
				Description: stringPtr("Will fail"),
			},
			setupServer: func(_ *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					switch {
					case r.Method == http.MethodGet && (r.URL.Path == "/api/v4/projects/testuser%2Fnonexistent" || r.URL.Path == "/api/v4/projects/testuser/nonexistent"):
						// Return 404 for nonexistent project
						w.WriteHeader(http.StatusNotFound)
						_ = json.NewEncoder(w).Encode(map[string]string{
							"message": "404 Not Found",
						})
					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			expectedError: true,
			errorContains: "failed to get project for update",
		},
		{
			name:     "handle API error",
			owner:    "testuser",
			repoName: "testrepo",
			updates: ports.UpdateRepositoryOptions{
				Description: stringPtr("Will fail"),
			},
			setupServer: func(_ *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					// Always return 500 error
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Internal Server Error",
					})
				}))
			},
			expectedError: true,
			errorContains: "failed to",
		},
		{
			name:     "empty updates should not make update call",
			owner:    "testuser",
			repoName: "testrepo",
			updates:  ports.UpdateRepositoryOptions{},
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
					switch {
					case r.Method == http.MethodGet && (r.URL.Path == testProjectPath || r.URL.Path == testProjectPathAlternate):
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockGitLabProject())
					case r.Method == http.MethodPut && r.URL.Path == testProjectIDPath:
						// This should not be called for empty updates
						t.Errorf("Update API should not be called for empty updates")
						w.WriteHeader(http.StatusOK)
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

			// Create GitLab client with test server
			gitlabClient, err := gitlab.NewClient("test-token", gitlab.WithBaseURL(server.URL))
			require.NoError(t, err)

			// Create project service
			mockLogger := &mockLoggerForTest{}
			projectService := NewProjectService(gitlabClient, mockLogger)

			// Execute update
			err = projectService.UpdateProject(context.Background(), testCase.owner, testCase.repoName, testCase.updates)

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

// Test that UpdateProject correctly builds the GitLab API request.
func TestProjectService_UpdateProject_RequestBuilding(t *testing.T) {
	t.Parallel()

	requestReceived := false

	var capturedRequest map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen // Standard HTTP handler params
		switch {
		case r.Method == http.MethodGet && (r.URL.Path == "/api/v4/projects/owner%2Frepo" || r.URL.Path == "/api/v4/projects/owner/repo"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   123,
				"name": "repo",
				"path": "repo",
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v4/projects/123":
			requestReceived = true

			err := json.NewDecoder(r.Body).Decode(&capturedRequest)
			if err != nil {
				t.Errorf("Failed to decode request: %v", err)
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   123,
				"name": "repo",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	gitlabClient, err := gitlab.NewClient("test-token", gitlab.WithBaseURL(server.URL))
	require.NoError(t, err)

	mockLogger := &mockLoggerForTest{}
	projectService := NewProjectService(gitlabClient, mockLogger)

	// Test with all possible update fields
	updates := ports.UpdateRepositoryOptions{
		Description:   stringPtr("Test description"),
		Visibility:    stringPtr("private"),
		DefaultBranch: stringPtr("develop"),
		Topics:        []string{"topic1", "topic2"},
	}

	err = projectService.UpdateProject(context.Background(), "owner", "repo", updates)
	require.NoError(t, err)
	assert.True(t, requestReceived, "Request should have been made")

	// Verify the request was built correctly
	assert.Equal(t, "Test description", capturedRequest["description"])
	assert.Equal(t, "private", capturedRequest["visibility"])
	assert.Equal(t, "develop", capturedRequest["default_branch"])

	// Topics should be sent as array
	if topics, ok := capturedRequest["topics"].([]interface{}); ok {
		assert.Len(t, topics, 2)
	}
}
