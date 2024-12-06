#!/usr/bin/env bash
# SPDX-FileCopyrightText: Josef Andersson
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
  "github.com/vektra/mockery/v2@v2.50.0"
  "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2"
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

function install_syft() {
  echo -e "${GREEN}Installing Syft...${NC}"
  local install_dir="${HOME}/.local/bin"
  mkdir -p "$install_dir"
  if ! command -v curl &>/dev/null; then
    echo "curl is not installed. Please install curl and try again."
    exit 1
  fi
  curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b "$install_dir"
  echo -e "${YELLOW}Make sure ${install_dir} is in your PATH to use Syft.${NC}"
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
