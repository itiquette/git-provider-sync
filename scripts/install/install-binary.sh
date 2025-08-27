#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Project-agnostic binary installer with automatic architecture detection
# Usage: ./scripts/install-binary.sh [binary_name] [bin_dir] [install_dir]
# Defaults: binary_name=$(basename "$(pwd)"), bin_dir=./bin, install_dir=$HOME/.local/bin

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly NC=$'\033[0m' # No Color

# Parse arguments with defaults
readonly BINARY_NAME="${1:-${BINARY_NAME:-$(basename "$(pwd)")}}"
readonly BIN_DIR="${2:-${BIN_DIR:-./bin}}"
readonly INSTALL_DIR="${3:-${INSTALL_DIR:-$HOME/.local/bin}}"

log() {
  printf "${YELLOW}▸${NC} %s\n" "$1"
}

success() {
  printf "${GREEN}✓${NC} %s\n" "$1"
}

fail() {
  printf "${RED}✗${NC} %s\n" "$1" >&2
  exit 1
}

# Detect system architecture
detect_architecture() {
  local arch
  arch=$(uname -m)

  case $arch in
  x86_64)
    echo "amd64"
    ;;
  aarch64 | arm64)
    echo "arm64"
    ;;
  *)
    fail "Unsupported architecture: $arch"
    ;;
  esac
}

main() {
  # Detect architecture
  local arch
  arch=$(detect_architecture)
  local os="${OS:-linux}"

  # Find source binary - try arch-specific first, then generic
  local source_binary="${BIN_DIR}/${BINARY_NAME}-${os}-${arch}"
  if [[ ! -f "$source_binary" ]]; then
    source_binary="${BIN_DIR}/${BINARY_NAME}"
    if [[ ! -f "$source_binary" ]]; then
      fail "Binary not found: tried ${BINARY_NAME}-${os}-${arch} and ${BINARY_NAME} in ${BIN_DIR}"
    fi
  fi

  # Create install directory if it doesn't exist
  if [[ ! -d "$INSTALL_DIR" ]]; then
    log "Creating directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR" || fail "Failed to create directory: $INSTALL_DIR"
  fi

  # Install binary
  local dest="${INSTALL_DIR}/${BINARY_NAME}"
  log "Installing ${BINARY_NAME} (${arch}) to ${dest}"

  cp "$source_binary" "$dest" || fail "Failed to copy binary"
  chmod +x "$dest" || fail "Failed to make binary executable"

  success "Successfully installed ${BINARY_NAME} to ${dest}"

  # Check if install dir is in PATH
  if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    printf "${YELLOW}Note:${NC} ${INSTALL_DIR} is not in your PATH\n"
    printf "Add it with: export PATH=\"\$PATH:${INSTALL_DIR}\"\n"
  fi
}

main "$@"
