// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package shared

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	// OAuth token patterns.
	oauthTokenPattern = regexp.MustCompile(`\b(oauth2?|token|bearer|api[_-]?key|access[_-]?token|auth[_-]?token)[:\s=]+[\w\-\.]+\b`)
	// Git URL with embedded credentials pattern.
	gitCredPattern = regexp.MustCompile(`(https?://)([^:/@]+):([^@]+)@`)
	// Basic auth in URL pattern.
	basicAuthPattern = regexp.MustCompile(`(https?://)([^:]+):([^@]+)@`)
	// SSH URL pattern - preserve user but hide key info
	// SshPattern = regexp.MustCompile(`(ssh://)?([^@]+)@([^:]+):(.+)`) // Currently unused
	// Authorization header pattern.
	authHeaderPattern = regexp.MustCompile(`(?i)(authorization|auth|api[_-]?key|x-api-key|x-auth-token):\s*[^\n\r]*`)
	// Token in query parameters.
	queryTokenPattern = regexp.MustCompile(`([?&])(token|api[_-]?key|access[_-]?token|auth|key)=([^&\s]+)`)
)

// SanitizeURL removes sensitive credentials from URLs
// handles:
// - Basic auth (https://user:pass@domain.com -> https://***:***@domain.com)
// - OAuth tokens in URLs
// - Git credentials
// - SSH URLs (preserves structure but hides sensitive parts).
func SanitizeURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// First try to parse as URL for structured handling
	if parsedURL, err := url.Parse(rawURL); err == nil {
		// Handle userinfo (username:password)
		if parsedURL.User != nil {
			// Always replace with ***:*** to hide both username and password
			parsedURL.User = url.UserPassword("***", "***")
		}

		// Sanitize query parameters
		if parsedURL.RawQuery != "" {
			parsedURL.RawQuery = sanitizeQueryParams(parsedURL.RawQuery)
		}

		sanitized := parsedURL.String()

		// Additional pattern-based sanitization for cases URL parsing might miss
		sanitized = sanitizePatterns(sanitized)

		return sanitized
	}

	// If URL parsing fails, fall back to pattern-based sanitization
	return sanitizePatterns(rawURL)
}

// SanitizeQueryParams removes sensitive parameters from query strings.
func sanitizeQueryParams(query string) string {
	params, err := url.ParseQuery(query)
	if err != nil {
		// If parsing fails, do pattern-based replacement
		return queryTokenPattern.ReplaceAllString(query, "${1}${2}=***")
	}

	// List of sensitive parameter names
	sensitiveParams := []string{
		"token", "api_key", "apikey", "api-key",
		"access_token", "access-token", "auth_token", "auth-token",
		"key", "secret", "password", "pwd", "pass",
		"authorization", "auth", "credential",
	}

	for _, param := range sensitiveParams {
		if _, exists := params[param]; exists {
			params.Set(param, "***")
		}
		// Also check with different cases
		upperParam := strings.ToUpper(param)
		if _, exists := params[upperParam]; exists {
			params.Set(upperParam, "***")
		}

		lowerParam := strings.ToLower(param)
		if _, exists := params[lowerParam]; exists {
			params.Set(lowerParam, "***")
		}
	}

	return params.Encode()
}

// SanitizePatterns applies regex patterns to sanitize various credential formats.
func sanitizePatterns(str string) string {
	// Replace basic auth credentials
	str = basicAuthPattern.ReplaceAllString(str, "${1}***:***@")

	// Replace git credentials (might have oauth2 as username)
	str = gitCredPattern.ReplaceAllString(str, "${1}***:***@")

	// Replace OAuth tokens
	str = oauthTokenPattern.ReplaceAllStringFunc(str, func(match string) string {
		parts := regexp.MustCompile(`[:\s=]+`).Split(match, 2)
		if len(parts) == 2 {
			return parts[0] + ":***"
		}

		return match
	})

	// Replace authorization headers
	str = authHeaderPattern.ReplaceAllStringFunc(str, func(match string) string {
		// Find the colon position
		colonIndex := strings.Index(strings.ToLower(match), ":")
		if colonIndex != -1 {
			return match[:colonIndex] + ": ***"
		}

		return match
	})

	// Replace tokens in query parameters
	str = queryTokenPattern.ReplaceAllString(str, "${1}${2}=***")

	return str
}

