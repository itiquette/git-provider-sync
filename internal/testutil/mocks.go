// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

// MockLogger provides a mock logger for testing.
type MockLogger struct {
	mock.Mock
}

// Debug logs a debug message.
func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

// Info logs an info message.
func (m *MockLogger) Info(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

// Warn logs a warning message.
func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

// Error logs an error message.
func (m *MockLogger) Error(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

// Trace logs a trace message.
func (m *MockLogger) Trace(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

// IsLevelEnabled checks if a log level is enabled.
func (m *MockLogger) IsLevelEnabled(level string) bool {
	args := m.Called(level)

	return args.Bool(0)
}

// NewMockLogger creates a new mock logger with default expectations.
// By default, it accepts all log calls without assertions.
func NewMockLogger(t *testing.T) *MockLogger {
	t.Helper()

	logger := &MockLogger{}

	// Set up permissive defaults - accept any logging
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
	logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	logger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()
	logger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
	logger.On("IsLevelEnabled", mock.Anything).Return(true).Maybe()

	return logger
}

// NewStrictMockLogger creates a mock logger that requires explicit expectations.
func NewStrictMockLogger(t *testing.T) *MockLogger {
	t.Helper()

	return &MockLogger{}
}

// MockFileSystem provides a mock filesystem for testing.
type MockFileSystem struct {
	mock.Mock
}

// MkdirAll creates a directory hierarchy.
func (m *MockFileSystem) MkdirAll(path string, perm uint32) error {
	args := m.Called(path, perm)

	return args.Error(0) //nolint:wrapcheck // Mock function, testify errors are clear
}

// WriteFile writes data to a file.
func (m *MockFileSystem) WriteFile(path string, data []byte, perm uint32) error {
	args := m.Called(path, data, perm)

	return args.Error(0) //nolint:wrapcheck // Mock function, testify errors are clear
}

// ReadFile reads a file.
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock function, testify errors are clear
	}

	data, ok := args.Get(0).([]byte)
	if !ok {
		return nil, args.Error(1) //nolint:wrapcheck // Mock function, testify errors are clear
	}

	return data, args.Error(1) //nolint:wrapcheck // Mock function, testify errors are clear
}

// Exists checks if a path exists.
func (m *MockFileSystem) Exists(path string) bool {
	args := m.Called(path)

	return args.Bool(0)
}

// Remove removes a file or directory.
func (m *MockFileSystem) Remove(path string) error {
	args := m.Called(path)

	return args.Error(0) //nolint:wrapcheck // Mock function, testify errors are clear
}

// NewMockFileSystem creates a new mock filesystem with common defaults.
func NewMockFileSystem(t *testing.T) *MockFileSystem {
	t.Helper()

	fs := &MockFileSystem{} //nolint:varnamelen // fs is standard abbreviation

	// Set common defaults
	fs.On("MkdirAll", mock.Anything, mock.Anything).Return(nil).Maybe()
	fs.On("Exists", mock.Anything).Return(true).Maybe()

	return fs
}

// StandardMocks provides a collection of commonly used mocks.
//
//nolint:containedctx // Test utility needs to store context for test lifecycle
type StandardMocks struct {
	Logger     *MockLogger
	FileSystem *MockFileSystem
	Context    context.Context
}

// NewStandardMocks creates a standard set of mocks for testing.
func NewStandardMocks(t *testing.T) *StandardMocks {
	t.Helper()

	return &StandardMocks{
		Logger:     NewMockLogger(t),
		FileSystem: NewMockFileSystem(t),
		Context:    context.Background(),
	}
}

// AssertExpectations verifies all mock expectations were met.
func (sm *StandardMocks) AssertExpectations(t *testing.T) {
	t.Helper()

	sm.Logger.AssertExpectations(t)
	sm.FileSystem.AssertExpectations(t)
}

// MockBuilder provides a fluent API for setting up mocks.
type MockBuilder struct {
	t     *testing.T
	mocks *StandardMocks
}

// NewMockBuilder creates a new mock builder.
func NewMockBuilder(t *testing.T) *MockBuilder {
	t.Helper()

	return &MockBuilder{
		t:     t,
		mocks: NewStandardMocks(t),
	}
}

// WithLogger configures the mock logger.
func (mb *MockBuilder) WithLogger(setupFunc func(*MockLogger)) *MockBuilder {
	setupFunc(mb.mocks.Logger)

	return mb
}

// WithFileSystem configures the mock filesystem.
func (mb *MockBuilder) WithFileSystem(setupFunc func(*MockFileSystem)) *MockBuilder {
	setupFunc(mb.mocks.FileSystem)

	return mb
}

// Build returns the configured mocks.
func (mb *MockBuilder) Build() *StandardMocks {
	return mb.mocks
}

// ExpectLoggerCalls sets up common logger expectations.
func ExpectLoggerCalls(logger *MockLogger, level string) {
	switch level {
	case "debug":
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()

		fallthrough
	case "info":
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()

		fallthrough
	case "warn":
		logger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()

		fallthrough
	case "error":
		logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()
	case "trace":
		logger.On("Trace", mock.Anything, mock.Anything, mock.Anything).Maybe()
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe()
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
		logger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
		logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()
	}
}

// MockRepository provides a mock repository for testing.
type MockRepository struct {
	mock.Mock
}

// Clone mocks cloning a repository.
func (m *MockRepository) Clone(ctx context.Context, url, path string) error {
	args := m.Called(ctx, url, path)

	return args.Error(0) //nolint:wrapcheck // Mock function, testify errors are clear
}

// Push mocks pushing to a repository.
func (m *MockRepository) Push(ctx context.Context, remote string) error {
	args := m.Called(ctx, remote)

	return args.Error(0) //nolint:wrapcheck // Mock function, testify errors are clear
}

// Pull mocks pulling from a repository.
func (m *MockRepository) Pull(ctx context.Context, remote string) error {
	args := m.Called(ctx, remote)

	return args.Error(0) //nolint:wrapcheck // Mock function, testify errors are clear
}

// NewMockRepository creates a new mock repository.
func NewMockRepository(t *testing.T) *MockRepository {
	t.Helper()

	return &MockRepository{}
}

// MockProvider provides a mock Git provider for testing.
type MockProvider struct {
	mock.Mock
}

// ListRepositories mocks listing repositories.
func (m *MockProvider) ListRepositories(ctx context.Context, owner string) ([]string, error) {
	args := m.Called(ctx, owner)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock function, testify errors are clear
	}

	repos, ok := args.Get(0).([]string)
	if !ok {
		return nil, args.Error(1) //nolint:wrapcheck // Mock function, testify errors are clear
	}

	return repos, args.Error(1) //nolint:wrapcheck // Mock function, testify errors are clear
}

