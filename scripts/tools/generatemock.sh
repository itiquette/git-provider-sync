#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Generates mock interfaces for the hexagonal architecture using mockery
# Usage: ./scripts/tools/generatemock.sh
#
# The script generates mocks for:
# 1. Hexagonal architecture ports (domain interfaces)
# 2. External dependencies (GitLab API client, etc.)
#
# Mock files are generated in the following directory structure:
# - generated/mocks/mockhexagonal: Mocks for hexagonal ports
# - generated/mocks/mockgitlab: Mocks for external dependencies
#
# Dependencies:
# - mockery v2.x: This script requires mockery tool to be installed
# - Go: The project and its dependencies should be properly set up
# - .mockery.yaml: Configuration file for mockery
#
# Note: Ensure you run this script from the root directory of the project
#
set -euo pipefail

# Define color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Define directories
OUTPUT_DIR="generated/mocks"

# Check if mockery config file exists
check_mockery_config() {
  if [ ! -f ".mockery.yaml" ]; then
    printf '%bError: .mockery.yaml configuration file not found%b\n' "$RED" "$NC" >&2
    printf '%bPlease ensure the .mockery.yaml file exists in the project root%b\n' "$YELLOW" "$NC" >&2
    return 1
  fi
}

# Function to create output directories
prepare_directories() {
  printf '%bPreparing output directories...%b\n' "$BLUE" "$NC"
  mkdir -p "${OUTPUT_DIR}/mockhexagonal"
  mkdir -p "${OUTPUT_DIR}/mockgitlab"
  mkdir -p "${OUTPUT_DIR}/mockdomain"
}

# Function to clean old auto-generated mocks (preserve manual mocks)
clean_old_mocks() {
  printf '%bGenerating fresh mock files...%b\n' "$BLUE" "$NC"
}

printf '%bStarting hexagonal architecture mock generation...%b\n' "$BLUE" "$NC"

# Check prerequisites
check_mockery_config

# Prepare environment
prepare_directories
clean_old_mocks

# Check for existing manual mocks first
printf '%bChecking for existing manual mocks...%b\n' "$BLUE" "$NC"
if [ -f "${OUTPUT_DIR}/mockhexagonal/MockLogger.go" ] && [ -f "${OUTPUT_DIR}/mockhexagonal/MockRepositoryProvider.go" ]; then
  printf '%b✓ Manual hexagonal mocks found%b\n' "$GREEN" "$NC"
fi

# Try mockery generation (may fail due to complex imports)
printf '%bAttempting automatic mock generation...%b\n' "$YELLOW" "$NC"

# Generate mocks using mockery v2.53+ with command line approach
printf '%bGenerating Logger mock...%b\n' "$BLUE" "$NC"
if mockery --name Logger --dir internal/domain/ports --output "${OUTPUT_DIR}/mockhexagonal" --outpkg mockhexagonal --filename MockLogger.go --disable-version-string --with-expecter 2>/dev/null; then
  printf '%b✓ Logger mock generated automatically%b\n' "$GREEN" "$NC"
else
  printf '%b! Logger mock failed - will use manual approach%b\n' "$YELLOW" "$NC"
fi

printf '%bGenerating RepositoryProvider mock...%b\n' "$BLUE" "$NC"
if mockery --name RepositoryProvider --dir internal/domain/ports --output "${OUTPUT_DIR}/mockhexagonal" --outpkg mockhexagonal --filename MockRepositoryProvider.go --disable-version-string --with-expecter 2>/dev/null; then
  printf '%b✓ RepositoryProvider mock generated automatically%b\n' "$GREEN" "$NC"
else
  printf '%b! RepositoryProvider mock failed - will use manual approach%b\n' "$YELLOW" "$NC"
fi

printf '%bGenerating GitOperations mock...%b\n' "$BLUE" "$NC"
if mockery --name GitOperations --dir internal/domain/ports --output "${OUTPUT_DIR}/mockhexagonal" --outpkg mockhexagonal --filename MockGitOperations.go --disable-version-string --with-expecter 2>/dev/null; then
  printf '%b✓ GitOperations mock generated automatically%b\n' "$GREEN" "$NC"
else
  printf '%b! GitOperations mock failed - will use manual approach%b\n' "$YELLOW" "$NC"
fi

printf '%bGenerating Configuration mock...%b\n' "$BLUE" "$NC"
if mockery --name Configuration --dir internal/domain/ports --output "${OUTPUT_DIR}/mockhexagonal" --outpkg mockhexagonal --filename MockConfiguration.go --disable-version-string --with-expecter 2>/dev/null; then
  printf '%b✓ Configuration mock generated automatically%b\n' "$GREEN" "$NC"
else
  printf '%b! Configuration mock failed - will use manual approach%b\n' "$YELLOW" "$NC"
fi

# Generate external library mocks (specific interfaces only)
printf '%bGenerating GitLab API mocks for specific interfaces...%b\n' "$YELLOW" "$NC"
printf '%bTargeting only the interfaces we actually use%b\n' "$BLUE" "$NC"

# Download dependencies
go mod download >/dev/null 2>&1

# Generate specific interfaces for core git-provider-sync operations
# Selection based on hexagonal architecture ports - only mock what we actually use
interfaces_to_mock=(
  "ProjectsServiceInterface"          # Repository CRUD operations (create, list, delete)
  "GroupsServiceInterface"            # Organization/group repository listing
  "UsersServiceInterface"             # User repository access and permissions
  "BranchesServiceInterface"          # Branch operations and default branch management
  "ProtectedBranchesServiceInterface" # Branch protection rule management
)

success_count=0
for interface in "${interfaces_to_mock[@]}"; do
  printf '%bGenerating mock for %s...%b\n' "$BLUE" "$interface" "$NC"
  if timeout 30 mockery --name "${interface}" --srcpkg gitlab.com/gitlab-org/api/client-go --output "${OUTPUT_DIR}/mockgitlab" --filename "Mock${interface}.go" --disable-version-string --with-expecter 2>/dev/null; then
    printf '%b✓ %s mock generated%b\n' "$GREEN" "$interface" "$NC"
    success_count=$((success_count + 1))
  else
    printf '%b! %s mock failed (may not exist or have dependency issues)%b\n' "$YELLOW" "$interface" "$NC"
  fi
done

if [[ $success_count -gt 0 ]]; then
  printf '%b✓ Generated %d GitLab API mocks%b\n' "$GREEN" "$success_count" "$NC"
else
  printf '%b! No GitLab API mocks were generated%b\n' "$YELLOW" "$NC"
  printf '%bExisting mocks will be preserved%b\n' "$BLUE" "$NC"
fi

# Summary
printf '%bMock generation process completed!%b\n' "$GREEN" "$NC"
printf '%bGenerated files can be found in:%b\n' "$BLUE" "$NC"
printf "%s\n" "  - ${OUTPUT_DIR}/mockhexagonal/ (hexagonal architecture ports)"
printf "%s\n" "  - ${OUTPUT_DIR}/mockgitlab/ (external dependencies)"

# Check what was actually generated
if [ -d "${OUTPUT_DIR}" ]; then
  MOCK_COUNT=$(find "${OUTPUT_DIR}" -name "*.go" -type f | wc -l)
  printf '%bTotal mock files generated: %d%b\n' "$GREEN" "$MOCK_COUNT" "$NC"
else
  printf '%bNo mock files were generated%b\n' "$YELLOW" "$NC"
fi
