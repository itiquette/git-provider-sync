// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package mirror

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Mock implementations for testing

type mockGitOps struct{}

func (m *mockGitOps) Clone(_ context.Context, _ ports.CloneOptions) (ports.GitRepository, error) {
	return &mockGitRepository{}, nil
}

func (m *mockGitOps) Open(_ context.Context, _ string) (ports.GitRepository, error) {
	return &mockGitRepository{}, nil
}

func (m *mockGitOps) Cleanup(_ context.Context, _ string) error {
	return nil
}

func (m *mockGitOps) Init(_ context.Context, _ string, _ ports.InitOptions) (ports.GitRepository, error) {
	return &mockGitRepository{}, nil
}

func (m *mockGitOps) SupportsURL(_ string) bool {
	return true
}

func (m *mockGitOps) GetName() string {
	return "mock-git-ops"
}

func (m *mockGitOps) CreateTmpDir(ctx context.Context, _, _ string) (context.Context, error) {
	return ctx, nil
}

func (m *mockGitOps) GetTmpDirPath(_ context.Context) (string, error) {
	return "/tmp/mock", nil
}

func (m *mockGitOps) DeleteTmpDir(_ context.Context) error {
	return nil
}

type mockRepoProvider struct{}

func (m *mockRepoProvider) Create(_ context.Context, _ ports.ProviderConfig, options ports.CreateRepositoryOptions) (entities.Repository, error) {
	repo, err := entities.NewRepositoryBuilder().WithName(options.Name)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository name: %w", err)
	}

	// Add a mock URL so the repository is valid
	repo, err = repo.WithHTTPSURL("https://mock-provider.example.com/owner/" + options.Name + ".git")
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set HTTPS URL: %w", err)
	}

	built, err := repo.Build()
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to build repository: %w", err)
	}

	return built, nil
}

func (m *mockRepoProvider) Update(_ context.Context, _ ports.ProviderConfig, _ string, _ ports.UpdateRepositoryOptions) error {
	return nil
}

func (m *mockRepoProvider) Get(_ context.Context, _ ports.ProviderConfig, name string) (entities.Repository, error) {
	repo, err := entities.NewRepositoryBuilder().WithName(name)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository name: %w", err)
	}

	// Add a mock URL so the repository is valid
	repo, err = repo.WithHTTPSURL("https://mock-provider.example.com/owner/" + name + ".git")
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set HTTPS URL: %w", err)
	}

	built, err := repo.Build()
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to build repository: %w", err)
	}

	return built, nil
}

func (m *mockRepoProvider) List(_ context.Context, _ ports.ProviderConfig) ([]entities.Repository, error) {
	return []entities.Repository{}, nil
}

func (m *mockRepoProvider) Delete(_ context.Context, _ ports.ProviderConfig, _ string) error {
	return nil
}

// Missing RepositoryProvider interface methods.
func (m *mockRepoProvider) CreateForPush(_ context.Context, _ ports.CreateRepositoryRequest) (string, error) {
	return "https://example.com/repo.git", nil
}

func (m *mockRepoProvider) SetDefaultBranch(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockRepoProvider) Exists(_ context.Context, _ ports.RepositoryExistsRequest) (bool, string, error) {
	return true, "https://example.com/repo.git", nil
}

func (m *mockRepoProvider) ValidateName(_ string) error {
	return nil
}

func (m *mockRepoProvider) GetProtection(_ context.Context, _ ports.ProviderConfig, _, _ string) (ports.BranchProtection, error) {
	return ports.BranchProtection{}, nil
}

func (m *mockRepoProvider) SetProtection(_ context.Context, _ ports.ProviderConfig, _, _ string, _ ports.BranchProtection) error {
	return nil
}

func (m *mockRepoProvider) RemoveProtection(_ context.Context, _ ports.ProviderConfig, _, _ string) error {
	return nil
}

func (m *mockRepoProvider) ListProtectedBranches(_ context.Context, _ ports.ProviderConfig, _ string) ([]string, error) {
	return []string{}, nil
}

func (m *mockRepoProvider) GetInfo() ports.ProviderInfo {
	return ports.ProviderInfo{Name: "mock-provider", Type: "git"}
}

func (m *mockRepoProvider) PushToProvider(_ context.Context, _ ports.ProviderConfig, _ string, _ ports.SyncOptions) (ports.SyncResult, error) {
	return ports.SyncResult{}, nil
}

func (m *mockRepoProvider) ProjectExists(_ context.Context, _, _ string) (bool, string, error) {
	return true, "123", nil
}

