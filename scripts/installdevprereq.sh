#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 itiquette/git-provider-sync
#
# SPDX-License-Identifier: CC0-1.0

# Description: This script installs various Go tools and Syft for the gitprovidersync project.
#              It ensures that all necessary development and analysis tools are available.
#
# Usage: ./install_go_tools.sh
#
# The script performs the following actions:
# 1. Installs a set of Go tools using 'go install', including:
#    - mockery: For generating mock interfaces
#    - golangci-lint: A Go linter
#    - goreleaser: For releasing Go projects
#    - wsl: Whitespace linter
#    - cosign: For signing and verifying software artifacts
#    - staticcheck: A Go static analysis tool
#    - govulncheck: For checking Go vulnerabilities
#    - scorecard: For checking GitHub project best practices
# 2. Installs Syft (a software bill of materials generator) in ~/.local/bin
# 3. Runs 'go mod tidy' to ensure the go.mod file is up to date
#
# Dependencies:
# - Go: This script requires Go to be installed and properly configured
# - curl: Required for installing Syft
#
# Note:
# - After running the script, make sure to add ~/.local/bin to your PATH if it's not already there
#

set -euo pipefail

# Define color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Define Go tools to install
GO_TOOLS=(
  "github.com/vektra/mockery/v2@v2.53.3"
  "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
  "github.com/goreleaser/goreleaser/v2@latest"
  "github.com/bombsimon/wsl/v4/cmd/wsl@master"
  "github.com/sigstore/cosign/v2/cmd/cosign@latest"
  "honnef.co/go/tools/cmd/staticcheck@latest"
  "golang.org/x/vuln/cmd/govulncheck@latest"
  "github.com/ossf/scorecard@main"
)

function install_go_tools() {
  echo -e "${GREEN}Installing Go tools...${NC}"
  for tool in "${GO_TOOLS[@]}"; do
    echo "Installing $tool"
    go install "$tool"
  done
}

install_syft() {
  echo -e "${GREEN}Installing Syft...${NC}"
  local install_dir="${HOME}/.local/bin"
  local version="v1.19.0"
  local arch="linux_arm64"

  mkdir -p "$install_dir"
  [[ -x "$(command -v curl)" ]] || {
    echo "curl required"
    exit 1
  }

  local base_url="https://github.com/anchore/syft/releases/download/${version}"
  local tarball="syft_${version#v}_${arch}.tar.gz"
  local checksums="syft_${version#v}_checksums.txt"

  (
    cd /tmp
    curl -sSLO "${base_url}/${tarball}"
    curl -sSLO "${base_url}/${checksums}"

    grep "${tarball}" "${checksums}" | sha256sum -c || {
      echo "Checksum verification failed"
      exit 1
    }

    tar -xzf "${tarball}" -C "${install_dir}" syft
  )
  [[ -x "${install_dir}/syft" ]] && echo -e "${YELLOW}Syft installed to ${install_dir}${NC}" || echo "Installation failed"
}

function main() {
  install_go_tools
  install_syft

  echo -e "${GREEN}Running go mod tidy...${NC}"
  go mod tidy

  echo -e "${GREEN}All pre-requisites installed successfully!${NC}"
  echo -e "${YELLOW}Note: Ensure ${HOME}/.local/bin is in your PATH for Syft to be accessible.${NC}"
}

main
