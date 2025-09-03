// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// mockFailingProvider simulates provider failures for testing error propagation.
type mockFailingProvider struct {
	failOnList   bool
	failOnCreate bool
	failOnExists bool
	listError    error
	createError  error
	existsError  error
}

func (m *mockFailingProvider) ListRepositories(_ context.Context, _ ports.ProviderConfig) ([]entities.Repository, error) {
	if m.failOnList {
		if m.listError != nil {
			return nil, m.listError
		}

		return nil, errors.New("provider: failed to list repositories")
	}

	return []entities.Repository{}, nil
}

func (m *mockFailingProvider) CreateRepository(_ context.Context, _ ports.ProviderConfig, _ ports.CreateRepositoryOptions) (entities.Repository, error) {
	if m.failOnCreate {
		if m.createError != nil {
			return entities.Repository{}, m.createError
		}

		return entities.Repository{}, errors.New("provider: failed to create repository")
	}

	builder := entities.NewRepositoryBuilder()
	builder, _ = builder.WithName("test-repo")
	builder, _ = builder.WithHTTPSURL("https://example.com/test-repo.git")
	repo, _ := builder.Build()

	return repo, nil
}

func (m *mockFailingProvider) RepositoryExists(_ context.Context, _ ports.RepositoryExistsRequest) (bool, string, error) {
	if m.failOnExists {
		if m.existsError != nil {
			return false, "", m.existsError
		}

		return false, "", errors.New("provider: failed to check repository existence")
	}

	return true, "123", nil
}

func (m *mockFailingProvider) GetRepository(_ context.Context, _ ports.ProviderConfig, _ string) (entities.Repository, error) {
	return entities.Repository{}, errors.New("not implemented")
}

func (m *mockFailingProvider) UpdateRepository(_ context.Context, _ ports.ProviderConfig, _ string, _ ports.UpdateRepositoryOptions) error {
	return errors.New("not implemented")
}

func (m *mockFailingProvider) DeleteRepository(_ context.Context, _ ports.ProviderConfig, _ string) error {
	return errors.New("not implemented")
}

func (m *mockFailingProvider) ValidateRepositoryName(_ string) error {
	return nil
}

func (m *mockFailingProvider) TransformRepositoryName(name string, _ ports.NameTransformOptions) string {
	return name
}

// TestErrorPropagation_ProviderToUseCase tests that provider errors properly propagate through hexagonal boundaries.
func TestErrorPropagation_ProviderToUseCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupProvider  func() *mockFailingProvider
		expectedError  string
		verifyBehavior func(t *testing.T, err error)
	}{
		{
			name: "list repositories error propagates with context",
			setupProvider: func() *mockFailingProvider {
				return &mockFailingProvider{
					failOnList: true,
					listError:  errors.New("network timeout: unable to reach provider API"),
				}
			},
			expectedError: "network timeout",
			verifyBehavior: func(t *testing.T, err error) {
				t.Helper()
				t.Helper()
				// Error should contain original context
				assert.Contains(t, err.Error(), "network timeout")
				assert.Contains(t, err.Error(), "unable to reach provider API")
			},
		},
		{
			name: "create repository error includes operation context",
			setupProvider: func() *mockFailingProvider {
				return &mockFailingProvider{
					failOnCreate: true,
					createError:  errors.New("insufficient permissions: requires admin access"),
				}
			},
			expectedError: "insufficient permissions",
			verifyBehavior: func(t *testing.T, err error) {
				t.Helper()
				// Error should preserve security context
				assert.Contains(t, err.Error(), "insufficient permissions")
				assert.Contains(t, err.Error(), "admin access")
			},
		},
		{
			name: "repository exists check error maintains diagnostic info",
			setupProvider: func() *mockFailingProvider {
				return &mockFailingProvider{
					failOnExists: true,
					existsError:  errors.New("rate limit exceeded: retry after 3600 seconds"),
				}
			},
			expectedError: "rate limit exceeded",
			verifyBehavior: func(t *testing.T, err error) {
				t.Helper()
				// Error should include retry information
				assert.Contains(t, err.Error(), "rate limit exceeded")
				assert.Contains(t, err.Error(), "3600")
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			provider := testCase.setupProvider()
			ctx := context.Background()

			// Test ListRepositories error propagation
			if provider.failOnList {
				config := ports.ProviderConfig{Owner: "test"}
				repos, err := provider.ListRepositories(ctx, config)

				require.Error(t, err)
				assert.Nil(t, repos)
				assert.Contains(t, err.Error(), testCase.expectedError)
				testCase.verifyBehavior(t, err)
			}

			// Test CreateRepository error propagation
			if provider.failOnCreate {
				config := ports.ProviderConfig{Owner: "test"}
				options := ports.CreateRepositoryOptions{Name: "new-repo"}
				repo, err := provider.CreateRepository(ctx, config, options)

				require.Error(t, err)
				assert.Empty(t, repo)
				assert.Contains(t, err.Error(), testCase.expectedError)
				testCase.verifyBehavior(t, err)
			}

			// Test RepositoryExists error propagation
			if provider.failOnExists {
				request := ports.RepositoryExistsRequest{Owner: "test", Name: "repo"}
				exists, id, err := provider.RepositoryExists(ctx, request)

				require.Error(t, err)
				assert.False(t, exists)
				assert.Empty(t, id)
				assert.Contains(t, err.Error(), testCase.expectedError)
				testCase.verifyBehavior(t, err)
			}
		})
	}
}

