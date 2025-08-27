// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestProjectInfoSetters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "SetASCIIName sets ASCII name flag",
			testFunc: func(t *testing.T) {
				t.Helper()
				projectInfo := &ProjectInfo{}

				projectInfo.SetASCIIName(true)
				assert.True(t, projectInfo.ASCIIName)

				projectInfo.SetASCIIName(false)
				assert.False(t, projectInfo.ASCIIName)
			},
		},
		{
			name: "SetCleanName sets clean name",
			testFunc: func(t *testing.T) {
				t.Helper()
				projectInfo := &ProjectInfo{}

				projectInfo.SetCleanName("clean-name")
				assert.Equal(t, "clean-name", projectInfo.CleanName)

				projectInfo.SetCleanName("")
				assert.Empty(t, projectInfo.CleanName)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestProjectInfoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		projectInfo  ProjectInfo
		expectedName string
	}{
		{
			name: "returns original name when ASCIIName is false",
			projectInfo: ProjectInfo{
				OriginalName: "my-original-repo",
				CleanName:    "my-clean-repo",
				ASCIIName:    false,
			},
			expectedName: "my-original-repo",
		},
		{
			name: "returns clean name when ASCIIName is true",
			projectInfo: ProjectInfo{
				OriginalName: "my-original-repo",
				CleanName:    "my-clean-repo",
				ASCIIName:    true,
			},
			expectedName: "my-clean-repo",
		},
		{
			name: "handles empty names",
			projectInfo: ProjectInfo{
				OriginalName: "",
				CleanName:    "",
				ASCIIName:    true,
			},
			expectedName: "",
		},
		{
			name: "handles special characters",
			projectInfo: ProjectInfo{
				OriginalName: "repo@with#special$chars",
				CleanName:    "repo-with-special-chars",
				ASCIIName:    true,
			},
			expectedName: "repo-with-special-chars",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := testCase.projectInfo.Name(ctx)
			assert.Equal(t, testCase.expectedName, result)
		})
	}
}

func TestProjectInfoTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		projectInfo        ProjectInfo
		expectedIsZeroTime bool
	}{
		{
			name: "returns zero time when LastActivityAt is nil",
			projectInfo: ProjectInfo{
				LastActivityAt: nil,
			},
			expectedIsZeroTime: true,
		},
		{
			name: "returns actual time when LastActivityAt is set",
			projectInfo: ProjectInfo{
				LastActivityAt: func() *time.Time {
					t := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)

					return &t
				}(),
			},
			expectedIsZeroTime: false,
		},
		{
			name: "returns zero time for empty time pointer",
			projectInfo: ProjectInfo{
				LastActivityAt: func() *time.Time {
					var t time.Time

					return &t
				}(),
			},
			expectedIsZeroTime: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.projectInfo.Time()

			if testCase.expectedIsZeroTime {
				assert.True(t, result.IsZero())
			} else {
				assert.False(t, result.IsZero())

				if testCase.projectInfo.LastActivityAt != nil {
					assert.Equal(t, *testCase.projectInfo.LastActivityAt, result)
				}
			}
		})
	}
}

func TestProjectInfoDebugLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectInfo ProjectInfo
		testFunc    func(t *testing.T, event *zerolog.Event)
	}{
		{
			name: "debug log with all fields populated",
			projectInfo: ProjectInfo{
				OriginalName:  "test-repo",
				CleanName:     "test-repo-clean",
				HTTPSURL:      "https://github.com/user/test-repo.git",
				SSHURL:        "git@github.com:user/test-repo.git",
				DefaultBranch: "main",
				Description:   "A test repository for unit testing",
				Visibility:    "public",
				LastActivityAt: func() *time.Time {
					t := time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC)

					return &t
				}(),
				ProjectID: "12345",
				ASCIIName: false,
			},
			testFunc: func(t *testing.T, event *zerolog.Event) {
				t.Helper()
				assert.NotNil(t, event)
				// Note: We can't easily verify the actual log content without
				// additional test infrastructure, but we can verify the event is created
			},
		},
		{
			name: "debug log with minimal fields",
			projectInfo: ProjectInfo{
				OriginalName:   "minimal-repo",
				HTTPSURL:       "https://example.com/repo.git",
				DefaultBranch:  "master",
				LastActivityAt: nil,
			},
			testFunc: func(t *testing.T, event *zerolog.Event) {
				t.Helper()
				assert.NotNil(t, event)
			},
		},
		{
			name: "debug log with description containing linebreaks",
			projectInfo: ProjectInfo{
				OriginalName:  "repo-with-desc",
				Description:   "Line 1\nLine 2\rLine 3\r\nLine 4",
				HTTPSURL:      "https://example.com/repo.git",
				DefaultBranch: "develop",
			},
			testFunc: func(t *testing.T, event *zerolog.Event) {
				t.Helper()
				assert.NotNil(t, event)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			logger := zerolog.New(nil)
			event := testCase.projectInfo.DebugLog(&logger)

			testCase.testFunc(t, event)
		})
	}
}

func TestProjectInfoComplexScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "project info state management",
			testFunc: func(t *testing.T) {
				t.Helper()
				projectInfo := &ProjectInfo{
					OriginalName: "original-name",
					CleanName:    "clean-name",
					ASCIIName:    false,
				}

				// Initially returns original name
				ctx := context.Background()
				assert.Equal(t, "original-name", projectInfo.Name(ctx))

				// After setting ASCII flag, returns clean name
				projectInfo.SetASCIIName(true)
				assert.Equal(t, "clean-name", projectInfo.Name(ctx))

				// Update clean name and verify
				projectInfo.SetCleanName("updated-clean-name")
				assert.Equal(t, "updated-clean-name", projectInfo.Name(ctx))

				// Turn off ASCII flag, should return original again
				projectInfo.SetASCIIName(false)
				assert.Equal(t, "original-name", projectInfo.Name(ctx))
			},
		},
		{
			name: "time handling edge cases",
			testFunc: func(t *testing.T) {
				t.Helper()
				projectInfo := ProjectInfo{}

				// Initially nil, should return zero time
				assert.True(t, projectInfo.Time().IsZero())

				// Set to a specific time
				specificTime := time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC)
				projectInfo.LastActivityAt = &specificTime

				retrievedTime := projectInfo.Time()
				assert.Equal(t, specificTime, retrievedTime)
				assert.False(t, retrievedTime.IsZero())

				// Set back to nil
				projectInfo.LastActivityAt = nil
				assert.True(t, projectInfo.Time().IsZero())
			},
		},
		{
			name: "comprehensive project info with all fields",
			testFunc: func(t *testing.T) {
				t.Helper()
				activityTime := time.Date(2023, 6, 15, 9, 45, 30, 0, time.UTC)
				projectInfo := ProjectInfo{
					OriginalName:   "My_Special-Repo@2023",
					CleanName:      "my-special-repo-2023",
					HTTPSURL:       "https://git.example.com/org/My_Special-Repo@2023.git",
					SSHURL:         "git@git.example.com:org/My_Special-Repo@2023.git",
					DefaultBranch:  "main",
					Description:    "A comprehensive test repository\nwith multiple lines\nof description",
					Visibility:     "private",
					LastActivityAt: &activityTime,
					ProjectID:      "98765",
					ASCIIName:      false,
				}

				// Test name method with different ASCII settings
				ctx := context.Background()
				assert.Equal(t, "My_Special-Repo@2023", projectInfo.Name(ctx))

				projectInfo.SetASCIIName(true)
				assert.Equal(t, "my-special-repo-2023", projectInfo.Name(ctx))

				// Test time method
				assert.Equal(t, activityTime, projectInfo.Time())
				assert.False(t, projectInfo.Time().IsZero())

				// Test debug log creation
				logger := zerolog.New(nil)
				event := projectInfo.DebugLog(&logger)
				assert.NotNil(t, event)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestProjectInfo_EmptyFieldsAndInvalidData_HandlesGracefully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "unicode characters in names",
			testFunc: func(t *testing.T) {
				t.Helper()
				projectInfo := ProjectInfo{
					OriginalName: "项目-测试-🚀",
					CleanName:    "project-test",
					ASCIIName:    false,
				}

				ctx := context.Background()
				assert.Equal(t, "项目-测试-🚀", projectInfo.Name(ctx))

				projectInfo.SetASCIIName(true)
				assert.Equal(t, "project-test", projectInfo.Name(ctx))
			},
		},
		{
			name: "very long URLs and descriptions",
			testFunc: func(t *testing.T) {
				t.Helper()
				longURL := "https://very-long-domain-name-that-exceeds-normal-length.example.com/organization-with-very-long-name/repository-with-extremely-long-name-that-might-cause-issues.git"
				longDesc := "This is a very long description that spans multiple lines and contains various special characters !@#$%^&*()_+-=[]{}|;':\",./<>? and unicode characters like 项目 and emojis 🚀🎉💻"

				projectInfo := ProjectInfo{
					OriginalName:  "test-repo",
					HTTPSURL:      longURL,
					Description:   longDesc,
					DefaultBranch: "main",
				}

				// Should handle long content without issues
				ctx := context.Background()
				assert.Equal(t, "test-repo", projectInfo.Name(ctx))

				logger := zerolog.New(nil)
				event := projectInfo.DebugLog(&logger)
				assert.NotNil(t, event)
			},
		},
		{
			name: "empty and whitespace-only fields",
			testFunc: func(t *testing.T) {
				t.Helper()
				projectInfo := ProjectInfo{
					OriginalName:  "   ",
					CleanName:     "\t\n",
					HTTPSURL:      "",
					Description:   "   \n\t   ",
					DefaultBranch: "",
				}

				ctx := context.Background()
				assert.Equal(t, "   ", projectInfo.Name(ctx))

				projectInfo.SetASCIIName(true)
				assert.Equal(t, "\t\n", projectInfo.Name(ctx))

				logger := zerolog.New(nil)
				event := projectInfo.DebugLog(&logger)
				assert.NotNil(t, event)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}