// SanitizeStringMap sanitizes all URLs and potential secrets in a map.
func SanitizeStringMap(inputMap map[string]any) map[string]any {
	if inputMap == nil {
		return nil
	}

	result := make(map[string]any, len(inputMap))
	for key, value := range inputMap {
		result[key] = sanitizeMapValue(key, value)
	}

	return result
}

// SanitizeMapValue sanitizes a single map value based on its key and type.
func sanitizeMapValue(key string, value any) any {
	// Check if key suggests it contains sensitive data
	lowerKey := strings.ToLower(key)
	if containsSensitiveKey(lowerKey) {
		return "***"
	}

	switch val := value.(type) {
	case string:
		return sanitizeStringValue(val)
	case map[string]any:
		return SanitizeStringMap(val)
	default:
		return value
	}
}

// SanitizeStringValue sanitizes a string value that might contain sensitive data.
func sanitizeStringValue(val string) string {
	switch {
	case looksLikeURL(val) || containsCredentials(val):
		return SanitizeURL(val)
	case looksLikeToken(val):
		return "***"
	default:
		return val
	}
}

// ContainsSensitiveKey checks if a key name suggests sensitive data.
func containsSensitiveKey(key string) bool {
	// Exact matches for common sensitive field names
	exactMatches := []string{
		"password", "passwd", "pwd", "secret", "token",
		"auth", "credential", "private", "signature",
	}

	for _, match := range exactMatches {
		if key == match {
			return true
		}
	}

	// Contains checks for compound names
	sensitivePatterns := []string{
		"password", "passwd", "pwd", "secret", "token",
		"api_key", "apikey", "api-key",
		"access_token", "auth_token", "access-token", "auth-token",
		"private_key", "private-key", "credential",
		"authorization", "signature",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(key, pattern) {
			return true
		}
	}

	// "key" alone is too generic, only flag it if it's part of a compound
	if key == "key" {
		return false // generic "key" is not sensitive
	}

	return false
}

// LooksLikeURL checks if a string appears to be a URL.
func looksLikeURL(str string) bool {
	return strings.HasPrefix(str, "http://") ||
		strings.HasPrefix(str, "https://") ||
		strings.HasPrefix(str, "ssh://") ||
		strings.HasPrefix(str, "git://") ||
		strings.HasPrefix(str, "git@") ||
		strings.Contains(str, "://")
}

// ContainsURL checks if a string contains URL patterns (exported for use in logging).
func ContainsURL(str string) bool {
	return strings.Contains(str, "://") ||
		strings.Contains(str, "@") && (strings.Contains(str, "git@") || strings.Contains(str, ":"))
}

// ContainsCredentials checks if a string contains credential patterns.
func containsCredentials(str string) bool {
	// Check for basic auth pattern
	if strings.Contains(str, "@") && strings.Contains(str, ":") {
		// Could be user:pass@domain
		parts := strings.Split(str, "@")
		if len(parts) >= 2 {
			userPart := parts[0]
			if strings.Contains(userPart, ":") {
				return true
			}
		}
	}

	// Check for oauth/token patterns
	lowerStr := strings.ToLower(str)
	credentialIndicators := []string{
		"oauth", "bearer", "token", "api_key", "apikey",
		"access_token", "auth_token", "authorization",
	}

	for _, indicator := range credentialIndicators {
		if strings.Contains(lowerStr, indicator) {
			return true
		}
	}

	return false
}

// LooksLikeToken checks if a string appears to be a token.
func looksLikeToken(str string) bool {
	// Tokens are typically long random strings
	// Check if it's a long string with mixed alphanumeric characters
	if len(str) < 20 {
		return false
	}

	// Check if it looks like a token (mix of letters and numbers, possibly with - or _)
	if matched, _ := regexp.MatchString(`^[\w\-\.]+$`, str); matched {
		// Has both letters and numbers?
		hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(str)
		hasNumber := regexp.MustCompile(`\d`).MatchString(str)

		return hasLetter && hasNumber && len(str) > 30
	}

	return false
}

// SanitizeError sanitizes error messages to remove sensitive information.
func SanitizeError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Sanitize URLs in error message
	errStr = sanitizePatterns(errStr)

	// Additional sanitization for common error patterns
	// Git errors often include URLs
	if strings.Contains(errStr, "://") {
		words := strings.Fields(errStr)
		for i, word := range words {
			if looksLikeURL(word) {
				words[i] = SanitizeURL(word)
			}
		}

		errStr = strings.Join(words, " ")
	}

	return errStr
}
