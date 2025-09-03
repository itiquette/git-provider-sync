// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides pure validation functions for domain entities.
// All functions are pure with no side effects, making them easily testable.
package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Common validation errors.
var (
	ErrEmpty         = errors.New("value cannot be empty")
	ErrTooLong       = errors.New("value is too long")
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidChars  = errors.New("contains invalid characters")
	ErrReservedName  = errors.New("reserved name")
	ErrInvalidURL    = errors.New("invalid URL")
)

// Func is a pure function that validates a string value.
// This is the core building block for composable validation.
type Func func(string) error

// Option configures validation behavior using functional options.
type Option func(*validationConfig)

// validationConfig holds validation configuration.
type validationConfig struct {
	maxLength      int
	minLength      int
	pattern        *regexp.Regexp
	reservedWords  []string
	requiredPrefix string
	requiredSuffix string
}

// WithMaxLength sets maximum length validation.
func WithMaxLength(maxLength int) Option {
	return func(c *validationConfig) {
		c.maxLength = maxLength
	}
}

// WithMinLength sets minimum length validation.
func WithMinLength(minLength int) Option {
	return func(c *validationConfig) {
		c.minLength = minLength
	}
}

// WithPattern sets regex pattern validation.
func WithPattern(pattern string) Option {
	return func(c *validationConfig) {
		c.pattern = regexp.MustCompile(pattern)
	}
}

// WithReservedWords sets reserved words that should be rejected.
func WithReservedWords(words ...string) Option {
	return func(c *validationConfig) {
		c.reservedWords = words
	}
}

// WithRequiredPrefix ensures value starts with prefix.
func WithRequiredPrefix(prefix string) Option {
	return func(c *validationConfig) {
		c.requiredPrefix = prefix
	}
}

// WithRequiredSuffix ensures value ends with suffix.
func WithRequiredSuffix(suffix string) Option {
	return func(c *validationConfig) {
		c.requiredSuffix = suffix
	}
}

// CreateValidator creates a validation function with the given options.
// This is a pure function factory that returns a pure validator.
func CreateValidator(opts ...Option) Func {
	cfg := buildConfig(opts)
	validators := buildValidators(cfg)

	return Compose(validators...)
}

// buildConfig applies options to create config.
func buildConfig(opts []Option) *validationConfig {
	cfg := &validationConfig{
		maxLength: -1,
		minLength: -1,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// buildValidators creates validators from config.
func buildValidators(cfg *validationConfig) []Func {
	var validators []Func

	if cfg.minLength > 0 {
		validators = append(validators, MinLength(cfg.minLength))
	}

	if cfg.maxLength > 0 {
		validators = append(validators, MaxLength(cfg.maxLength))
	}

	if cfg.pattern != nil {
		validators = append(validators, Pattern(cfg.pattern.String()))
	}

	if len(cfg.reservedWords) > 0 {
		validators = append(validators, ReservedWords(cfg.reservedWords...))
	}

	if cfg.requiredPrefix != "" {
		validators = append(validators, Prefix(cfg.requiredPrefix))
	}

	if cfg.requiredSuffix != "" {
		validators = append(validators, Suffix(cfg.requiredSuffix))
	}

	return validators
}

// Compose combines multiple validation functions into one.
// All validators must pass for the composite to pass.
func Compose(validators ...Func) Func {
	return func(value string) error {
		for _, validator := range validators {
			if err := validator(value); err != nil {
				return err
			}
		}

		return nil
	}
}

// Any returns a validator that passes if ANY validator passes.
// Useful for alternative formats.
func Any(validators ...Func) Func {
	return func(value string) error {
		for _, validator := range validators {
			if err := validator(value); err == nil {
				return nil // One passed, that's enough
			}
		}

		return fmt.Errorf("none of %d validators passed", len(validators))
	}
}

// Pure validation functions for common cases

// NotEmpty validates that a string is not empty.
func NotEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrEmpty
	}

	return nil
}

// MaxLength returns a validator for maximum length.
func MaxLength(maxLen int) Func {
	return func(value string) error {
		if len(value) > maxLen {
			return fmt.Errorf("%w: maximum %d characters", ErrTooLong, maxLen)
		}

		return nil
	}
}

// MinLength returns a validator for minimum length.
func MinLength(minLen int) Func {
	return func(value string) error {
		if len(value) < minLen {
			return fmt.Errorf("too short: minimum %d characters", minLen)
		}

		return nil
	}
}

// Pattern returns a validator for regex pattern matching.
func Pattern(pattern string) Func {
	re := regexp.MustCompile(pattern)

	return func(value string) error {
		if !re.MatchString(value) {
			return fmt.Errorf("%w: does not match required pattern", ErrInvalidFormat)
		}

		return nil
	}
}

// AlphaNumeric validates alphanumeric with common separators.
func AlphaNumeric(value string) error {
	return Pattern(`^[a-zA-Z0-9._-]+$`)(value)
}

