// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	model "itiquette/git-provider-sync/internal/model/configuration"
)

func TestPushOptionRefSpecModification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		force          bool
		expectedPrefix string
	}{
		{
			name:           "force push adds plus prefix",
			force:          true,
			expectedPrefix: "+",
		},
		{
			name:           "normal push has no prefix",
			force:          false,
			expectedPrefix: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			authCfg := model.AuthConfig{Protocol: "https"}
			result := NewPushOption("https://example.com/repo.git", false, testCase.force, authCfg)

			for _, refSpec := range result.RefSpecs {
				if testCase.force {
					assert.Equal(t, byte('+'), refSpec[0], "RefSpec should start with + for force push: %s", refSpec)
				} else {
					assert.NotEqual(t, byte('+'), refSpec[0], "RefSpec should not start with + for normal push: %s", refSpec)
				}
			}
		})
	}
}
