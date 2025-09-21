#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Performs container security scanning using dockle and trivy on built binaries
# Usage: ./scripts/security/container-security-scan.sh [bin_dir] [executable_name]
# Dependencies: podman, dockle (optional), trivy (optional), git
# Output: Security scan results, exit code 1 if critical vulnerabilities found

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly CYAN=$'\033[1;36m'
readonly NC=$'\033[0m' # No Color

readonly PROJECT_MARKERS=("go.mod" "justfile" "go.sum")
readonly BIN_DIR="${1:-./bin}"
readonly EXECUTABLE="${2:-$(basename "$(pwd)")}"

log() {
  printf "${YELLOW}▸${NC} %s\n" "$1"
}

success() {
  printf "${GREEN}✓${NC} %s\n" "$1"
}

fail() {
  printf "${RED}✗${NC} %s\n" "$1" >&2
}

# Validates that required project marker files exist in current directory
validate_project_directory() {
  local missing=()
  for marker in "${PROJECT_MARKERS[@]}"; do
    if [[ ! -f "$marker" ]]; then
      missing+=("$marker")
    fi
  done
  if ((${#missing[@]} > 0)); then
    fail "Missing project markers: ${missing[*]}"
    return 1
  fi
  return 0
}

validate_binaries() {
  if [[ ! -d "$BIN_DIR" ]]; then
    fail "Binary directory not found: $BIN_DIR"
    log "Build binaries first with: just build-binary"
    return 1
  fi

  local required_binaries=("$BIN_DIR/$EXECUTABLE-linux-amd64" "$BIN_DIR/$EXECUTABLE-linux-arm64")
  local missing=()

  for binary in "${required_binaries[@]}"; do
    if [[ ! -f "$binary" ]]; then
      missing+=("$(basename "$binary")")
    fi
  done

  if ((${#missing[@]} > 0)); then
    fail "Missing required binaries: ${missing[*]}"
    log "Build binaries first with: just build-binary"
    return 1
  fi

  return 0
}

validate_tools() {
  local missing_tools=()

  if ! command -v podman >/dev/null 2>&1; then
    missing_tools+=("podman")
  fi

  if ! command -v dockle >/dev/null 2>&1; then
    log "dockle not found - container best practices scan will be skipped"
  fi

  if ! command -v trivy >/dev/null 2>&1; then
    log "trivy not found - vulnerability scan will be skipped"
  fi

  if ((${#missing_tools[@]} > 0)); then
    fail "Missing required tools: ${missing_tools[*]}"
    log "Install missing tools to proceed"
    return 1
  fi

  return 0
}

# Builds multi-architecture container image using podman buildx
build_container_image() {
  local build_tag="$1"

  printf "${CYAN}→${NC} Building multi-arch container image: %s\n" "$build_tag"

  if ! podman buildx build \
    --no-cache \
    -t "$EXECUTABLE:$build_tag" \
    --build-arg DIRPATH="$BIN_DIR/" \
    --platform=linux/amd64,linux/arm64 \
    -f Containerfile .; then
    fail "Failed to build container image"
    return 1
  fi

  return 0
}

# Runs dockle container best practices scan if tool is available
run_dockle_scan() {
  local build_tag="$1"

  printf "${CYAN}→${NC} Running dockle (container best practices scan)...\n"

  if ! command -v dockle >/dev/null 2>&1; then
    log "dockle not found, skipping container best practices scan"
    return 0
  fi

  local temp_tar
  temp_tar=$(mktemp /tmp/container-dockle-scan-XXXXXX.tar)

  # Save image to tar for dockle analysis
  if ! podman save -o "$temp_tar" "localhost/$EXECUTABLE:$build_tag"; then
    fail "Failed to save container image for dockle analysis"
    return 1
  fi

  # Run dockle scan
  local scan_result=0
  dockle --input "$temp_tar" || scan_result=$?

  if [[ $scan_result -ne 0 ]]; then
    log "dockle scan completed with warnings/errors (exit code: $scan_result)"
  fi

  return 0
}

# Runs trivy vulnerability scan with critical severity check
run_trivy_scan() {
  local build_tag="$1"

  printf "${CYAN}→${NC} Running trivy (container vulnerability scan)...\n"

  if ! command -v trivy >/dev/null 2>&1; then
    log "trivy not found, skipping container vulnerability scanning"
    return 0
  fi

  # Run trivy scan with critical severity exit code
  if ! trivy image --exit-code 1 --severity CRITICAL "localhost/$EXECUTABLE:$build_tag"; then
    fail "trivy found critical vulnerabilities in container image"
    return 1
  fi

  return 0
}

# Removes temporary container image after scanning
cleanup_image() {
  local build_tag="$1"

  log "Cleaning up temporary image..."
  podman rmi "localhost/$EXECUTABLE:$build_tag" || true
}

main() {
  # Validate environment
  validate_project_directory || exit 1
  validate_binaries || exit 1
  validate_tools || exit 1

  # Generate unique tag for this specific build
  local build_tag
  build_tag="scan-$(git rev-parse --short HEAD)-$(date +%s)-$$"

  # Build container image
  if ! build_container_image "$build_tag"; then
    exit 1
  fi

  # Run security scans
  local scan_failed=false

  if ! run_dockle_scan "$build_tag"; then
    scan_failed=true
  fi

  if ! run_trivy_scan "$build_tag"; then
    scan_failed=true
  fi

  # Always cleanup, even if scans failed
  cleanup_image "$build_tag"

  if [[ "$scan_failed" == "true" ]]; then
    fail "Container security scans failed"
    log "Check scan output above for specific issues"
    log "Ensure dockle and trivy are installed and up-to-date"
    exit 1
  fi

  success "Container vulnerability scans completed"
}

main "$@"
