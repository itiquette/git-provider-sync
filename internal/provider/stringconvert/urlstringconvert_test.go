// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package stringconvert

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_AddBasicAuthToURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		username string
		password string
		want     string
	}{
		{
			name:     "simple http url",
			url:      "http://example.com",
			username: "user",
			password: "pass",
			want:     "http://user:pass@example.com",
		},
		{
			name:     "https url with path",
			url:      "https://example.com/repo.git",
			username: "user",
			password: "pass",
			want:     "https://user:pass@example.com/repo.git",
		},
		{
			name:     "url with existing auth",
			url:      "https://old:auth@example.com",
			username: "user",
			password: "pass",
			want:     "https://user:pass@example.com",
		},
		{
			name:     "url with special characters",
			url:      "https://example.com",
			username: "user@domain",
			password: "pass:word!",
			want:     "https://user%40domain:pass%3Aword%21@example.com",
		},
		{
			name:     "url with port",
			url:      "https://example.com:8080",
			username: "user",
			password: "pass",
			want:     "https://user:pass@example.com:8080",
		},
		{
			name:     "empty url",
			url:      "",
			username: "user",
			password: "pass",
			want:     "//user:pass@",
		},
	}

	ctx := context.Background()

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := AddBasicAuthToURL(ctx, tabletest.url, tabletest.username, tabletest.password)
			assert.Equal(t, tabletest.want, got)
		})
	}
}

func TestAuthService_RemoveBasicAuthFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "url with auth",
			url:  "https://user:pass@example.com",
			want: "https://user:SECRET@example.com",
		},
		{
			name: "url without auth",
			url:  "https://example.com",
			want: "https://example.com",
		},
		{
			name: "url with path and auth",
			url:  "https://user:pass@example.com/repo.git",
			want: "https://user:SECRET@example.com/repo.git",
		},
		{
			name: "url with special characters in auth",
			url:  "https://user%40domain:pass%3Aword%21@example.com",
			want: "https://user%40domain:SECRET@example.com",
		},
		{
			name: "url with port and auth",
			url:  "https://user:pass@example.com:8080",
			want: "https://user:SECRET@example.com:8080",
		},
		{
			name: "invalid url",
			url:  "://invalid",
			want: "://invalid",
		},
		{
			name: "empty url",
			url:  "",
			want: "",
		},
	}

	ctx := context.Background()

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := RemoveBasicAuthFromURL(ctx, tabletest.url, false)
			assert.Equal(t, tabletest.want, got)
		})
	}
}
