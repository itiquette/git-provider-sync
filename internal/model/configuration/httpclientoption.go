// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
)

// HTTPClientOption represents HTTP client configuration options.
type HTTPClientOption struct {
	Scheme      string `koanf:"scheme"`
	ProxyURL    string `koanf:"proxyurl"`
	Token       string `koanf:"token"`
	CertDirPath string `koanf:"certdirpath"`
}

func (p HTTPClientOption) String() string {
	return fmt.Sprintf("HTTPClientOption: ProxyURL %s, Token: %s",
		p.ProxyURL, maskToken())
}

func maskToken() string {
	return "<****>"
}
