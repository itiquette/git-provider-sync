// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package model provides configuration model types.
package model

import (
	"strconv"
	"strings"
)

// SyncRunOption represents synchronization run configuration options.
type SyncRunOption struct {
	ForcePush         bool   `koanf:"forcepush"`
	IgnoreInvalidName bool   `koanf:"ignoreinvalidname"`
	AlphaNumHyphName  bool   `koanf:"alphanumhyph_name"`
	ActiveFromLimit   string `koanf:"activefromlimit"`
}

func (p SyncRunOption) String() string {
	var parts []string

	parts = append(parts, "SyncRunOption{")
	parts = append(parts, "ForcePush: "+strconv.FormatBool(p.ForcePush))
	parts = append(parts, "IgnoreInvalidName: "+strconv.FormatBool(p.IgnoreInvalidName))
	parts = append(parts, "ASCIIName: "+strconv.FormatBool(p.AlphaNumHyphName))

	if p.ActiveFromLimit != "" {
		parts = append(parts, "ActiveFromLimit: "+p.ActiveFromLimit)
	}

	parts = append(parts, "}")

	return strings.Join(parts, " ")
}
