// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package validation_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/validation"
)

func TestFunctionalValidation(t *testing.T) {
	t.Parallel()

	t.Run("NotEmpty", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validation.NotEmpty("test"))
		require.Error(t, validation.NotEmpty(""))
		require.Error(t, validation.NotEmpty("  "))
	})

	t.Run("MaxLength", func(t *testing.T) {
		t.Parallel()

		validator := validation.MaxLength(5)
		require.NoError(t, validator("test"))
		require.NoError(t, validator("12345"))
		require.Error(t, validator("123456"))
	})

	t.Run("Pattern", func(t *testing.T) {
		t.Parallel()

		validator := validation.Pattern(`^[a-z]+$`)
		require.NoError(t, validator("test"))
		require.Error(t, validator("Test"))
		require.Error(t, validator("test123"))
	})

	t.Run("Compose", func(t *testing.T) {
		t.Parallel()

		validator := validation.Compose(
			validation.NotEmpty,
			validation.MaxLength(10),
			validation.AlphaNumeric,
		)

		require.NoError(t, validator("test123"))
		require.Error(t, validator(""))
		require.Error(t, validator("verylongname"))
		require.Error(t, validator("test@123"))
	})

	t.Run("Any", func(t *testing.T) {
		t.Parallel()

		validator := validation.Any(
			validation.HTTPSURL,
			validation.Pattern(`^git@.*`),
		)

		require.NoError(t, validator("https://github.com/repo"))
		require.NoError(t, validator("git@github.com:repo"))
		require.Error(t, validator("http://github.com/repo"))
	})
}

func TestValidationOptions(t *testing.T) {
	t.Parallel()

	t.Run("CreateValidator with options", func(t *testing.T) {
		t.Parallel()

		validator := validation.CreateValidator(
			validation.WithMaxLength(10),
			validation.WithMinLength(3),
			validation.WithPattern(`^[a-z]+$`),
			validation.WithRequiredPrefix("test"),
		)

		require.NoError(t, validator("testfoo"))
		require.Error(t, validator("te"))               // too short
		require.Error(t, validator("testverylongname")) // too long
		require.Error(t, validator("test123"))          // wrong pattern
		require.Error(t, validator("footest"))          // wrong prefix
	})

	t.Run("Reserved words", func(t *testing.T) {
		t.Parallel()

		validator := validation.CreateValidator(
			validation.WithReservedWords("admin", "root", "system"),
		)

		require.NoError(t, validator("user"))
		require.Error(t, validator("admin"))
		require.Error(t, validator("ADMIN")) // case insensitive
		require.Error(t, validator("root"))
	})
}

func TestProviderValidators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		validator func(string) error
		valid     []string
		invalid   []string
	}{
		{
			name:      "GitHub repository name",
			validator: validation.GitHubRepositoryName,
			valid:     []string{"repo", "my-repo", "my_repo", "repo.js", "123"},
			invalid:   []string{"", "-repo", "repo-", ".repo", "repo.", strings.Repeat("a", 101)},
		},
		{
			name:      "GitLab repository name",
			validator: validation.GitLabRepositoryName,
			valid:     []string{"repo", "my-repo", "my repo", "repo+plus"},
			invalid:   []string{"", "badges", "new", "edit", strings.Repeat("a", 256)},
		},
		{
			name:      "Gitea repository name",
			validator: validation.GiteaRepositoryName,
			valid:     []string{"repo", "my-repo", "my_repo", "repo.js"},
			invalid:   []string{"", "-repo", "_repo", ".repo", strings.Repeat("a", 256)},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			for _, valid := range testCase.valid {
				require.NoError(t, testCase.validator(valid), "should accept: %s", valid)
			}

			for _, invalid := range testCase.invalid {
				require.Error(t, testCase.validator(invalid), "should reject: %s", invalid)
			}
		})
	}
}

func TestValidatorBuilder(t *testing.T) {
	t.Parallel()

	t.Run("Builder pattern", func(t *testing.T) {
		t.Parallel()

		validator := validation.NewValidatorBuilder().
			Required().
			Max(20).
			Min(5).
			Matches(`^[a-z]+$`).
			Build()

		require.NoError(t, validator("hello"))
		require.NoError(t, validator("world"))
		require.Error(t, validator(""))                      // not required
		require.Error(t, validator("hi"))                    // too short
		require.Error(t, validator(strings.Repeat("a", 21))) // too long
		require.Error(t, validator("Hello"))                 // uppercase
	})

	t.Run("Custom validator", func(t *testing.T) {
		t.Parallel()

		palindrome := func(s string) error {
			runes := []rune(s)
			for i := range len(runes) / 2 {
				if runes[i] != runes[len(runes)-1-i] {
					return validation.ErrInvalidFormat
				}
			}

			return nil
		}

		validator := validation.NewValidatorBuilder().
			Required().
			Custom(palindrome).
			Build()

		require.NoError(t, validator("racecar"))
		require.NoError(t, validator("noon"))
		require.Error(t, validator("hello"))
	})
}

func TestURLValidators(t *testing.T) {
	t.Parallel()

	t.Run("HTTPS URL", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validation.HTTPSURL("https://example.com"))
		require.Error(t, validation.HTTPSURL("http://example.com"))
		require.Error(t, validation.HTTPSURL("ftp://example.com"))
	})

	t.Run("Git URL", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validation.GitURL("https://github.com/user/repo"))
		require.NoError(t, validation.GitURL("git@github.com:user/repo"))
		require.NoError(t, validation.GitURL("ssh://git@github.com/user/repo"))
		require.Error(t, validation.GitURL("http://github.com/user/repo"))
		require.Error(t, validation.GitURL("ftp://github.com/user/repo"))
	})
}

func TestNoLeadingTrailing(t *testing.T) {
	t.Parallel()

	validator := validation.NoLeadingTrailing(".-_")

	require.NoError(t, validator("test"))
	require.NoError(t, validator("test-name"))
	require.Error(t, validator("-test"))
	require.Error(t, validator("test-"))
	require.Error(t, validator(".test"))
	require.Error(t, validator("test."))
	require.Error(t, validator("_test"))
	require.Error(t, validator("test_"))
}

func BenchmarkValidation(b *testing.B) {
	simple := validation.MaxLength(100)
	composed := validation.Compose(
		validation.NotEmpty,
		validation.MaxLength(100),
		validation.AlphaNumeric,
	)

	builder := validation.NewValidatorBuilder().
		Required().
		Max(100).
		Matches(`^[a-zA-Z0-9._-]+$`).
		Build()

	b.Run("simple", func(b *testing.B) {
		for range b.N {
			_ = simple("test-repo-name")
		}
	})

	b.Run("composed", func(b *testing.B) {
		for range b.N {
			_ = composed("test-repo-name")
		}
	})

	b.Run("builder", func(b *testing.B) {
		for range b.N {
			_ = builder("test-repo-name")
		}
	})
}