func (m *mockRepoProvider) Protect(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockRepoProvider) Unprotect(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockRepoProvider) IsValidProjectName(_ context.Context, _ string) bool {
	return true
}

func (m *mockRepoProvider) SupportsFeature(_ ports.ProviderFeature) bool {
	return true
}

func (m *mockRepoProvider) TransformName(name string, _ ports.NameTransformOptions) string {
	return name
}

type mockLogger struct{}

func (m *mockLogger) Debug(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLogger) Info(_ context.Context, _ string, _ map[string]any)  {}
func (m *mockLogger) Error(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLogger) Trace(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLogger) Warn(_ context.Context, _ string, _ map[string]any)  {}
func (m *mockLogger) Fatal(_ context.Context, _ string, _ map[string]any) {}
func (m *mockLogger) IsLevelEnabled(_ ports.LogLevel) bool                { return true }
func (m *mockLogger) GetLevel() ports.LogLevel                            { return ports.LogLevelInfo }

// Test Service constructors

func TestNewService(t *testing.T) {
	t.Parallel()

	gitOps := &mockGitOps{}
	repoProvider := &mockRepoProvider{}
	logger := &mockLogger{}
	config := Config{
		TempDirectory:   "/tmp/test",
		DefaultTimeout:  30 * time.Minute,
		MaxRetries:      3,
		RetryDelay:      time.Second,
		EnableMetrics:   true,
		EnableLogging:   true,
		DryRunByDefault: false,
		ForceByDefault:  false,
	}

	service := NewService(gitOps, repoProvider, logger, config)

	require.NotNil(t, service)
	assert.Equal(t, config, service.config)
	assert.NotNil(t, service.interpreter)
}

func TestNewMirrorService(t *testing.T) {
	t.Parallel()

	gitOps := &mockGitOps{}
	repoProvider := &mockRepoProvider{}
	logger := &mockLogger{}

	service := NewMirrorService(gitOps, repoProvider, logger)

	require.NotNil(t, service)
	assert.Equal(t, "/tmp/git-provider-sync", service.config.TempDirectory)
	assert.Equal(t, 30*time.Minute, service.config.DefaultTimeout)
	assert.Equal(t, 3, service.config.MaxRetries)
	assert.Equal(t, time.Second, service.config.RetryDelay)
	assert.True(t, service.config.EnableMetrics)
	assert.True(t, service.config.EnableLogging)
	assert.False(t, service.config.DryRunByDefault)
	assert.False(t, service.config.ForceByDefault)
}

func TestNewDryRunMirrorService(t *testing.T) {
	t.Parallel()

	gitOps := &mockGitOps{}
	repoProvider := &mockRepoProvider{}
	logger := &mockLogger{}

	service := NewDryRunMirrorService(gitOps, repoProvider, logger)

	require.NotNil(t, service)
	assert.Equal(t, "/tmp/git-provider-sync-dryrun", service.config.TempDirectory)
	assert.Equal(t, 30*time.Minute, service.config.DefaultTimeout)
	assert.Equal(t, 1, service.config.MaxRetries)
	assert.Equal(t, time.Second, service.config.RetryDelay)
	assert.True(t, service.config.EnableMetrics)
	assert.True(t, service.config.EnableLogging)
	assert.True(t, service.config.DryRunByDefault)
	assert.False(t, service.config.ForceByDefault)
}

// Test Service operations

