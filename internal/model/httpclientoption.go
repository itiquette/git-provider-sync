// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "fmt"

type HTTPClientOption struct {
	ProxyURL string `koanf:"proxyurl"`
	Token    string `koanf:"token"`
}

func (p HTTPClientOption) String() string {
	return fmt.Sprintf("HTTPClientOption: ProxyURL %s, Token: %s",
		p.ProxyURL, maskToken(p.Token))
}
