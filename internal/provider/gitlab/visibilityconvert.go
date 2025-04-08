// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import gitlab "gitlab.com/gitlab-org/api/client-go"

var PUBLIC = "public"

func getVisibility(vis gitlab.VisibilityValue) string {
	switch vis {
	case gitlab.PublicVisibility:
		return PUBLIC
	case gitlab.PrivateVisibility:
		return "private"
	case gitlab.InternalVisibility:
		return "internal"
	default:
		return PUBLIC
	}
}

func toVisibility(vis string) gitlab.VisibilityValue {
	switch vis {
	case "private":
		return gitlab.PrivateVisibility
	case "internal":
		return gitlab.InternalVisibility
	case PUBLIC:
		return gitlab.PublicVisibility
	default:
		return gitlab.InternalVisibility
	}
}
