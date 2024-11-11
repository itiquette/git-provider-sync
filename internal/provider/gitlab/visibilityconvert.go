// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import "github.com/xanzy/go-gitlab"

func getVisibility(vis gitlab.VisibilityValue) string {
	switch vis {
	case gitlab.PublicVisibility:
		return "public"
	case gitlab.PrivateVisibility:
		return "private"
	case gitlab.InternalVisibility:
		return "internal"
	default:
		return "public"
	}
}

func toVisibility(vis string) gitlab.VisibilityValue {
	switch vis {
	case "private":
		return gitlab.PrivateVisibility
	case "internal":
		return gitlab.InternalVisibility
	default:
		return gitlab.PublicVisibility
	}
}
