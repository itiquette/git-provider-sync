// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package stringconvert

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveNonAlphaNumericChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"remove non alphanumeric chars", "!a@1l#&()pha -[{num}]:;eric',?/*sy`~m$^b+1ols=<>", "a1lpha-numericsymb1ols"},
		{"empty string", "", ""},
		{"only alphanumeric", "abc123", "abc123"},
		{"with underscores and hyphens", "hello_world-123", "hello_world-123"},
	}

	ctx := context.Background()

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := RemoveNonAlphaNumericChars(ctx, tabletest.input)
			assert.Equal(t, tabletest.want, got)
		})
	}
}

func TestRemoveLinebreaks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"CRLF", "hello\r\nworld", "hello world"},
		{"LF", "hello\nworld", "hello world"},
		{"CR", "hello\rworld", "hello world"},
		{"mixed", "hello\r\nworld\ntest\rline", "hello world test line"},
		{"no linebreaks", "helloworld", "helloworld"},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := RemoveLinebreaks(tabletest.input)
			assert.Equal(t, tabletest.want, got)
		})
	}
}

func TestFileNameWithoutExt(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     string
	}{
		{"with extension", "file.txt", "file"},
		{"without extension", "file", "file"},
		{"multiple dots", "file.name.txt", "file.name"},
		{"hidden file", ".gitignore", ".gitignore"},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := FileNameWithoutExt(tabletest.fileName)
			assert.Equal(t, tabletest.want, got)
		})
	}
}

func TestCleanString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"mixed content", "Hello, \nWorld! 123", "HelloWorld123"},
		{"empty string", "", ""},
		{"only spaces and linebreaks", " \n\r\t", ""},
	}

	ctx := context.Background()

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := CleanString(ctx, tabletest.input)
			assert.Equal(t, tabletest.want, got)
		})
	}
}
