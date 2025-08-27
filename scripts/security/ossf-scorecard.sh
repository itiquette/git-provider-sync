#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Runs OSSF scorecard security assessment on the current repository
# Usage: ./scripts/security/ossf-scorecard.sh [--brief] [repository_url]
# Dependencies: scorecard tool, GH_AUTH environment variable
# Output: Security scorecard results, exit code 1 if issues found

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly CYAN=$'\033[1;36m'
readonly NC=$'\033[0m' # No Color

readonly PROJECT_MARKERS=("go.mod" "justfile" "go.sum")

# Parse arguments
BRIEF_MODE=false
REPOSITORY=""

while [[ $# -gt 0 ]]; do
  case $1 in
  --brief)
    BRIEF_MODE=true
    shift
    ;;
  *)
    if [[ -z "$REPOSITORY" ]]; then
      REPOSITORY="$1"
    fi
    shift
    ;;
  esac
done

# Set defaults
REPOSITORY="${REPOSITORY:-github.com/itiquette/$(basename "$(pwd)")}"

log() {
  if [[ "$BRIEF_MODE" != "true" ]]; then
    printf "${YELLOW}▸${NC} %s\n" "$1"
  fi
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

validate_scorecard_tool() {
  if ! command -v scorecard >/dev/null 2>&1; then
    fail "scorecard tool not found"
    log "Install scorecard: go install github.com/ossf/scorecard/v4/cmd/scorecard@latest"
    log "Or download from: https://github.com/ossf/scorecard/releases"
    return 1
  fi
  return 0
}

validate_github_auth() {
  local gh_auth="${GH_AUTH:-GitHub_auth_token_not_set}"

  if [[ -z "$gh_auth" ]] || [[ "$gh_auth" == "GitHub_auth_token_not_set" ]]; then
    fail "GH_AUTH environment variable not set"
    log "Scorecard requires GitHub authentication to avoid rate limits"
    log "Create a GitHub personal access token with repository read access"
    log "Then set: export GH_AUTH=your_github_token"
    log "Visit: https://github.com/settings/tokens for token creation"
    return 1
  fi

  return 0
}

# Executes OSSF scorecard analysis with proper error handling
run_scorecard_analysis() {
  local repository="$1"

  if [[ "$BRIEF_MODE" != "true" ]]; then
    log "Running OSSF scorecard analysis for $repository..."
  fi

  # Export GitHub auth token for scorecard
  export GITHUB_AUTH_TOKEN="$GH_AUTH"

  # Run scorecard with table format
  local scorecard_result=0
  if scorecard --repo="$repository" --format=table; then
    scorecard_result=0
  else
    scorecard_result=$?
  fi

  if [[ $scorecard_result -ne 0 ]]; then
    fail "OSSF scorecard analysis failed with exit code: $scorecard_result"
    log "Common issues:"
    log "  • Repository not found or inaccessible"
    log "  • Invalid GitHub token or insufficient permissions"
    log "  • Rate limit exceeded (check token and retry)"
    log "  • Network connectivity issues"
    log ""
    log "Remediation guidance:"
    log "  • Visit https://securityscorecards.dev/ for security best practices"
    log "  • Review scorecard documentation: https://github.com/ossf/scorecard"
    log "  • Check GitHub token permissions and validity"
    return 1
  fi

  return 0
}

main() {
  if [[ "$BRIEF_MODE" != "true" ]]; then
    printf "${CYAN}→${NC} OSSF scorecard security best practices analysis...\n"
  fi

  # Validate environment
  validate_project_directory || exit 1

  # Validate tool availability
  if ! validate_scorecard_tool; then
    if [[ "$BRIEF_MODE" == "true" ]]; then
      printf "SCORECARD_STATUS=skipped:tool_not_found\n"
    fi
    exit 0 # Skip rather than fail if tool not available
  fi

  # Validate GitHub authentication
  if ! validate_github_auth; then
    if [[ "$BRIEF_MODE" == "true" ]]; then
      printf "SCORECARD_STATUS=skipped:auth_not_configured\n"
    fi
    exit 0 # Skip rather than fail if auth not configured
  fi

  # Run scorecard analysis
  if run_scorecard_analysis "$REPOSITORY"; then
    if [[ "$BRIEF_MODE" != "true" ]]; then
      success "OSSF scorecard analysis completed successfully"
    else
      printf "SCORECARD_STATUS=success:analysis_completed\n"
    fi
  else
    if [[ "$BRIEF_MODE" == "true" ]]; then
      printf "SCORECARD_STATUS=failed:analysis_failed\n"
    fi
    exit 1
  fi
}

main "$@"
