// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSyncRunMetainfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		ctxID          int
		source         string
		target         string
		total          int
		expectedFields func(t *testing.T, meta *SyncRunMetainfo)
	}{
		{
			name:   "create sync run metainfo with all fields",
			ctxID:  12345,
			source: "github.com/source/repo",
			target: "gitlab.com/target/repo",
			total:  100,
			expectedFields: func(t *testing.T, meta *SyncRunMetainfo) {
				t.Helper()
				assert.Equal(t, 12345, meta.CtxID)
				assert.Equal(t, "github.com/source/repo", meta.Source)
				assert.Equal(t, "gitlab.com/target/repo", meta.Target)
				assert.Equal(t, 100, meta.Total)
				assert.NotNil(t, meta.Fail)
				assert.Empty(t, *meta.Fail) // Should be empty initially
			},
		},
		{
			name:   "create sync run metainfo with zero values",
			ctxID:  0,
			source: "",
			target: "",
			total:  0,
			expectedFields: func(t *testing.T, meta *SyncRunMetainfo) {
				t.Helper()
				assert.Equal(t, 0, meta.CtxID)
				assert.Empty(t, meta.Source)
				assert.Empty(t, meta.Target)
				assert.Equal(t, 0, meta.Total)
				assert.NotNil(t, meta.Fail)
				assert.Empty(t, *meta.Fail)
			},
		},
		{
			name:   "create sync run metainfo with negative values",
			ctxID:  -1,
			source: "negative test",
			target: "negative target",
			total:  -5,
			expectedFields: func(t *testing.T, meta *SyncRunMetainfo) {
				t.Helper()
				assert.Equal(t, -1, meta.CtxID)
				assert.Equal(t, "negative test", meta.Source)
				assert.Equal(t, "negative target", meta.Target)
				assert.Equal(t, -5, meta.Total)
				assert.NotNil(t, meta.Fail)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			meta := NewSyncRunMetainfo(testCase.ctxID, testCase.source, testCase.target, testCase.total)

			require.NotNil(t, meta)
			testCase.expectedFields(t, meta)
		})
	}
}

func TestSyncRunMetainfoAddFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "add single failure",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "source", "target", 10)

				meta.AddFailure("validation", "invalid repository name")

				failures := (*meta.Fail)["validation"]
				assert.Len(t, failures, 1)
				assert.Equal(t, "invalid repository name", failures[0])
			},
		},
		{
			name: "add multiple failures of same type",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "source", "target", 10)

				meta.AddFailure("network", "timeout error")
				meta.AddFailure("network", "connection refused")
				meta.AddFailure("network", "dns resolution failed")

				failures := (*meta.Fail)["network"]
				assert.Len(t, failures, 3)
				assert.Equal(t, "timeout error", failures[0])
				assert.Equal(t, "connection refused", failures[1])
				assert.Equal(t, "dns resolution failed", failures[2])
			},
		},
		{
			name: "add failures of different types",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "source", "target", 10)

				meta.AddFailure("validation", "invalid name")
				meta.AddFailure("network", "timeout")
				meta.AddFailure("auth", "unauthorized")

				assert.Len(t, *meta.Fail, 3)
				assert.Equal(t, "invalid name", (*meta.Fail)["validation"][0])
				assert.Equal(t, "timeout", (*meta.Fail)["network"][0])
				assert.Equal(t, "unauthorized", (*meta.Fail)["auth"][0])
			},
		},
		{
			name: "add failure with nil fail map",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := &SyncRunMetainfo{
					CtxID:  1,
					Source: "test",
					Target: "test",
					Total:  1,
					Fail:   nil, // Explicitly set to nil
				}

				meta.AddFailure("test", "failure message")

				assert.NotNil(t, meta.Fail)
				assert.Len(t, *meta.Fail, 1)
				assert.Equal(t, "failure message", (*meta.Fail)["test"][0])
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

func TestSyncRunMetainfoString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func() *SyncRunMetainfo
		expectedFields []string
	}{
		{
			name: "string representation without failures",
			setupFunc: func() *SyncRunMetainfo {
				return NewSyncRunMetainfo(123, "github.com/user/repo", "gitlab.com/user/repo", 50)
			},
			expectedFields: []string{
				"CtxID: 123",
				"Source: github.com/user/repo",
				"Target: gitlab.com/user/repo",
				"Total: 50",
				"No failures",
			},
		},
		{
			name: "string representation with single failure type",
			setupFunc: func() *SyncRunMetainfo {
				meta := NewSyncRunMetainfo(456, "source", "target", 25)
				meta.AddFailure("validation", "invalid repo name")
				meta.AddFailure("validation", "repo already exists")

				return meta
			},
			expectedFields: []string{
				"CtxID: 456",
				"Source: source",
				"Target: target",
				"Total: 25",
				"Failures:",
				"validation: invalid repo name, repo already exists",
			},
		},
		{
			name: "string representation with multiple failure types",
			setupFunc: func() *SyncRunMetainfo {
				meta := NewSyncRunMetainfo(789, "src", "tgt", 75)
				meta.AddFailure("network", "timeout")
				meta.AddFailure("auth", "unauthorized")
				meta.AddFailure("validation", "invalid name")

				return meta
			},
			expectedFields: []string{
				"CtxID: 789",
				"Source: src",
				"Target: tgt",
				"Total: 75",
				"Failures:",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			meta := testCase.setupFunc()
			result := meta.String()

			assert.Contains(t, result, "SyncRunMetainfo{")

			for _, expectedField := range testCase.expectedFields {
				assert.Contains(t, result, expectedField)
			}
		})
	}
}

