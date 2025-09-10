// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package mirror

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Test pure planning functions

func TestPlanCloneAndMirror(t *testing.T) {
	t.Parallel()

	// Create isolated temp directory for test
	tmpDir := t.TempDir()

	source := RepositorySpec{
		URL:         "https://github.com/owner/source-repo.git",
		Name:        "source-repo",
		Owner:       "owner",
		Provider:    "github",
		Branch:      "main",
		LocalPath:   filepath.Join(tmpDir, "source"),
		IsPrivate:   false,
		Topics:      []string{"topic1", "topic2"},
		Description: "Source repository",
		Visibility:  "public",
		Auth:        AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
	}

	target := RepositorySpec{
		URL:         "https://gitlab.com/owner/target-repo.git",
		Name:        "target-repo",
		Owner:       "owner",
		Provider:    "gitlab",
		Branch:      "main",
		LocalPath:   filepath.Join(tmpDir, "target"),
		IsPrivate:   true,
		Topics:      []string{},
		Description: "Target repository",
		Visibility:  "private",
		Auth:        AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
	}

	options := OperationOptions{
		DryRun:               false,
		CreateIfNotExists:    true,
		UpdateDescription:    true,
		SyncVisibility:       true,
		SyncTopics:           true,
		SyncDefaultBranch:    true,
		SyncBranchProtection: false,
		PreservePullRequests: true,
		PreserveIssues:       true,
		EnableLFS:            false,
		Force:                false,
		Timeout:              30 * time.Minute,
		RetryPolicy: RetryPolicy{
			MaxAttempts: 3,
			Delay:       time.Second,
			Backoff:     BackoffStrategyExponential,
		},
	}

	operation := PlanCloneAndMirror(source, target, options)

	// Verify operation structure
	assert.Equal(t, OperationTypeCloneAndMirror, operation.Type)
	assert.Equal(t, source, operation.Source)
	assert.Equal(t, target, operation.Target)
	assert.Equal(t, options, operation.Options)
	assert.Equal(t, PriorityNormal, operation.Metadata.Priority)
	assert.NotEmpty(t, operation.Metadata.ID)
	assert.NotZero(t, operation.Metadata.CreatedAt)

	// Verify effects are planned correctly
	require.NotEmpty(t, operation.Effects)

	// Should have clone effect
	cloneEffect := findEffectByType(operation.Effects, EffectTypeCloneRepository)
	require.NotNil(t, cloneEffect)
	assert.Equal(t, "Clone source repository", cloneEffect.Description)
	assert.Equal(t, source.URL, cloneEffect.Parameters["url"])
	assert.Equal(t, source.LocalPath, cloneEffect.Parameters["local_path"])
	assert.Equal(t, source.Auth, cloneEffect.Parameters["auth"])
	assert.Equal(t, source.Branch, cloneEffect.Parameters["branch"])

	// Should have create repository effect (CreateIfNotExists=true)
	createEffect := findEffectByType(operation.Effects, EffectTypeCreateRepository)
	require.NotNil(t, createEffect)
	assert.Equal(t, "Create target repository if not exists", createEffect.Description)
	assert.Equal(t, target.Name, createEffect.Parameters["name"])
	assert.Equal(t, target.Owner, createEffect.Parameters["owner"])
	assert.Equal(t, target.Description, createEffect.Parameters["description"])
	assert.Equal(t, target.IsPrivate, createEffect.Parameters["private"])

	// Should have push effect
	pushEffect := findEffectByType(operation.Effects, EffectTypePushToRepository)
	require.NotNil(t, pushEffect)
	assert.Equal(t, "Push mirrored content to target", pushEffect.Description)
	assert.Equal(t, target.URL, pushEffect.Parameters["url"])
	assert.Equal(t, source.LocalPath, pushEffect.Parameters["local_path"])
	assert.Equal(t, target.Auth, pushEffect.Parameters["auth"])
	assert.Equal(t, options.Force, pushEffect.Parameters["force"])
	assert.Contains(t, pushEffect.DependsOn, "clone_repository")

	// Should have metadata update effects (based on options)
	updateDescEffect := findEffectByType(operation.Effects, EffectTypeUpdateDescription)
	require.NotNil(t, updateDescEffect)

	updateVisibilityEffect := findEffectByType(operation.Effects, EffectTypeUpdateVisibility)
	require.NotNil(t, updateVisibilityEffect)

	updateTopicsEffect := findEffectByType(operation.Effects, EffectTypeUpdateTopics)
	require.NotNil(t, updateTopicsEffect)

	// Should have cleanup effect (DryRun=false)
	cleanupEffect := findEffectByType(operation.Effects, EffectTypeCleanupTempFiles)
	require.NotNil(t, cleanupEffect)
	assert.Equal(t, source.LocalPath, cleanupEffect.Parameters["local_path"])

	// Verify validations
	require.Len(t, operation.Validations, 3)

	validationNames := make([]string, len(operation.Validations))
	for i, validation := range operation.Validations {
		validationNames[i] = validation.Name
	}

	assert.Contains(t, validationNames, "ValidSourceURL")
	assert.Contains(t, validationNames, "ValidTargetURL")
	assert.Contains(t, validationNames, "ValidAuth")
}