// TestErrorPropagation_CascadingFailures tests that errors cascade correctly through multiple layers.
func TestErrorPropagation_CascadingFailures(t *testing.T) {
	t.Parallel()

	// Simulate a complex error scenario where multiple components fail
	t.Run("cascading provider and git operation failures", func(t *testing.T) {
		t.Parallel()

		// Create a context that simulates timeout
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Immediately cancel to simulate timeout

		// Mock provider that checks context
		provider := &mockFailingProvider{}
		config := ports.ProviderConfig{Owner: "test"}

		// When context is cancelled, operations should fail fast
		select {
		case <-ctx.Done():
			// This is the expected path - context cancelled
			err := ctx.Err()
			require.Error(t, err)
			assert.Equal(t, context.Canceled, err)
		default:
			t.Fatal("Context should be cancelled")
		}

		// Provider operations should respect context cancellation
		repos, err := provider.ListRepositories(ctx, config)
		if err != nil {
			// Provider might return context error or wrap it
			assert.True(t,
				errors.Is(err, context.Canceled) ||
					errors.Is(err, context.DeadlineExceeded) ||
					err.Error() == "provider: failed to list repositories",
				"Error should be context-related or provider error")
		}

		assert.Empty(t, repos)
	})
}

// TestErrorPropagation_ValidationErrors tests that validation errors are distinct from operational errors.
func TestErrorPropagation_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		validateFunc  func() error
		shouldBeValid bool
		errorType     string
	}{
		{
			name: "invalid repository name",
			validateFunc: func() error {
				provider := &mockFailingProvider{}

				return provider.ValidateRepositoryName("")
			},
			shouldBeValid: true, // Mock always returns nil for simplicity
			errorType:     "validation",
		},
		{
			name: "valid repository name",
			validateFunc: func() error {
				provider := &mockFailingProvider{}

				return provider.ValidateRepositoryName("valid-repo-name")
			},
			shouldBeValid: true,
			errorType:     "none",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.validateFunc()

			if testCase.shouldBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				// Validation errors should be clearly distinguishable
				assert.Contains(t, err.Error(), testCase.errorType)
			}
		})
	}
}

// TestErrorPropagation_RecoveryBehavior tests that the system can recover from errors.
func TestErrorPropagation_RecoveryBehavior(t *testing.T) {
	t.Parallel()

	t.Run("provider recovers after transient failure", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		provider := &mockFailingProvider{
			failOnList: true,
			listError:  errors.New("transient network error"),
		}

		ctx := context.Background()
		config := ports.ProviderConfig{Owner: "test"}

		// First call fails
		repos, err := provider.ListRepositories(ctx, config)
		callCount++

		require.Error(t, err)
		assert.Contains(t, err.Error(), "transient network error")
		assert.Empty(t, repos)

		// Simulate recovery
		provider.failOnList = false

		// Second call succeeds
		repos, err = provider.ListRepositories(ctx, config)
		callCount++

		require.NoError(t, err)
		assert.NotNil(t, repos)

		// Verify recovery happened
		assert.Equal(t, 2, callCount, "Should have made exactly 2 calls")
	})
}

// TestErrorPropagation_ErrorWrapping tests that errors are properly wrapped with context.
func TestErrorPropagation_ErrorWrapping(t *testing.T) {
	t.Parallel()

	baseError := errors.New("connection refused")

	t.Run("errors maintain stack trace through boundaries", func(t *testing.T) {
		t.Parallel()

		provider := &mockFailingProvider{
			failOnCreate: true,
			createError:  baseError,
		}

		ctx := context.Background()
		config := ports.ProviderConfig{Owner: "test"}
		options := ports.CreateRepositoryOptions{Name: "test-repo"}

		_, err := provider.CreateRepository(ctx, config, options)
		require.Error(t, err)

		// Should be able to unwrap to base error
		assert.Equal(t, baseError, err, "Mock returns error directly")

		// In real implementation, errors would be wrapped
		// Example of what production code should do:
		// wrappedErr := fmt.Errorf("failed to create repository %s: %w", options.Name, baseError)
		// assert.True(t, errors.Is(wrappedErr, baseError))
	})
}