func TestSyncRunMetainfo_FailureTracking(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "GetFailuresByType returns correct failures",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "src", "tgt", 10)
				meta.AddFailure("network", "error1")
				meta.AddFailure("network", "error2")
				meta.AddFailure("auth", "error3")

				networkFailures := meta.GetFailuresByType("network")
				assert.Len(t, networkFailures, 2)
				assert.Equal(t, "error1", networkFailures[0])
				assert.Equal(t, "error2", networkFailures[1])

				authFailures := meta.GetFailuresByType("auth")
				assert.Len(t, authFailures, 1)
				assert.Equal(t, "error3", authFailures[0])

				nonExistentFailures := meta.GetFailuresByType("nonexistent")
				assert.Nil(t, nonExistentFailures)
			},
		},
		{
			name: "HasFailures returns correct status",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "src", "tgt", 10)

				// Initially no failures
				assert.False(t, meta.HasFailures())

				// Add a failure
				meta.AddFailure("test", "error")
				assert.True(t, meta.HasFailures())
			},
		},
		{
			name: "GetTotalFailures returns correct count",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "src", "tgt", 10)

				// Initially zero failures
				assert.Equal(t, 0, meta.GetTotalFailures())

				// Add failures
				meta.AddFailure("network", "error1")
				meta.AddFailure("network", "error2")
				meta.AddFailure("auth", "error3")

				assert.Equal(t, 3, meta.GetTotalFailures())
			},
		},
		{
			name: "methods handle nil fail map",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := &SyncRunMetainfo{
					CtxID:  1,
					Source: "test",
					Target: "test",
					Total:  1,
					Fail:   nil,
				}

				assert.Nil(t, meta.GetFailuresByType("test"))
				assert.False(t, meta.HasFailures())
				assert.Equal(t, 0, meta.GetTotalFailures())
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

func TestSyncRunMetainfoContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "WithSyncRunMetainfo and GetSyncRunMetainfo work correctly",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(123, "source", "target", 50)
				ctx := context.Background()

				// Add to context
				newCtx := WithSyncRunMetainfo(ctx, meta)
				assert.NotEqual(t, ctx, newCtx)

				// Retrieve from context
				retrievedMeta, ok := GetSyncRunMetainfo(newCtx)
				assert.True(t, ok)
				assert.Equal(t, meta, retrievedMeta)
			},
		},
		{
			name: "GetSyncRunMetainfo returns false for empty context",
			testFunc: func(t *testing.T) {
				t.Helper()
				ctx := context.Background()

				retrievedMeta, ok := GetSyncRunMetainfo(ctx)
				assert.False(t, ok)
				assert.Nil(t, retrievedMeta)
			},
		},
		{
			name: "ContainsFailure works with context",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "src", "tgt", 10)
				meta.AddFailure("invalid", "repo1")
				meta.AddFailure("invalid", "repo2")

				ctx := WithSyncRunMetainfo(context.Background(), meta)

				assert.True(t, ContainsFailure(ctx, "repo1"))
				assert.True(t, ContainsFailure(ctx, "repo2"))
				assert.False(t, ContainsFailure(ctx, "repo3"))
			},
		},
		{
			name: "ContainsFailure returns false for empty context",
			testFunc: func(t *testing.T) {
				t.Helper()
				ctx := context.Background()

				assert.False(t, ContainsFailure(ctx, "anything"))
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

func TestSyncRunMetainfo_EmptyValues_HandlesGracefully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "very large failure maps",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "src", "tgt", 10000)

				// Add many failures
				for range 1000 {
					meta.AddFailure("bulk", "error")
				}

				assert.Equal(t, 1000, meta.GetTotalFailures())
				assert.True(t, meta.HasFailures())

				failures := meta.GetFailuresByType("bulk")
				assert.Len(t, failures, 1000)
			},
		},
		{
			name: "special characters in failure keys and values",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "src", "tgt", 10)

				meta.AddFailure("key@with#special$chars", "value with 项目 unicode 🚀")
				meta.AddFailure("key\nwith\nnewlines", "value\nwith\nnewlines")
				meta.AddFailure("key\twith\ttabs", "value\twith\ttabs")

				assert.Equal(t, 3, meta.GetTotalFailures())

				// String representation should handle special characters
				str := meta.String()
				assert.Contains(t, str, "key@with#special$chars")
				assert.Contains(t, str, "value with 项目 unicode 🚀")
			},
		},
		{
			name: "very long source and target strings",
			testFunc: func(t *testing.T) {
				t.Helper()
				longSource := "https://very-long-domain-name.example.com/organization-with-very-long-name/repository-with-extremely-long-name"
				longTarget := "https://another-very-long-domain.example.com/another-org/another-very-long-repo-name"

				meta := NewSyncRunMetainfo(999, longSource, longTarget, 5000)

				assert.Equal(t, longSource, meta.Source)
				assert.Equal(t, longTarget, meta.Target)

				str := meta.String()
				assert.Contains(t, str, longSource)
				assert.Contains(t, str, longTarget)
			},
		},
		{
			name: "empty failure keys and values",
			testFunc: func(t *testing.T) {
				t.Helper()
				meta := NewSyncRunMetainfo(1, "src", "tgt", 10)

				meta.AddFailure("", "empty key")
				meta.AddFailure("empty_value", "")
				meta.AddFailure("", "")

				assert.Equal(t, 3, meta.GetTotalFailures())

				emptyKeyFailures := meta.GetFailuresByType("")
				assert.Len(t, emptyKeyFailures, 2)
				assert.Equal(t, "empty key", emptyKeyFailures[0])
				assert.Empty(t, emptyKeyFailures[1])

				emptyValueFailures := meta.GetFailuresByType("empty_value")
				assert.Len(t, emptyValueFailures, 1)
				assert.Empty(t, emptyValueFailures[0])
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
