// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"strings"
)

type HTTPClientOption struct {
	ProxyURL    string `koanf:"proxyurl"`
	Token       string `koanf:"token"`
	CertDirPath string `koanf:"certdirpath"`
}

func (p HTTPClientOption) String() string {
	return fmt.Sprintf("HTTPClientOption: ProxyURL %s, Token: %s",
		p.ProxyURL, maskToken(p.Token))
}

// maskToken is a helper function that masks all but the last 4 characters of a token.
// If the token is 4 characters or less, it masks all characters.
func maskToken(token string) string {
	if len(token) <= 4 {
		return strings.Repeat("*", len(token))
	}

	return strings.Repeat("*", len(token)-4) + token[len(token)-4:]
}
