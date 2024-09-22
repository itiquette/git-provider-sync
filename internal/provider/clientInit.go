// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/archive"
	"itiquette/git-provider-sync/internal/provider/directory"
	"itiquette/git-provider-sync/internal/provider/gitea"
	"itiquette/git-provider-sync/internal/provider/github"
	"itiquette/git-provider-sync/internal/provider/gitlab"

	"github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

var ErrNonSupportedProvider = errors.New("unsupported provider")

//nolint:ireturn
func NewGitProviderClient(ctx context.Context, option model.GitProviderClientOption) (interfaces.GitProvider, error) {
	var provider interfaces.GitProvider

	var err error

	httpClient, err := newHTTPClient(option)

	// go-git is setup globally for git, it's by the lib's design
	client.InstallProtocol("https", githttp.NewClient(httpClient))

	switch option.ProviderType {
	case configuration.GITEA:
		provider, err = gitea.NewGiteaClient(ctx, option, httpClient)
	case configuration.GITHUB:
		provider, err = github.NewGitHubClient(ctx, option, httpClient)
	case configuration.GITLAB:
		provider, err = gitlab.NewGitLabClient(ctx, option, httpClient)
	case configuration.ARCHIVE:
		provider = archive.Client{}
	case configuration.DIRECTORY:
		provider = directory.Client{}
	default:
		return nil, ErrNonSupportedProvider
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialized client: %s: %w", option, err)
	}

	return provider, nil
}

func newHTTPClient(option model.GitProviderClientOption) (*http.Client, error) {
	httpClient := &http.Client{}

	var proxyURL *url.URL

	var tlsConfig *tls.Config

	var err error

	if len(option.HTTPClient.CertDirPath) > 0 {
		tlsConfig, err = loadCertsFromDir(option.HTTPClient.CertDirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load certification dir %s. err: %w", option.HTTPClient.CertDirPath, err)
		}
	}

	proxyFunc, err := proxy(option)
	if err != nil {
		return nil, fmt.Errorf("failed to set proxy url %s. err: %w", proxyURL, err)
	}

	httpClient.Transport = &http.Transport{
		Proxy:             proxyFunc,
		ForceAttemptHTTP2: true,
		TLSClientConfig:   tlsConfig,
	}

	return httpClient, nil
}

func proxy(option model.GitProviderClientOption) (func(req *http.Request) (*url.URL, error), error) {
	proxyFunc := http.ProxyFromEnvironment

	if option.HTTPClient.ProxyURL != "" {
		proxyURL, err := url.Parse(option.HTTPClient.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing proxy URL: %w", err)
		}

		proxyFunc = http.ProxyURL(proxyURL)
	}

	return proxyFunc, nil
}

func loadCertsFromDir(dirPath string) (*tls.Config, error) {
	caCertPool := x509.NewCertPool()

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir path. %w", err)
	}

	fmt.Println(files)

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".crt" || filepath.Ext(file.Name()) == ".pem" {
			certPath := filepath.Join(dirPath, file.Name())

			caCert, err := os.ReadFile(certPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file. %w", err)
			}

			caCertPool.AppendCertsFromPEM(caCert)
		}
	}

	tlsConfig := &tls.Config{
		RootCAs:       caCertPool,
		MinVersion:    tls.VersionTLS12,
		MaxVersion:    tls.VersionTLS13,
		ClientAuth:    tls.NoClientCert,
		Renegotiation: tls.RenegotiateNever,
	}

	return tlsConfig, nil
}