func TestPlanCloneAndMirror_WithDryRun(t *testing.T) {
	t.Parallel()

	// Create isolated temp directory for test
	tmpDir := t.TempDir()

	source := RepositorySpec{
		URL:       "https://github.com/owner/source.git",
		Name:      "source",
		LocalPath: tmpDir + "/source",
	}

	target := RepositorySpec{
		URL:  "https://gitlab.com/owner/target.git",
		Name: "target",
	}

	options := OperationOptions{
		DryRun:            true,
		CreateIfNotExists: true,
		UpdateDescription: false,
		SyncVisibility:    false,
		SyncTopics:        false,
	}

	operation := PlanCloneAndMirror(source, target, options)

	// Should not have cleanup effect in dry run
	cleanupEffect := findEffectByType(operation.Effects, EffectTypeCleanupTempFiles)
	assert.Nil(t, cleanupEffect)

	// Should not have metadata update effects when sync options are false
	updateDescEffect := findEffectByType(operation.Effects, EffectTypeUpdateDescription)
	assert.Nil(t, updateDescEffect)

	updateVisibilityEffect := findEffectByType(operation.Effects, EffectTypeUpdateVisibility)
	assert.Nil(t, updateVisibilityEffect)

	updateTopicsEffect := findEffectByType(operation.Effects, EffectTypeUpdateTopics)
	assert.Nil(t, updateTopicsEffect)
}

func TestPlanSync(t *testing.T) {
	t.Parallel()

	// Create isolated temp directory for test
	tmpDir := t.TempDir()

	source := RepositorySpec{
		URL:       "https://github.com/owner/source-repo.git",
		Name:      "source-repo",
		LocalPath: tmpDir + "/source",
		Auth:      AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
	}

	target := RepositorySpec{
		URL:  "https://gitlab.com/owner/target-repo.git",
		Name: "target-repo",
		Auth: AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
	}

	options := OperationOptions{
		DryRun: false,
		Force:  true,
	}

	operation := PlanSync(source, target, options)

	// Verify operation structure
	assert.Equal(t, OperationTypeSync, operation.Type)
	assert.Equal(t, source, operation.Source)
	assert.Equal(t, target, operation.Target)
	assert.Equal(t, options, operation.Options)

	// Sync should have fewer effects than full clone and mirror
	require.Len(t, operation.Effects, 3) // clone, push, cleanup

	// Should have clone effect
	cloneEffect := findEffectByType(operation.Effects, EffectTypeCloneRepository)
	require.NotNil(t, cloneEffect)

	// Should have push effect
	pushEffect := findEffectByType(operation.Effects, EffectTypePushToRepository)
	require.NotNil(t, pushEffect)
	assert.Contains(t, pushEffect.DependsOn, "clone_repository")

	// Should have cleanup effect
	cleanupEffect := findEffectByType(operation.Effects, EffectTypeCleanupTempFiles)
	require.NotNil(t, cleanupEffect)

	// Should NOT have create repository effect (sync assumes repos exist)
	createEffect := findEffectByType(operation.Effects, EffectTypeCreateRepository)
	assert.Nil(t, createEffect)

	// Verify validations
	require.Len(t, operation.Validations, 2)

	validationNames := make([]string, len(operation.Validations))
	for i, validation := range operation.Validations {
		validationNames[i] = validation.Name
	}

	assert.Contains(t, validationNames, "ValidSourceURL")
	assert.Contains(t, validationNames, "ValidTargetURL")
}

func TestPlanSync_WithDryRun(t *testing.T) {
	t.Parallel()

	// Create isolated temp directory for test
	tmpDir := t.TempDir()

	source := RepositorySpec{
		URL:       "https://github.com/owner/source.git",
		Name:      "source",
		LocalPath: tmpDir + "/source",
	}

	target := RepositorySpec{
		URL:  "https://gitlab.com/owner/target.git",
		Name: "target",
	}

	options := OperationOptions{DryRun: true}

	operation := PlanSync(source, target, options)

	// Should not have cleanup effect in dry run
	cleanupEffect := findEffectByType(operation.Effects, EffectTypeCleanupTempFiles)
	assert.Nil(t, cleanupEffect)

	// Should have only clone and push effects
	require.Len(t, operation.Effects, 2)
}

