// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func newGitLibPullOption(name string, url string, auth transport.AuthMethod) git.PullOptions {
	return git.PullOptions{
		RemoteName: name,
		RemoteURL:  url,
		Auth:       auth,
	}
}

func newGitLibCloneOption(url string, mirror bool, auth transport.AuthMethod) git.CloneOptions {
	return git.CloneOptions{
		Auth:   auth,
		URL:    url,
		Mirror: mirror,
	}
}

func newGitLibPushOption(url string, refSpec []string, prune bool, auth transport.AuthMethod) git.PushOptions {
	refSpecs := make([]gogitconfig.RefSpec, 0, 20)

	for _, r := range refSpec {
		refSpec := gogitconfig.RefSpec(r)
		refSpecs = append(refSpecs, refSpec)
	}

	return git.PushOptions{
		Auth:      auth,
		RemoteURL: url,
		RefSpecs:  refSpecs,
		Prune:     prune, //TO-DO: open an issue - wont allow for protected branch
	}
}