// GetRepository mocks getting a repository.
func (m *MockProvider) GetRepository(ctx context.Context, owner, name string) (interface{}, error) {
	args := m.Called(ctx, owner, name)

	return args.Get(0), args.Error(1)
}

// CreateRepository mocks creating a repository.
func (m *MockProvider) CreateRepository(ctx context.Context, owner, name string) error {
	args := m.Called(ctx, owner, name)

	return args.Error(0) //nolint:wrapcheck // Mock function, testify errors are clear
}

// NewMockProvider creates a new mock provider.
func NewMockProvider(t *testing.T) *MockProvider {
	t.Helper()

	return &MockProvider{}
}

// TestMockSuite provides a complete suite of mocks for integration testing.
//
//nolint:containedctx // Test utility needs to store context for test lifecycle
type TestMockSuite struct {
	t          *testing.T
	Logger     *MockLogger
	FileSystem *MockFileSystem
	Repository *MockRepository
	Provider   *MockProvider
	Context    context.Context
}

// NewTestMockSuite creates a complete mock suite.
func NewTestMockSuite(t *testing.T) *TestMockSuite {
	t.Helper()

	return &TestMockSuite{
		t:          t,
		Logger:     NewMockLogger(t),
		FileSystem: NewMockFileSystem(t),
		Repository: NewMockRepository(t),
		Provider:   NewMockProvider(t),
		Context:    context.Background(),
	}
}

// AssertAllExpectations verifies all mock expectations in the suite.
func (suite *TestMockSuite) AssertAllExpectations() {
	suite.t.Helper()

	suite.Logger.AssertExpectations(suite.t)
	suite.FileSystem.AssertExpectations(suite.t)
	suite.Repository.AssertExpectations(suite.t)
	suite.Provider.AssertExpectations(suite.t)
}

// SetupSuccessScenario sets up mocks for a successful operation.
func (suite *TestMockSuite) SetupSuccessScenario() {
	suite.Repository.On("Clone", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.Repository.On("Push", mock.Anything, mock.Anything).Return(nil)
	suite.Repository.On("Pull", mock.Anything, mock.Anything).Return(nil)

	suite.Provider.On("ListRepositories", mock.Anything, mock.Anything).
		Return([]string{"repo1", "repo2"}, nil)
	suite.Provider.On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(struct{}{}, nil)
	suite.Provider.On("Create", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
}

// SetupFailureScenario sets up mocks for a failure scenario.
func (suite *TestMockSuite) SetupFailureScenario(errMsg string) error {
	err := &MockError{Message: errMsg}

	suite.Repository.On("Clone", mock.Anything, mock.Anything, mock.Anything).Return(err)
	suite.Repository.On("Push", mock.Anything, mock.Anything).Return(err)
	suite.Repository.On("Pull", mock.Anything, mock.Anything).Return(err)

	suite.Provider.On("ListRepositories", mock.Anything, mock.Anything).
		Return([]string{}, err)
	suite.Provider.On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, err)
	suite.Provider.On("Create", mock.Anything, mock.Anything, mock.Anything).
		Return(err)

	return err
}

// MockError provides a simple error for testing.
type MockError struct {
	Message string
}

// Error returns the error message.
func (e *MockError) Error() string {
	return e.Message
}
