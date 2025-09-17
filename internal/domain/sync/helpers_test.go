// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// TestMockSetup provides common mock setup helpers for sync tests.
type TestMockSetup struct {
	Config     *mockConfiguration
	Repo       *mockRepositoryProvider
	Git        *mockGitOperations
	Logger     *mockLogger
	Archive    *mockArchiveOperations
	FileSystem *mockFileSystem
}

// NewTestMockSetup creates a new set of initialized mocks for testing.
func NewTestMockSetup() *TestMockSetup {
	return &TestMockSetup{
		Config:     &mockConfiguration{},
		Repo:       &mockRepositoryProvider{},
		Git:        &mockGitOperations{},
		Logger:     &mockLogger{},
		Archive:    &mockArchiveOperations{},
		FileSystem: &mockFileSystem{},
	}
}

// SetupBasicExpectations sets up common mock expectations for basic operations.
func (s *TestMockSetup) SetupBasicExpectations(t *testing.T) {
	t.Helper()

	// Setup default logger expectations - accept any log level
	s.Logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
	s.Logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
	s.Logger.On("Warn", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
	s.Logger.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
	s.Logger.On("Trace", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
	s.Logger.On("IsLevelEnabled", mock.Anything).Return(true).Maybe()
}

// SetupTempDirExpectations sets up expectations for temporary directory operations.
func (s *TestMockSetup) SetupTempDirExpectations(t *testing.T, basePath string) {
	t.Helper()

	tempPath := filepath.Join(basePath, "tmp", "test")

	s.Git.On("CreateTmpDir", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(context.Background(), nil).Maybe()
	s.Git.On("GetTmpDirPath", mock.Anything).Return(tempPath, nil).Maybe()
	s.Git.On("DeleteTmpDir", mock.Anything).Return(nil).Maybe()
}

// SetupConfigurationLoad sets up configuration loading expectations with a test config.
func (s *TestMockSetup) SetupConfigurationLoad(t *testing.T, config ports.AppConfiguration) {
	t.Helper()

	s.Config.On("Load", mock.Anything, mock.AnythingOfType("ports.ConfigurationSource")).
		Return(config, nil)
}

// AssertExpectations verifies all mock expectations were met.
func (s *TestMockSetup) AssertExpectations(t *testing.T) {
	t.Helper()

	s.Config.AssertExpectations(t)
	s.Repo.AssertExpectations(t)
	s.Git.AssertExpectations(t)
	s.Logger.AssertExpectations(t)
	s.Archive.AssertExpectations(t)
	s.FileSystem.AssertExpectations(t)
}

// CreateTestEnvironmentConfig creates a basic test environment configuration.
func CreateTestEnvironmentConfig(sourceProvider, mirrorProvider string) ports.EnvironmentConfiguration {
	return ports.EnvironmentConfiguration{
		Enabled: true,
		Source: ports.SourceConfiguration{
			ProviderType: sourceProvider,
			Domain:       sourceProvider + ".example.com",
			Owner:        "test-owner",
			Authentication: ports.AuthenticationConfiguration{
				Token: "test-token",
			},
		},
		Mirrors: map[string]ports.MirrorConfiguration{
			"backup": {
				Enabled:      true,
				ProviderType: mirrorProvider,
				Path:         filepath.Join("testdata", "backup"),
			},
		},
	}
}

// SetupSyncTestWithConfig creates a complete test setup with the given configuration.
func SetupSyncTestWithConfig(t *testing.T, envName string, envConfig ports.EnvironmentConfiguration) (*TestMockSetup, ports.AppConfiguration) {
	t.Helper()

	mocks := NewTestMockSetup()
	mocks.SetupBasicExpectations(t)

	testConfig := ports.AppConfiguration{
		Environments: map[string]ports.EnvironmentConfiguration{
			envName: envConfig,
		},
	}

	mocks.SetupConfigurationLoad(t, testConfig)
	mocks.SetupTempDirExpectations(t, "testdata")

	return mocks, testConfig
}

// VerifySuccessResponse verifies a sync response indicates success.
func VerifySuccessResponse(t *testing.T, response interface {
	GetSuccess() bool
	GetErrors() []error
}, err error) {
	t.Helper()

	require.NoError(t, err)

	type successChecker interface {
		GetSuccess() bool
		GetErrors() []error
	}

	if resp, ok := response.(successChecker); ok {
		require.True(t, resp.GetSuccess())
		require.Empty(t, resp.GetErrors())
	}
}
