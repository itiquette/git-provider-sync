#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Project-agnostic cross-platform binary builder
# Usage: ./scripts/build-cross-platform.sh [goos] [goarch] [bin_dir] [binary_name] [main_path]
# Defaults: goos=linux, goarch=amd64, bin_dir=./bin, binary_name=<project>, main_path=cmd/<binary>/main.go or main.go

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly CYAN=$'\033[1;36m'
readonly NC=$'\033[0m' # No Color

# Parse arguments with defaults
TARGET_OS="${1:-${GOOS:-linux}}"
TARGET_ARCH="${2:-${GOARCH:-amd64}}"
readonly BIN_DIR="${3:-${BIN_DIR:-./bin}}"
readonly BINARY_NAME="${4:-${BINARY_NAME:-$(basename "$(pwd)")}}"
readonly MAIN_PATH="${5:-${MAIN_PATH:-}}"

# Build configuration
readonly BUILD_TAGS="netgo,osusergo" # Pure Go networking and user resolution
readonly BUILDMODE_DEFAULT="pie"     # Position Independent Executable for security
export CGO_ENABLED="0"               # Static binaries

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

# Find main package
find_main_package() {
  local binary_name="$1"
  local custom_path="$2"

  # If custom path provided, use it
  if [[ -n "$custom_path" ]]; then
    if [[ -f "$custom_path" ]] || [[ -d "$custom_path" ]]; then
      echo "$custom_path"
      return 0
    fi
  fi

  # Try common locations
  local locations=(
    "cmd/${binary_name}/main.go"
    "cmd/${binary_name}"
    "main.go"
    "."
  )

  for loc in "${locations[@]}"; do
    if [[ -f "$loc" ]] || [[ -d "$loc" ]]; then
      echo "$loc"
      return 0
    fi
  done

  return 1
}

# Get build mode based on OS
get_build_mode() {
  local os="$1"

  # BSD platforms don't support PIE without CGO
  case "$os" in
  freebsd | openbsd | netbsd)
    echo "exe"
    ;;
  *)
    echo "$BUILDMODE_DEFAULT"
    ;;
  esac
}

# Generate build metadata
generate_build_metadata() {
  local version commit build_date

  # Try git for version info
  if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
    version=$(git describe --tags --dirty --always 2>/dev/null || echo "dev")
    commit=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    build_date=$(git log -1 --format=%cI 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ)
  else
    version="dev"
    commit="unknown"
    build_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  fi

  echo "$version|$commit|$build_date"
}

# Execute build
execute_build() {
  local output_binary="$BIN_DIR/$BINARY_NAME-$TARGET_OS-$TARGET_ARCH"
  [[ "$TARGET_OS" == "windows" ]] && output_binary="${output_binary}.exe"

  local main_pkg
  main_pkg=$(find_main_package "$BINARY_NAME" "$MAIN_PATH") || {
    fail "Cannot find main package for $BINARY_NAME"
  }

  local buildmode
  buildmode=$(get_build_mode "$TARGET_OS")

  # Get metadata
  IFS='|' read -r version commit build_date <<<"$(generate_build_metadata)"

  # Build flags
  local ldflags="-w -buildid="
  ldflags="$ldflags -X main.version=$version"
  ldflags="$ldflags -X main.commit=$commit"
  ldflags="$ldflags -X main.date=$build_date"

  log "Building $BINARY_NAME for $TARGET_OS/$TARGET_ARCH"
  log "Main package: $main_pkg"
  log "Output: $output_binary"

  # Apply CPU optimizations
  case "$TARGET_ARCH" in
  amd64)
    export GOAMD64="${GOAMD64:-v2}"
    log "AMD64 optimization: $GOAMD64"
    ;;
  arm64)
    export GOARM64="${GOARM64:-v8.0}"
    log "ARM64 optimization: $GOARM64"
    ;;
  esac

  # Build
  GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" CGO_ENABLED="$CGO_ENABLED" \
    go build \
    -buildmode="$buildmode" \
    -trimpath \
    -buildvcs=false \
    -tags="$BUILD_TAGS" \
    -ldflags="$ldflags" \
    -o "$output_binary" \
    "$main_pkg" || fail "Build failed"

  success "Built: $(basename "$output_binary")"
}

main() {
  printf "${CYAN}→${NC} Cross-platform build for $TARGET_OS/$TARGET_ARCH\n\n"

  # Check Go is available
  command -v go >/dev/null 2>&1 || fail "Go not found"

  # Create output directory
  mkdir -p "$BIN_DIR"

  # Execute build
  execute_build

  success "Build completed"
}

main "$@"