// Test validation functions

func TestValidateOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		operation          Operation
		expectedFailures   int
		expectedErrorCodes []string
	}{
		{
			name: "valid operation",
			operation: Operation{
				Source: RepositorySpec{
					URL:  "https://github.com/owner/source.git",
					Name: "source",
					Auth: AuthSpec{Type: ports.AuthTypeToken, Token: "token"},
				},
				Target: RepositorySpec{
					URL:  "https://gitlab.com/owner/target.git",
					Name: "target",
					Auth: AuthSpec{Type: ports.AuthTypeToken, Token: "token"},
				},
				Validations: []ValidationRule{
					{Name: "ValidSourceURL", Predicate: validateSourceURL},
					{Name: "ValidTargetURL", Predicate: validateTargetURL},
					{Name: "ValidAuth", Predicate: validateAuth},
				},
			},
			expectedFailures: 0,
		},
		{
			name: "invalid operation - empty URLs",
			operation: Operation{
				Source: RepositorySpec{
					URL:  "",
					Name: "source",
					Auth: AuthSpec{Type: ports.AuthTypeToken, Token: "token"},
				},
				Target: RepositorySpec{
					URL:  "",
					Name: "target",
					Auth: AuthSpec{Type: ports.AuthTypeToken, Token: "token"},
				},
				Validations: []ValidationRule{
					{Name: "ValidSourceURL", Predicate: validateSourceURL},
					{Name: "ValidTargetURL", Predicate: validateTargetURL},
				},
			},
			expectedFailures:   2,
			expectedErrorCodes: []string{"EMPTY_SOURCE_URL", "EMPTY_TARGET_URL"},
		},
		{
			name: "invalid operation - missing auth",
			operation: Operation{
				Source: RepositorySpec{
					URL:  "https://github.com/owner/source.git",
					Name: "source",
					Auth: AuthSpec{Type: ports.AuthTypeToken}, // No token provided
				},
				Target: RepositorySpec{
					URL:  "https://gitlab.com/owner/target.git",
					Name: "target",
					Auth: AuthSpec{Type: ports.AuthTypeSSH}, // No SSH key provided
				},
				Validations: []ValidationRule{
					{Name: "ValidAuth", Predicate: validateAuth},
				},
			},
			expectedFailures:   1,
			expectedErrorCodes: []string{"MISSING_SOURCE_AUTH"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			results := ValidateOperation(test.operation)

			assert.Len(t, results, test.expectedFailures)

			if test.expectedFailures > 0 {
				actualCodes := make([]string, len(results))
				for i, result := range results {
					actualCodes[i] = result.Code
					assert.False(t, result.Valid)
					assert.NotEmpty(t, result.Message)
				}

				for _, expectedCode := range test.expectedErrorCodes {
					assert.Contains(t, actualCodes, expectedCode)
				}
			}
		})
	}
}

// Test individual validation functions

func TestValidateSourceURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		operation   Operation
		expectValid bool
		expectCode  string
	}{
		{
			name: "valid source URL",
			operation: Operation{
				Source: RepositorySpec{URL: "https://github.com/owner/repo.git"},
			},
			expectValid: true,
		},
		{
			name: "empty source URL",
			operation: Operation{
				Source: RepositorySpec{URL: ""},
			},
			expectValid: false,
			expectCode:  "EMPTY_SOURCE_URL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := validateSourceURL(test.operation)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.Equal(t, test.expectCode, result.Code)
				assert.NotEmpty(t, result.Message)
			}
		})
	}
}

func TestValidateTargetURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		operation   Operation
		expectValid bool
		expectCode  string
	}{
		{
			name: "valid target URL",
			operation: Operation{
				Target: RepositorySpec{URL: "https://gitlab.com/owner/repo.git"},
			},
			expectValid: true,
		},
		{
			name: "empty target URL",
			operation: Operation{
				Target: RepositorySpec{URL: ""},
			},
			expectValid: false,
			expectCode:  "EMPTY_TARGET_URL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := validateTargetURL(test.operation)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.Equal(t, test.expectCode, result.Code)
				assert.NotEmpty(t, result.Message)
			}
		})
	}
}

func TestValidateAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		operation   Operation
		expectValid bool
		expectCode  string
	}{
		{
			name: "valid auth - both tokens provided",
			operation: Operation{
				Source: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeToken, Token: "source-token"},
				},
				Target: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
				},
			},
			expectValid: true,
		},
		{
			name: "valid auth - no auth required",
			operation: Operation{
				Source: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeNone},
				},
				Target: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeNone},
				},
			},
			expectValid: true,
		},
		{
			name: "valid auth - SSH keys provided",
			operation: Operation{
				Source: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeSSH, SSHKeyPath: "/path/to/key"},
				},
				Target: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeSSH, SSHKey: "ssh-key-content"},
				},
			},
			expectValid: true,
		},
		{
			name: "invalid auth - missing source token",
			operation: Operation{
				Source: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeToken}, // No token
				},
				Target: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeToken, Token: "target-token"},
				},
			},
			expectValid: false,
			expectCode:  "MISSING_SOURCE_AUTH",
		},
		{
			name: "invalid auth - missing target SSH key",
			operation: Operation{
				Source: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeSSH, SSHKeyPath: "/path/to/key"},
				},
				Target: RepositorySpec{
					Auth: AuthSpec{Type: ports.AuthTypeSSH}, // No SSH key or path
				},
			},
			expectValid: false,
			expectCode:  "MISSING_TARGET_AUTH",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := validateAuth(test.operation)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.Equal(t, test.expectCode, result.Code)
				assert.NotEmpty(t, result.Message)
			}
		})
	}
}

// Test helper functions

func TestGenerateOperationID(t *testing.T) {
	t.Parallel()

	source := RepositorySpec{
		Owner: "source-owner",
		Name:  "source-repo",
	}

	target := RepositorySpec{
		Owner: "target-owner",
		Name:  "target-repo",
	}

	id := generateOperationID(source, target)

	expected := "source-owner/source-repo->target-owner/target-repo"
	assert.Equal(t, expected, id)
}

func TestBuildRepositorySpec(t *testing.T) {
	t.Parallel()

	auth := AuthSpec{
		Type:  ports.AuthTypeToken,
		Token: "test-token",
	}

	spec := BuildRepositorySpec("github", "owner", "repo", "https://github.com/owner/repo.git", "main", auth)

	assert.Equal(t, "https://github.com/owner/repo.git", spec.URL)
	assert.Equal(t, "repo", spec.Name)
	assert.Equal(t, "owner", spec.Owner)
	assert.Equal(t, "github", spec.Provider)
	assert.Equal(t, "main", spec.Branch)
	assert.Equal(t, auth, spec.Auth)
}

func TestBuildOperationOptions(t *testing.T) {
	t.Parallel()

	options := BuildOperationOptions()

	assert.False(t, options.DryRun)
	assert.True(t, options.CreateIfNotExists)
	assert.True(t, options.UpdateDescription)
	assert.True(t, options.SyncVisibility)
	assert.True(t, options.SyncTopics)
	assert.True(t, options.SyncDefaultBranch)
	assert.False(t, options.SyncBranchProtection) // Security default
	assert.True(t, options.PreservePullRequests)
	assert.True(t, options.PreserveIssues)
	assert.False(t, options.EnableLFS)
	assert.False(t, options.Force)
	assert.Equal(t, 30*time.Minute, options.Timeout)
	assert.Equal(t, 3, options.RetryPolicy.MaxAttempts)
	assert.Equal(t, time.Second, options.RetryPolicy.Delay)
	assert.Equal(t, BackoffStrategyExponential, options.RetryPolicy.Backoff)
}

// Test option functions

func TestWithForce_ModifyOptions_SetsBooleanCorrectly(t *testing.T) {
	t.Parallel()

	base := OperationOptions{Force: false}

	// Test enabling force
	option := WithForce(true)
	result := option(base)
	assert.True(t, result.Force)

	// Test disabling force
	option = WithForce(false)
	result = option(base)
	assert.False(t, result.Force)
}

func TestWithTimeout_ModifyOptions_SetsDurationCorrectly(t *testing.T) {
	t.Parallel()

	base := OperationOptions{Timeout: 30 * time.Minute}

	option := WithTimeout(15 * time.Minute)
	result := option(base)
	assert.Equal(t, 15*time.Minute, result.Timeout)
}

func TestApplyOperationOptions(t *testing.T) {
	t.Parallel()

	base := OperationOptions{
		DryRun:  false,
		Force:   false,
		Timeout: 30 * time.Minute,
	}

	result := ApplyOperationOptions(base,
		WithDryRun(true),
		WithForce(true),
		WithTimeout(15*time.Minute),
	)

	assert.True(t, result.DryRun)
	assert.True(t, result.Force)
	assert.Equal(t, 15*time.Minute, result.Timeout)

	// Test with no options
	result = ApplyOperationOptions(base)
	assert.Equal(t, base, result)
}

// Finds effect by type.
func findEffectByType(effects []Effect, effectType EffectType) *Effect {
	for _, effect := range effects {
		if effect.Type == effectType {
			return &effect
		}
	}

	return nil
}
