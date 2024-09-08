// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
)

func newGitGoPullOption(name string, url string) git.PullOptions {
	return git.PullOptions{
		RemoteName: name,
		RemoteURL:  url,
	}
}

func newGitGoCloneOption(url string, mirror bool) git.CloneOptions {
	return git.CloneOptions{
		URL:    url,
		Mirror: mirror,
	}
}

func newGitGoPushOption(url string, refSpec []string, prune bool) git.PushOptions {
	refSpecs := make([]gogitconfig.RefSpec, 0, 20)

	for _, r := range refSpec {
		refSpec := gogitconfig.RefSpec(r)
		refSpecs = append(refSpecs, refSpec)
	}

	return git.PushOptions{
		RemoteURL: url,
		RefSpecs:  refSpecs,
		Prune:     prune, //TO-DO: open an issue - wont allow for protected branch
	}
}
