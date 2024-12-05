// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlib

import (
	"github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func (serv *Service) buildCloneOptions(url string, mirror bool, auth transport.AuthMethod) *git.CloneOptions {
	return &git.CloneOptions{
		Auth:   auth,
		Mirror: mirror,
		URL:    url,
	}
}

func (serv *Service) buildPullOptions(remote string, url string, auth transport.AuthMethod) *git.PullOptions {
	return &git.PullOptions{
		Auth:       auth,
		RemoteURL:  url,
		RemoteName: remote,
	}
}

func (serv *Service) buildPushOptions(url string, refSpec []string, prune bool, auth transport.AuthMethod) git.PushOptions {
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