func TestService_MirrorRepository(t *testing.T) {
	t.Parallel()

	service := createTestService()
	ctx := context.Background()

	sourceBuilder := entities.NewRepositoryBuilder()
	sourceBuilder, err := sourceBuilder.WithName("test-repo")
	require.NoError(t, err)
	sourceBuilder, err = sourceBuilder.WithHTTPSURL("https://test-github.example.com/owner/test-repo.git")
	require.NoError(t, err)

	sourceBuilder = sourceBuilder.WithProviderType("github")
	sourceBuilder, err = sourceBuilder.WithDefaultBranch("main")
	require.NoError(t, err)

	sourceBuilder = sourceBuilder.WithPrivate(false)
	sourceBuilder = sourceBuilder.WithDescription("Test repository")
	sourceBuilder = sourceBuilder.WithVisibility("public")
	source, err := sourceBuilder.Build()
	require.NoError(t, err)

	targetBuilder := entities.NewRepositoryBuilder()
	targetBuilder, err = targetBuilder.WithName("test-repo-mirror")
	require.NoError(t, err)
	targetBuilder, err = targetBuilder.WithHTTPSURL("https://test-gitlab.example.com/owner/test-repo-mirror.git")
	require.NoError(t, err)

	targetBuilder = targetBuilder.WithProviderType("gitlab")
	targetBuilder, err = targetBuilder.WithDefaultBranch("main")
	require.NoError(t, err)

	targetBuilder = targetBuilder.WithPrivate(true)
	targetBuilder = targetBuilder.WithDescription("Mirror of test repository")
	targetBuilder = targetBuilder.WithVisibility("private")
	target, err := targetBuilder.Build()
	require.NoError(t, err)

	sourceAuth := AuthSpec{
		Type:  ports.AuthTypeToken,
		Token: "source-token",
	}

	targetAuth := AuthSpec{
		Type:  ports.AuthTypeToken,
		Token: "target-token",
	}

	result := service.MirrorRepository(ctx, source, sourceAuth, target, targetAuth)

	// Result should be successful (mocked operations succeed)
	assert.True(t, result.Success)
	require.NoError(t, result.Error)
	assert.Equal(t, OperationTypeCloneAndMirror, result.Operation.Type)
	assert.Equal(t, "test-repo", result.Operation.Source.Name)
	assert.Equal(t, "test-repo-mirror", result.Operation.Target.Name)
}

func TestService_SyncRepository(t *testing.T) {
	t.Parallel()

	service := createTestService()
	ctx := context.Background()

	sourceBuilder := entities.NewRepositoryBuilder()
	sourceBuilder, _ = sourceBuilder.WithName("test-repo")
	sourceBuilder, _ = sourceBuilder.WithHTTPSURL("https://test-github.example.com/owner/test-repo.git")
	sourceBuilder = sourceBuilder.WithProviderType("github")
	source, _ := sourceBuilder.Build()

	targetBuilder := entities.NewRepositoryBuilder()
	targetBuilder, _ = targetBuilder.WithName("test-repo-sync")
	targetBuilder, _ = targetBuilder.WithHTTPSURL("https://test-gitlab.example.com/owner/test-repo-sync.git")
	targetBuilder = targetBuilder.WithProviderType("gitlab")
	target, _ := targetBuilder.Build()

	sourceAuth := AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"}
	targetAuth := AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"}

	result := service.SyncRepository(ctx, source, sourceAuth, target, targetAuth)

	assert.True(t, result.Success)
	require.NoError(t, result.Error)
	assert.Equal(t, OperationTypeSync, result.Operation.Type)
}

