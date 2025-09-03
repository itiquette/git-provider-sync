#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Executes security audit pipeline combining OSSF scorecard and container scanning
# Usage: ./scripts/security-audit.sh [bin_dir] [executable_name]
# Dependencies: scorecard (optional), container-security-scan.sh, podman, dockle, trivy
# Output: Combined security assessment results, exit code 1 if issues found

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

# Runs OSSF scorecard security best practices scan if tool and auth available
run_ossf_scorecard() {
  printf "${CYAN}→${NC} Running OSSF scorecard (security best practices)...\n"

  # Check if scorecard is available
  if ! command -v scorecard >/dev/null 2>&1; then
    log "scorecard not found, skipping OSSF scorecard check"
    return 0
  fi

  # Check for GitHub authentication
  if [[ -z "${GH_AUTH:-}" ]] || [[ "${GH_AUTH}" == "GitHub_auth_token_not_set" ]]; then
    log "GH_AUTH environment variable not set, skipping scorecard"
    log "Scorecard requires GitHub authentication to avoid rate limits"
    log "To run scorecard: export GH_AUTH=your_github_token"
    return 0
  fi

  # Run scorecard check
  local scorecard_result=0
  scorecard --repo=. --format=table || scorecard_result=$?

  if [[ $scorecard_result -ne 0 ]]; then
    fail "OSSF scorecard check failed with exit code: $scorecard_result"
    log "Review scorecard output above for specific security issues"
    log "Visit https://securityscorecards.dev/ for remediation guidance"
    return 1
  fi

  success "OSSF scorecard check completed"
  return 0
}

# Runs container security scan via dedicated script
run_container_security_scan() {
  printf "${CYAN}→${NC} Running container vulnerability scan (dockle + trivy)...\n"

  # Check if container security scan script exists
  if [[ ! -f "./scripts/security/container-security-scan.sh" ]]; then
    fail "Container security scan script not found: ./scripts/security/container-security-scan.sh"
    return 1
  fi

  # Run container security scan
  if ! ./scripts/security/container-security-scan.sh "$BIN_DIR" "$EXECUTABLE"; then
    fail "Container security scan failed"
    log "Check container scan output above for vulnerability details"
    log "Ensure container scanning tools are properly installed"
    return 1
  fi

  return 0
}

main() {
  printf "${CYAN}→${NC} Starting security audit...\n"
  printf "\n"

  # Validate environment
  validate_project_directory || exit 1

  # Track scan results
  local audit_failed=false

  # Run OSSF scorecard check
  if ! run_ossf_scorecard; then
    audit_failed=true
  fi

  printf "\n"

  # Run container vulnerability scan
  if ! run_container_security_scan; then
    audit_failed=true
  fi

  printf "\n"

  # Final result
  if [[ "$audit_failed" == "true" ]]; then
    fail "Security audit completed with failures"
    log "Address security issues identified above before proceeding"
    log "Consider reviewing security documentation and best practices"
    exit 1
  fi

  success "Complete security audit passed"
}

main "$@"