// URL validates that a string looks like a URL.
func URL(value string) error {
	if value == "" {
		return ErrEmpty
	}

	// Simple URL validation - check for protocol and basic structure
	if !strings.HasPrefix(value, "http://") &&
		!strings.HasPrefix(value, "https://") &&
		!strings.HasPrefix(value, "git@") &&
		!strings.HasPrefix(value, "ssh://") {
		return fmt.Errorf("%w: must start with http://, https://, git@, or ssh://", ErrInvalidURL)
	}

	return nil
}

// HTTPSURL validates that a URL uses HTTPS.
func HTTPSURL(value string) error {
	if !strings.HasPrefix(value, "https://") {
		return fmt.Errorf("%w: must use HTTPS protocol", ErrInvalidURL)
	}

	return nil
}

// GitURL validates Git URLs (HTTPS or SSH).
func GitURL(value string) error {
	if strings.HasPrefix(value, "https://") ||
		strings.HasPrefix(value, "git@") ||
		strings.HasPrefix(value, "ssh://") {
		return nil
	}

	return fmt.Errorf("%w: must be a valid Git URL", ErrInvalidURL)
}

// NoLeadingTrailing returns a validator that checks for leading/trailing characters.
func NoLeadingTrailing(chars string) Func {
	return func(value string) error {
		for _, char := range chars {
			str := string(char)
			if strings.HasPrefix(value, str) {
				return fmt.Errorf("%w: cannot start with '%s'", ErrInvalidFormat, str)
			}

			if strings.HasSuffix(value, str) {
				return fmt.Errorf("%w: cannot end with '%s'", ErrInvalidFormat, str)
			}
		}

		return nil
	}
}

// ReservedWords validates against reserved words.
func ReservedWords(words ...string) Func {
	return func(value string) error {
		valueLower := strings.ToLower(value)
		for _, reserved := range words {
			if valueLower == strings.ToLower(reserved) {
				return fmt.Errorf("%w: '%s'", ErrReservedName, value)
			}
		}

		return nil
	}
}

// Prefix validates that value starts with prefix.
func Prefix(prefix string) Func {
	return func(value string) error {
		if !strings.HasPrefix(value, prefix) {
			return fmt.Errorf("%w: must start with '%s'", ErrInvalidFormat, prefix)
		}

		return nil
	}
}

// Suffix validates that value ends with suffix.
func Suffix(suffix string) Func {
	return func(value string) error {
		if !strings.HasSuffix(value, suffix) {
			return fmt.Errorf("%w: must end with '%s'", ErrInvalidFormat, suffix)
		}

		return nil
	}
}

// Provider-specific validators using composition

// GitHubRepositoryName validates GitHub repository names.
func GitHubRepositoryName(name string) error {
	return Compose(
		NotEmpty,
		MaxLength(100),
		AlphaNumeric,
		NoLeadingTrailing(".-"),
	)(name)
}

// GitLabRepositoryName validates GitLab repository names.
func GitLabRepositoryName(name string) error {
	validator := Compose(
		NotEmpty,
		MaxLength(255),
		Pattern(`^[a-zA-Z0-9][a-zA-Z0-9.+\- ]*$`),
	)

	if err := validator(name); err != nil {
		return err
	}

	// Check reserved names
	reserved := CreateValidator(
		WithReservedWords("badges", "blame", "blob", "builds", "commits",
			"create", "edit", "files", "new", "raw", "refs", "tree", "wikis"),
	)

	return reserved(name)
}

// GiteaRepositoryName validates Gitea repository names.
func GiteaRepositoryName(name string) error {
	return Compose(
		NotEmpty,
		MaxLength(255),
		AlphaNumeric,
		NoLeadingTrailing(".-_"),
	)(name)
}

// ValidatorBuilder provides a fluent API for building validators.
type ValidatorBuilder struct {
	validators []Func
}

// NewValidatorBuilder creates a new validator builder.
func NewValidatorBuilder() *ValidatorBuilder {
	return &ValidatorBuilder{
		validators: []Func{},
	}
}

// Required adds a not-empty check.
func (b *ValidatorBuilder) Required() *ValidatorBuilder {
	b.validators = append(b.validators, NotEmpty)

	return b
}

// Max adds a maximum length check.
func (b *ValidatorBuilder) Max(length int) *ValidatorBuilder {
	b.validators = append(b.validators, MaxLength(length))

	return b
}

// Min adds a minimum length check.
func (b *ValidatorBuilder) Min(length int) *ValidatorBuilder {
	b.validators = append(b.validators, MinLength(length))

	return b
}

// Matches adds a pattern check.
func (b *ValidatorBuilder) Matches(pattern string) *ValidatorBuilder {
	b.validators = append(b.validators, Pattern(pattern))

	return b
}

// Custom adds a custom validation function.
func (b *ValidatorBuilder) Custom(fn Func) *ValidatorBuilder {
	b.validators = append(b.validators, fn)

	return b
}

// Build returns the composed validator.
func (b *ValidatorBuilder) Build() Func {
	return Compose(b.validators...)
}

// Example usage with functional options:
// validator := NewValidatorBuilder().
//     Required().
//     Max(100).
//     Matches(`^[a-z]+$`).
//     Build()
//
// err := validator("test")
