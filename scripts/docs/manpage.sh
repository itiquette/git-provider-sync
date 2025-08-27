#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Project-agnostic man page generator
# Usage: ./scripts/docs/manpage.sh [binary_name] [source_doc] [output_dir]
# Defaults: binary_name=$(basename "$(pwd)"), source_doc=docs/$(basename "$(pwd)").1.md, output_dir=generated/manpages

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly NC=$'\033[0m' # No Color

# Parse arguments with defaults
readonly BINARY_NAME="${1:-${BINARY_NAME:-$(basename "$(pwd)")}}"
readonly SOURCE_DOC="${2:-${SOURCE_DOC:-docs/${BINARY_NAME}.1.md}}"
readonly OUTPUT_DIR="${3:-${OUTPUT_DIR:-generated/manpages}}"

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

check_dependencies() {
  local missing=()

  # Check for go-md2man
  if ! command -v go-md2man >/dev/null 2>&1; then
    missing+=("go-md2man")
  fi

  # Check for gzip
  if ! command -v gzip >/dev/null 2>&1; then
    missing+=("gzip")
  fi

  if ((${#missing[@]} > 0)); then
    fail "Missing dependencies: ${missing[*]}"
  fi
}

generate_manpage() {
  # Check if source exists
  if [[ ! -f "$SOURCE_DOC" ]]; then
    fail "Source document not found: $SOURCE_DOC"
  fi

  # Create output directory
  mkdir -p "$OUTPUT_DIR"

  # Generate man page
  local man_file="$OUTPUT_DIR/${BINARY_NAME}.1"
  local gz_file="${man_file}.gz"

  log "Generating man page from $SOURCE_DOC"

  # Generate man page
  go-md2man -in "$SOURCE_DOC" -out "$man_file" || fail "Failed to generate man page"

  # Compress man page
  log "Compressing man page"
  gzip -9 -f "$man_file" || fail "Failed to compress man page"

  # Verify output
  if [[ ! -f "$gz_file" ]]; then
    fail "Man page generation failed: $gz_file not created"
  fi

  local size
  size=$(stat -f%z "$gz_file" 2>/dev/null || stat -c%s "$gz_file" 2>/dev/null || echo "unknown")
  success "Generated man page: $gz_file (size: $size bytes)"
}

show_installation_help() {
  printf "\n${YELLOW}Man Page Installation:${NC}\n"
  printf "  System-wide: ${GREEN}sudo cp %s/%s.1.gz /usr/share/man/man1/${NC}\n" "$OUTPUT_DIR" "$BINARY_NAME"
  printf "  User-local:  ${GREEN}cp %s/%s.1.gz ~/.local/share/man/man1/${NC}\n" "$OUTPUT_DIR" "$BINARY_NAME"
  printf "  View with:   ${GREEN}man %s${NC}\n" "$BINARY_NAME"
}

main() {
  log "Starting man page generation for $BINARY_NAME"

  check_dependencies
  generate_manpage
  show_installation_help

  success "Man page generation completed"
}

main "$@"
