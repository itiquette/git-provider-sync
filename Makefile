# SPDX-FileCopyrightText: Josef Andersson
#
# SPDX-License-Identifier: CC0-1.0

CC 		= go build
CFLAGS		= -trimpath
LDFLAGS		= all=-w -s -X main.version=$$(git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0' | tr -d '\n') -X main.commit=$$(git rev-parse HEAD) -X main.date=$$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GCFLAGS 	= all=
ASMFLAGS 	= all=
CGO_ENABLED 	= 0

# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./
DIR 		= ./dist
EXECUTABLE  	= gitprovidersync
SHELL := /bin/bash # Use bash syntax
GH_AUTH?=GitHub_auth_token_not_set

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ==================================================================================== #
# GENERATE
# ==================================================================================== #

## generate: generate mocks, manpages, completions
.PHONY: generate
generate: generate/manpage generate/completion generate/mock

## generate/manpage: generate manpage under ./generated/manpages
.PHONY: generate/manpage
generate/manpage:
	./scripts/manpage.sh

## generate/completion: generate completions under ./generated/completions
.PHONY: generate/completion
generate/completion:
	./scripts/completions.sh

## generate/mock: generate mocks under ./generated/mocks
.PHONY: generate/mock
generate/mock:
	./scripts/generatemock.sh

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## quality: run all quality control checks
.PHONY: quality
quality: clean quality/tidy quality/wsl quality/golangcilint quality/megalint quality/license quality/commit
	go mod verify
	go vet ./...
	staticcheck -checks=all,-ST1000,-U1000 ./...
	govulncheck ./...
	go test -count=1 -race -buildvcs=false -vet=off $$(go list './...' | grep -v generated)

## quality/tidy: format code and tidy modfile
.PHONY: quality/tidy
quality/tidy: clean
	go fmt ./...
	go mod tidy -v

MEGALINTER_DEF_WORKSPACE ?= /repo
## quality/megalint: quality control check with MegaLinter
.PHONY: quality/megalint
quality/megalint: clean
	podman run --rm --volume $$(pwd):/repo -e MEGALINTER_CONFIG='development/megalinter.yml' -e REPORT_OUTPUT_FOLDER=$(DIR)/megalinter-reports -e DEFAULT_WORKSPACE=${MEGALINTER_DEF_WORKSPACE} -e LOG_LEVEL=INFO ghcr.io/oxsecurity/megalinter-go:v8.0.0


## quality/golangcilint: quality control check with golangci-lint
.PHONY: quality/golangcilint
quality/golangcilint:
	golangci-lint run --fix

## quality/wsl: quality control check with wsl
.PHONY: quality/wsl
quality/wsl:
	wsl --fix ./...

## quality/license: license control check with reuse
.PHONY: quality/license
quality/license:
	podman run --rm --volume $$(pwd):/data docker.io/fsfe/reuse:4.0.3-debian lint

## quality/openssfscorecard: security repo control check with openssf scorecard
.PHONY: quality/openssfscorecard
quality/openssfscorecard:
	GITHUB_AUTH_TOKEN=${GH_AUTH} scorecard --repo=github.com/itiquette/git-provider-sync	

COMPARETOBRANCH ?= main
## quality/commit: commit format check 
.PHONY: quality/commit
quality/commit:
	@if [[ $$(git rev-list --count ${COMPARETOBRANCH}..) -gt 0 ]]; then \
		podman run --rm -i --volume $$(pwd):/repo -w /repo ghcr.io/siderolabs/conform:v0.1.0-alpha.30 enforce --base-branch=${COMPARETOBRANCH}; \
	else \
		echo "no new commits found in branch compared to ${COMPARETOBRANCH}, skipping commit lint"; \
	fi


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #


## build: run quality checks, tests and build local release packages (don't publish)
.PHONY: build
build: quality build/gorelease

## build/plain: build the application with go build
.PHONY: build/plain
build/plain: clean
	@export GOOS=linux; export GOARCH=amd64; export CGO_ENABLED=$(CGO_ENABLED); $(CC) $(CFLAGS) -o=$(DIR)/$(EXECUTABLE)-$${GOOS}-$${GOARCH} -ldflags="$(LDFLAGS)" -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" ${MAIN_PACKAGE_PATH};        
	@export GOOS=linux; export GOARCH=arm64; export CGO_ENABLED=$(CGO_ENABLED); $(CC) $(CFLAGS) -o=$(DIR)/$(EXECUTABLE)-$${GOOS}-$${GOARCH} -ldflags="$(LDFLAGS)" -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" ${MAIN_PACKAGE_PATH};        

## build/plainwithimage: build the application with go build and a linux amd/arm container
.PHONY: build/plainwithimage
build/plainwithimage: clean build/plain
	podman build -t git-provider-sync:dev --build-arg DIRPATH=$(DIR)/ --platform=linux/amd64,linux/arm64  -f Containerfile .

## build/drygorelease: build release packages with goreleaser (don't publish)
.PHONY: build/drygorelease
build/drygorelease: clean
	goreleaser check
	goreleaser release --clean --snapshot

## build/gorelease: build release packages with goreleaser (publish)
.PHONY: build/gorelease
build/gorelease: clean
	goreleaser check
	goreleaser release --clean

## test: run all tests
.PHONY: test
test:
	go test -v -count=1 -race -buildvcs=false $$(go list './...' | grep -v generated)

## test/coverage: run all tests and display coverage
.PHONY: test/coverage
test/coverage: clean
	go test -v -count=1 -race -buildvcs=false -coverprofile=$(DIR)/coverage.out $$(go list './...' | grep -v generated)
	go tool cover -html $(DIR)/coverage.out -o $(DIR)/coverage.html
	open $(DIR)/coverage.html

## upgrade: upgrade all dependencies
.PHONY: upgrade
upgrade:
	go get -u -t ./...
	go mod tidy

## upgrade/list: list all updated dependencies
.PHONY: upgrade/list
upgrade/list:
	go list -u -m all


# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

## install/pre-req: install pre-req projects for developing (golangci-lint, mockery, syft)
.PHONY: install/prereq
install/prereq:
	@./scripts/installdevprereq.sh

## clean/gocache: remove go and golangci-lint cache
.PHONY: clean/gocache
clean/gocache:
	go clean -cache --modcache --testcache --fuzzcache
	golangci-lint cache clean

## clean: remove local tmp dirs from previous runs 
.PHONY: clean
clean:
	rm -rf $(DIR)
	mkdir -p $(DIR)/megalinter-reports/sarif
