// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package model

import (
	"strconv"
	"strings"
)

type SyncRunOption struct {
	ForcePush          bool   `koanf:"forcepush"`
	IgnoreInvalidName  bool   `koanf:"ignoreinvalidname"`
	CleanupInvalidName bool   `koanf:"cleanupinvalidname"`
	ActiveFromLimit    string `koanf:"activefromlimit"`
}

func (p SyncRunOption) String() string {
	var parts []string
	parts = append(parts, "SyncRunOption{")
	parts = append(parts, "ForcePush: "+strconv.FormatBool(p.ForcePush))
	parts = append(parts, "IgnoreInvalidName: "+strconv.FormatBool(p.IgnoreInvalidName))
	parts = append(parts, "CleanupInvalidName: "+strconv.FormatBool(p.CleanupInvalidName))

	if p.ActiveFromLimit != "" {
		parts = append(parts, "ActiveFromLimit: "+p.ActiveFromLimit)
	}

	parts = append(parts, "}")

	return strings.Join(parts, " ")
}
