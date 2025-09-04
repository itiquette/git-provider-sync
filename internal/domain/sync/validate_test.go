// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/generated/mocks/mockhexagonal"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
)

// TestValidateSyncResponse_Structure tests the response structure.
func TestValidateSyncResponse_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response sync.ValidateSyncResponse
		check    func(t *testing.T, resp sync.ValidateSyncResponse)
	}{
		{
			name: "valid configuration response",
			response: sync.ValidateSyncResponse{
				Valid:              true,
				Errors:             []sync.ValidationError{},
				Warnings:           []sync.ValidationWarning{},
				RepositoryCount:    10,
				EstimatedDuration:  "5m",
				RecommendedActions: []string{},
			},
			check: func(t *testing.T, resp sync.ValidateSyncResponse) {
				t.Helper()
				assert.True(t, resp.Valid)
				assert.Empty(t, resp.Errors)
				assert.Equal(t, 10, resp.RepositoryCount)
			},
		},
		{
			name: "invalid with errors",
			response: sync.ValidateSyncResponse{
				Valid: false,
				Errors: []sync.ValidationError{
					{
						Type:    sync.ErrorTypeConfiguration,
						Message: "Invalid provider",
						Field:   "provider_type",
					},
				},
			},
			check: func(t *testing.T, resp sync.ValidateSyncResponse) {
				t.Helper()
				assert.False(t, resp.Valid)
				assert.Len(t, resp.Errors, 1)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.check(t, testCase.response)
		})
	}
}

// TestValidateSyncUseCase_BasicValidation tests basic validation scenarios.
func TestValidateSyncUseCase_BasicValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMock     func(*mockhexagonal.RepositoryProvider)
		request       sync.ValidateSyncRequest
		expectValid   bool
		expectErrors  int
		expectWarning bool
	}{
		{
			name: "valid configuration",
			setupMock: func(m *mockhexagonal.RepositoryProvider) {
				m.On("ListRepositories", mock.Anything, mock.Anything).
					Return([]entities.Repository{}, nil).Once()
			},
			request: sync.ValidateSyncRequest{
				SourceConfig: ports.ProviderConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
					AuthConfig: ports.AuthenticationConfig{
						Token: "test-token",
					},
				},
				MirrorTargets: []entities.MirrorTarget{
					entities.NewMirrorTarget(
						"backup", "directory", "", "", filepath.Join("testdata", "backup"),
						entities.NewAuthConfigWithToken("", ""),
						true,
					),
				},
			},
			expectValid:   true,
			expectErrors:  0,
			expectWarning: false,
		},
		{
			name: "missing owner",
			setupMock: func(m *mockhexagonal.RepositoryProvider) {
				// Mock should handle any calls even with invalid config
				m.On("ListRepositories", mock.Anything, mock.Anything).
					Return([]entities.Repository{}, nil).Maybe()
			},
			request: sync.ValidateSyncRequest{
				SourceConfig: ports.ProviderConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "", // Missing owner
					AuthConfig: ports.AuthenticationConfig{
						Token: "test-token",
					},
				},
			},
			expectValid:   false,
			expectErrors:  3, // missing owner, invalid format, no mirrors
			expectWarning: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockProvider := new(mockhexagonal.RepositoryProvider)

			testCase.setupMock(mockProvider)
			defer mockProvider.AssertExpectations(t)

			useCase := sync.NewValidateSyncUseCase(mockProvider, nil)

			response, err := useCase.Execute(context.Background(), testCase.request)
			require.NoError(t, err)

			assert.Equal(t, testCase.expectValid, response.Valid, "Validation result mismatch")
			assert.Len(t, response.Errors, testCase.expectErrors, "Error count mismatch")

			if testCase.expectWarning {
				assert.NotEmpty(t, response.Warnings, "Expected warnings")
			}
		})
	}
}
