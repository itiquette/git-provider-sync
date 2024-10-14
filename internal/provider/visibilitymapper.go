// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"fmt"
	"strings"
)

func mapVisibility(fromProvider, toProvider, visibility string) (string, error) {
	// Normalize inputs to lowercase
	fromProvider = strings.ToLower(fromProvider)
	toProvider = strings.ToLower(toProvider)
	visibility = strings.ToLower(visibility)

	if strings.EqualFold(fromProvider, toProvider) {
		return visibility, nil
	}

	// Define mapping based on the provided tables
	mappings := map[string]map[string]map[string]string{
		"gitlab": {
			"github": {"public": "public", "internal": "private", "private": "private"},
			"gitea":  {"public": "public", "internal": "private", "private": "private"},
		},
		"github": {
			"gitlab": {"public": "public", "private": "private"},
			"gitea":  {"public": "public", "private": "private"},
		},
		"gitea": {
			"gitlab": {"public": "public", "private": "private", "limited": "private"},
			"github": {"public": "public", "private": "private", "limited": "private"},
		},
	}

	// Check if the fromProvider is valid
	providerMap, isOK := mappings[fromProvider]
	if !isOK {
		return "", fmt.Errorf("invalid source provider: %s", fromProvider)
	}

	// Check if the toProvider is valid
	targetMap, isOK := providerMap[toProvider]
	if !isOK {
		return "", fmt.Errorf("invalid target provider: %s", toProvider)
	}

	// Check if the visibility is valid and get the mapped visibility
	mappedVisibility, isOK := targetMap[visibility]
	if !isOK {
		return "", fmt.Errorf("invalid visibility for %s: %s", fromProvider, visibility)
	}

	return mappedVisibility, nil
}