func TestService_MirrorRepositories(t *testing.T) {
	t.Parallel()

	service := createTestService()
	ctx := context.Background()

	source1Builder := entities.NewRepositoryBuilder()
	source1Builder, _ = source1Builder.WithName("repo1")
	source1Builder, _ = source1Builder.WithHTTPSURL("https://test-github.example.com/owner/repo1.git")
	source1Builder = source1Builder.WithProviderType("github")
	source1, _ := source1Builder.Build()
	target1Builder := entities.NewRepositoryBuilder()
	target1Builder, _ = target1Builder.WithName("repo1-mirror")
	target1Builder, _ = target1Builder.WithHTTPSURL("https://test-gitlab.example.com/owner/repo1-mirror.git")
	target1Builder = target1Builder.WithProviderType("gitlab")
	target1, _ := target1Builder.Build()

	source2Builder := entities.NewRepositoryBuilder()
	source2Builder, _ = source2Builder.WithName("repo2")
	source2Builder, _ = source2Builder.WithHTTPSURL("https://test-github.example.com/owner/repo2.git")
	source2Builder = source2Builder.WithProviderType("github")
	source2, _ := source2Builder.Build()
	target2Builder := entities.NewRepositoryBuilder()
	target2Builder, _ = target2Builder.WithName("repo2-mirror")
	target2Builder, _ = target2Builder.WithHTTPSURL("https://test-gitlab.example.com/owner/repo2-mirror.git")
	target2Builder = target2Builder.WithProviderType("gitlab")
	target2, _ := target2Builder.Build()

	pairs := []SourceMirrorPair{
		{
			Source:     source1,
			SourceAuth: AuthSpec{Type: ports.AuthTypeToken, Token: "token1"},
			Target:     target1,
			TargetAuth: AuthSpec{Type: ports.AuthTypeToken, Token: "token1"},
		},
		{
			Source:     source2,
			SourceAuth: AuthSpec{Type: ports.AuthTypeToken, Token: "token2"},
			Target:     target2,
			TargetAuth: AuthSpec{Type: ports.AuthTypeToken, Token: "token2"},
		},
	}

	result := service.MirrorRepositories(ctx, pairs)

	assert.Equal(t, 2, result.TotalOperations)
	assert.Equal(t, 2, result.SuccessfulOperations)
	assert.Equal(t, 0, result.FailedOperations)
	assert.True(t, result.Success)
	assert.Len(t, result.Results, 2)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestService_ValidateRepositoryPair(t *testing.T) {
	t.Parallel()

	service := createTestService()
	ctx := context.Background()

	tests := []struct {
		name           string
		source         entities.Repository
		target         entities.Repository
		sourceAuth     AuthSpec
		targetAuth     AuthSpec
		expectValid    bool
		expectedErrors int
	}{
		{
			name: "valid repository pair",
			source: func() entities.Repository {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("valid-repo")
				builder, _ = builder.WithHTTPSURL("https://test-github.example.com/owner/valid-repo.git")
				builder = builder.WithProviderType("github")
				repo, _ := builder.Build()

				return repo
			}(),
			target: func() entities.Repository {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("valid-target")
				builder, _ = builder.WithHTTPSURL("https://test-gitlab.example.com/owner/valid-target.git")
				builder = builder.WithProviderType("gitlab")
				repo, _ := builder.Build()

				return repo
			}(),
			sourceAuth:     AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
			targetAuth:     AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
			expectValid:    true,
			expectedErrors: 0,
		},
		{
			name: "empty source URL",
			source: func() entities.Repository {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("invalid-repo")
				builder, _ = builder.WithHTTPSURL("")
				builder = builder.WithProviderType("github")
				repo, _ := builder.Build()

				return repo
			}(),
			target: func() entities.Repository {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("valid-target")
				builder, _ = builder.WithHTTPSURL("https://test-gitlab.example.com/owner/valid-target.git")
				builder = builder.WithProviderType("gitlab")
				repo, _ := builder.Build()

				return repo
			}(),
			sourceAuth:     AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
			targetAuth:     AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
			expectValid:    false,
			expectedErrors: 3, // URL empty, Name empty (build failed), Provider empty (build failed)
		},
		{
			name: "empty repository names",
			source: func() entities.Repository {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("")
				builder, _ = builder.WithHTTPSURL("https://test-github.example.com/owner/repo.git")
				builder = builder.WithProviderType("github")
				repo, _ := builder.Build()

				return repo
			}(),
			target: func() entities.Repository {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("")
				builder, _ = builder.WithHTTPSURL("https://test-gitlab.example.com/owner/repo.git")
				builder = builder.WithProviderType("gitlab")
				repo, _ := builder.Build()

				return repo
			}(),
			sourceAuth:     AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
			targetAuth:     AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
			expectValid:    false,
			expectedErrors: 4, // Both builds fail: source (URL+Name+Provider empty) + target (URL+Name+Provider empty)
		},
		{
			name: "unsupported provider",
			source: func() entities.Repository {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("repo")
				builder, _ = builder.WithHTTPSURL("https://test-github.example.com/owner/repo.git")
				builder = builder.WithProviderType("unsupported")
				repo, _ := builder.Build()

				return repo
			}(),
			target: func() entities.Repository {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("repo-target")
				builder, _ = builder.WithHTTPSURL("https://test-gitlab.example.com/owner/repo-target.git")
				builder = builder.WithProviderType("gitlab")
				repo, _ := builder.Build()

				return repo
			}(),
			sourceAuth:     AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
			targetAuth:     AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
			expectValid:    false,
			expectedErrors: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.ValidateRepositoryPair(ctx, test.source, test.sourceAuth, test.target, test.targetAuth)

			assert.Equal(t, test.expectValid, result.Valid)
			assert.Len(t, result.Results, test.expectedErrors)
			assert.Equal(t, test.source.Name(), result.Source.Name)
			assert.Equal(t, test.target.Name(), result.Target.Name)
		})
	}
}

