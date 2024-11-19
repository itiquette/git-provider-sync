#!/usr/bin/env bash
# SPDX-FileCopyrightText: Josef Andersson
#
# SPDX-License-Identifier: CC0-1.0

# Description: This script generates mock interfaces for the gitprovidersync project
#              using the mockery tool. It creates mocks for external dependencies
#              and internal interfaces.
#
# Usage: ./generate_mocks.sh
#
# The script performs the following actions:
# 1. Generates mocks for the go-git-providers library
# 2. Generates mocks for internal interfaces defined in the project
# 3. Generates mocks for the go-gitlab library
#
# Mock files are generated in the following directory structure:
# - generated/mocks/mockgitprovider: Mocks for go-git-providers
# - generated/mocks/mockgogit: Mocks for internal interfaces
# - generated/mocks/mockgitlab: Mocks for go-gitlab
#
# Dependencies:
# - mockery: This script requires the mockery tool to be installed and available in the PATH
# - Go: The project and its dependencies should be properly set up
#
# Note: Ensure you run this script from the root directory of the gitprovidersync project
#
set -euo pipefail

# Define color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Define common mockery options
MOCKERY_CMD="mockery --with-expecter"
OUTPUT_DIR="generated/mocks"

# Function to run mockery command with error handling
run_mockery() {
  if ! $MOCKERY_CMD "$@"; then
    echo -e "${RED}Error generating mock: $*${NC}" >&2
    return 1
  fi
}

echo -e "${BLUE}Starting mock generation...${NC}"

# Generate mocks for internal interfaces
echo -e "${YELLOW}Generating mocks for internal interfaces...${NC}"
INTERNAL_INTERFACES=(
  "GitRepository"
  "GitRemote"
  "SourceReader"
  "TargetWriter"
  "GitProvider"
  "GitInterface"
  "FilterServicer"
  "ProjectServicer"
  "ProtectionServicer"
)
for interface in "${INTERNAL_INTERFACES[@]}"; do
  echo -e "${BLUE}Generating mock for ${interface}...${NC}"
  run_mockery --dir=internal/interfaces --name="${interface}" --output "${OUTPUT_DIR}"/mockgogit
done

# Generate mocks for go-gitlab
echo -e "${YELLOW}Generating mocks for go-gitlab...${NC}"
run_mockery --all --srcpkg=github.com/xanzy/go-gitlab --output "${OUTPUT_DIR}"/mockgitlab

# Generate mocks for internal interfaces
echo -e "${YELLOW}Generating mocks for target interfaces...${NC}"
TARGET_INTERFACES=(
  "GitLibOperation"
  "CommandExecutor"
  "BranchManager"
)
for interface in "${TARGET_INTERFACES[@]}"; do
  echo -e "${BLUE}Generating mock for ${interface}...${NC}"
  run_mockery --dir=internal/target --name="${interface}" --output "${OUTPUT_DIR}"/mockgogit
done

echo -e "${GREEN}Mock generation completed successfully.${NC}"
