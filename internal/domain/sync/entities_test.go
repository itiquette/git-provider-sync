// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// TestCreateTestRepositoryEntity tests pure business logic of repository entity creation.
func TestCreateTestRepositoryEntity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		repoName     string
		expectedName string
		expectedURL  string
	}{
		{
			name:         "standard repository name",
			repoName:     "test-repo",
			expectedName: "test-repo",
			expectedURL:  "https://github.com/test/test-repo.git",
		},
		{
			name:         "repository with numbers",
			repoName:     "project-v2",
			expectedName: "project-v2",
			expectedURL:  "https://github.com/test/project-v2.git",
		},
		{
			name:         "single character repository",
			repoName:     "x",
			expectedName: "x",
			expectedURL:  "https://github.com/test/x.git",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Test the business logic of creating repository entities
			repo := createTestRepositoryEntity(testCase.repoName)

			// Verify entity was created correctly
			assert.Equal(t, testCase.expectedName, repo.Name())
			assert.Equal(t, testCase.expectedURL, repo.HTTPSURL())
			assert.Equal(t, "git@github.com:test/"+testCase.repoName+".git", repo.SSHURL())
			assert.Equal(t, "main", repo.DefaultBranch())
			assert.Equal(t, "Test repository for "+testCase.repoName, repo.Description())
			assert.Equal(t, "public", repo.Visibility())
			assert.False(t, repo.LastActivityAt().IsZero())
		})
	}
}

// TestCreateTestRepositoryEntityBuilder tests the builder pattern logic.
func TestCreateTestRepositoryEntityBuilder(t *testing.T) {
	t.Parallel()

	// Test that the builder pattern works correctly for repository creation
	repoName := "builder-test"
	repo := createTestRepositoryEntity(repoName)

	// Verify all builder steps were executed correctly
	require.NotEmpty(t, repo.Name())
	require.NotEmpty(t, repo.HTTPSURL())
	require.NotEmpty(t, repo.SSHURL())
	require.NotEmpty(t, repo.DefaultBranch())
	require.NotEmpty(t, repo.Description())
	require.NotEmpty(t, repo.Visibility())

	// Verify the LastActivityAt timestamp is recent (within last minute)
	timeDiff := time.Since(repo.LastActivityAt())
	assert.Less(t, timeDiff, time.Minute, "LastActivityAt should be recent")
}

// createTestRepositoryEntity creates a test Repository entity (moved from integration tests).
// This is pure business logic that doesn't need file system or git operations.
func createTestRepositoryEntity(name string) entities.Repository {
	builder := entities.NewRepositoryBuilder()
	builder, _ = builder.WithName(name)
	builder, _ = builder.WithHTTPSURL("https://github.com/test/" + name + ".git")
	builder, _ = builder.WithSSHURL("git@github.com:test/" + name + ".git")
	builder, _ = builder.WithDefaultBranch("main")
	builder = builder.WithDescription("Test repository for " + name)
	builder = builder.WithVisibility("public")
	builder = builder.WithLastActivityAt(time.Now())
	repo, _ := builder.Build()

	return repo
}