func TestService_PlanMirrorOperation(t *testing.T) {
	t.Parallel()

	service := createTestService()

	sourceBuilder := entities.NewRepositoryBuilder()
	sourceBuilder, _ = sourceBuilder.WithName("test-repo")
	sourceBuilder, _ = sourceBuilder.WithHTTPSURL("https://test-github.example.com/owner/test-repo.git")
	sourceBuilder = sourceBuilder.WithProviderType("github")
	source, _ := sourceBuilder.Build()

	targetBuilder := entities.NewRepositoryBuilder()
	targetBuilder, _ = targetBuilder.WithName("test-repo-mirror")
	targetBuilder, _ = targetBuilder.WithHTTPSURL("https://test-gitlab.example.com/owner/test-repo-mirror.git")
	targetBuilder = targetBuilder.WithProviderType("gitlab")
	target, _ := targetBuilder.Build()

	sourceAuth := AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"}
	targetAuth := AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"}

	operation := service.PlanMirrorOperation(source, sourceAuth, target, targetAuth)

	assert.Equal(t, OperationTypeCloneAndMirror, operation.Type)
	assert.Equal(t, "test-repo", operation.Source.Name)
	assert.Equal(t, "test-repo-mirror", operation.Target.Name)
	assert.NotEmpty(t, operation.Effects)
	assert.NotEmpty(t, operation.Validations)
}

// Test helper methods

func TestService_extractOwnerFromURL(t *testing.T) {
	t.Parallel()

	service := createTestService()

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS URL with .git",
			url:      "https://github.com/owner/repo.git",
			expected: "owner",
		},
		{
			name:     "HTTPS URL without .git",
			url:      "https://github.com/owner/repo",
			expected: "owner",
		},
		{
			name:     "SSH URL with git@",
			url:      "git@github.com:owner/repo.git",
			expected: "owner",
		},
		{
			name:     "SSH URL format",
			url:      "ssh://git@github.com/owner/repo.git",
			expected: "owner",
		},
		{
			name:     "GitLab HTTPS URL with nested groups",
			url:      "https://gitlab.com/group/subgroup/project.git",
			expected: "group",
		},
		{
			name:     "GitLab SSH URL with nested groups",
			url:      "git@gitlab.com:group/subgroup/project.git",
			expected: "group",
		},
		{
			name:     "Bitbucket HTTPS URL",
			url:      "https://bitbucket.org/workspace/repo.git",
			expected: "workspace",
		},
		{
			name:     "Gitea HTTPS URL",
			url:      "https://gitea.example.com/owner/repo.git",
			expected: "owner",
		},
		{
			name:     "GitHub Enterprise URL",
			url:      "https://github.enterprise.com/org/repo.git",
			expected: "org",
		},
		{
			name:     "URL with port number",
			url:      "ssh://git@gitlab.example.com:2222/owner/repo.git",
			expected: "owner",
		},
		{
			name:     "Invalid URL",
			url:      "not-a-url",
			expected: "unknown",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "unknown",
		},
		{
			name:     "URL with only domain",
			url:      "https://github.com/",
			expected: "unknown",
		},
		{
			name:     "Malformed SSH URL",
			url:      "git@github.com",
			expected: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.extractOwnerFromURL(test.url)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestService_buildRepositorySpecWithAuth(t *testing.T) {
	t.Parallel()

	service := createTestService()

	repoBuilder := entities.NewRepositoryBuilder()
	repoBuilder, _ = repoBuilder.WithName("test-repo")
	repoBuilder, _ = repoBuilder.WithHTTPSURL("https://test-github.example.com/owner/test-repo.git")
	repoBuilder = repoBuilder.WithProviderType("github")
	repoBuilder, _ = repoBuilder.WithDefaultBranch("main")
	repoBuilder = repoBuilder.WithPrivate(true)
	repoBuilder = repoBuilder.WithDescription("Test repository")
	repoBuilder = repoBuilder.WithVisibility("private")
	repo, _ := repoBuilder.Build()

	auth := AuthSpec{
		Type:  ports.AuthTypeToken,
		Token: "test-token",
	}

	spec := service.buildRepositorySpecWithAuth(repo, auth)

	assert.Equal(t, "test-repo", spec.Name)
	assert.Equal(t, "https://test-github.example.com/owner/test-repo.git", spec.URL)
	assert.Equal(t, "github", spec.Provider)
	assert.Equal(t, "main", spec.Branch)
	assert.Equal(t, "owner", spec.Owner)
	assert.True(t, spec.IsPrivate)
	assert.Equal(t, "Test repository", spec.Description)
	assert.Equal(t, "private", spec.Visibility)
	assert.Equal(t, auth, spec.Auth)
	assert.Contains(t, spec.LocalPath, service.config.TempDirectory)
	assert.Contains(t, spec.LocalPath, "test-repo")
}

func TestService_buildOperationOptions(t *testing.T) {
	t.Parallel()

	config := Config{
		DryRunByDefault: true,
		ForceByDefault:  true,
		DefaultTimeout:  15 * time.Minute,
		MaxRetries:      5,
		RetryDelay:      2 * time.Second,
	}

	service := NewService(&mockGitOps{}, &mockRepoProvider{}, &mockLogger{}, config)

	// Test with no options
	options := service.buildOperationOptions()

	assert.True(t, options.DryRun)
	assert.True(t, options.Force)
	assert.Equal(t, 15*time.Minute, options.Timeout)
	assert.Equal(t, 5, options.RetryPolicy.MaxAttempts)
	assert.Equal(t, 2*time.Second, options.RetryPolicy.Delay)
	assert.Equal(t, BackoffStrategyExponential, options.RetryPolicy.Backoff)

	// Test with custom options
	customOptions := service.buildOperationOptions(
		WithDryRun(false),
		WithForce(false),
		WithTimeout(10*time.Minute),
	)

	assert.False(t, customOptions.DryRun)
	assert.False(t, customOptions.Force)
	assert.Equal(t, 10*time.Minute, customOptions.Timeout)
}

// Test validation functions

func TestValidateRepositoryNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		operation    Operation
		expectValid  bool
		expectedCode string
	}{
		{
			name: "valid names",
			operation: Operation{
				Source: RepositorySpec{Name: "source-repo"},
				Target: RepositorySpec{Name: "target-repo"},
			},
			expectValid: true,
		},
		{
			name: "empty source name",
			operation: Operation{
				Source: RepositorySpec{Name: ""},
				Target: RepositorySpec{Name: "target-repo"},
			},
			expectValid:  false,
			expectedCode: "EMPTY_SOURCE_NAME",
		},
		{
			name: "empty target name",
			operation: Operation{
				Source: RepositorySpec{Name: "source-repo"},
				Target: RepositorySpec{Name: ""},
			},
			expectValid:  false,
			expectedCode: "EMPTY_TARGET_NAME",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := validateRepositoryNames(test.operation)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.Equal(t, test.expectedCode, result.Code)
				assert.NotEmpty(t, result.Message)
			}
		})
	}
}

func TestValidateProviderCompatibility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		operation    Operation
		expectValid  bool
		expectedCode string
	}{
		{
			name: "supported providers",
			operation: Operation{
				Source: RepositorySpec{Provider: "github"},
				Target: RepositorySpec{Provider: "gitlab"},
			},
			expectValid: true,
		},
		{
			name: "unsupported source provider",
			operation: Operation{
				Source: RepositorySpec{Provider: "unsupported"},
				Target: RepositorySpec{Provider: "github"},
			},
			expectValid:  false,
			expectedCode: "UNSUPPORTED_SOURCE_PROVIDER",
		},
		{
			name: "unsupported target provider",
			operation: Operation{
				Source: RepositorySpec{Provider: "github"},
				Target: RepositorySpec{Provider: "unsupported"},
			},
			expectValid:  false,
			expectedCode: "UNSUPPORTED_TARGET_PROVIDER",
		},
		{
			name: "directory and archive providers",
			operation: Operation{
				Source: RepositorySpec{Provider: "directory"},
				Target: RepositorySpec{Provider: "archive"},
			},
			expectValid: true,
		},
		{
			name: "gitea provider",
			operation: Operation{
				Source: RepositorySpec{Provider: "gitea"},
				Target: RepositorySpec{Provider: "github"},
			},
			expectValid: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := validateProviderCompatibility(test.operation)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.Equal(t, test.expectedCode, result.Code)
				assert.NotEmpty(t, result.Message)
			}
		})
	}
}

// Creates test service

func createTestService() *Service {
	config := Config{
		TempDirectory:   "/tmp/test",
		DefaultTimeout:  30 * time.Minute,
		MaxRetries:      3,
		RetryDelay:      time.Second,
		EnableMetrics:   true,
		EnableLogging:   true,
		DryRunByDefault: false,
		ForceByDefault:  false,
	}

	return NewService(&mockGitOps{}, &mockRepoProvider{}, &mockLogger{}, config)
}

// SyncOperations interface methods.
func (m *mockRepoProvider) PrepareForPush(_ context.Context, _ ports.CreateRepositoryRequest) (string, error) {
	return "project-id", nil
}

func (m *mockRepoProvider) VerifyTarget(_ context.Context, _, _ string) (bool, string, error) {
	return true, "project-id", nil
}

func (m *mockRepoProvider) LockForSync(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockRepoProvider) UnlockAfterSync(_ context.Context, _, _ string) error {
	return nil
}
