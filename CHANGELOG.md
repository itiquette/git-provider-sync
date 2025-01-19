# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.20] - 2025-01-19

### Added

- Add changelog flow

### Changed

- Tight up permissions
- Improve protect interface
- Improver provider interface
- Improve pullopt
- Bump min go-git, archive
- Refactor name
- Update github/codeql-action action to v3.28.0
- Bump deps
- Restructure config format
- Update actions/upload-artifact action to v4.5.0
- Update cgr.dev/chainguard/glibc-dynamic:latest-dev
- Improve configuration count text
- Update anchore/sbom-action action to v0.17.9
- Update actions/setup-go action to v5.2.0
- Update github/codeql-action action to v3.27.9
- Update cgr.dev/chainguard/glibc-dynamic:latest-dev docker digest to e7f1cd3
- Update go-git and misc

### Fixed

- Repair git binary function
- Set bare repos as default
- Repair flag ignore invalid name

## [0.0.19] - 2024-12-08

### Added

- Add gitbinary tests

### Changed

- Rename cleanup-name flag to ascii-name
- Rename field configurations to gitprovidersync
- Update github/codeql-action action to v3.27.6
- Update golang 1.23.4 and deps
- Improve tests
- Bump archive
- Update github/codeql-action action to v3.27.5
- Bump go-github,testify,chainguard
- Improve structure, tests

### Fixed

- Replace branch and tag protections with ruleset

## [0.0.18] - 2024-11-25

### Changed

- Improve structure, tests
- Improve test coverage, add badge
- Improve test coverage

## [0.0.17] - 2024-11-19

### Added

- Add disable functionality
- Add disable and enable protect

### Changed

- Update go-gitlab,ml
- Update step-security/harden-runner action to v2.10.2
- Clean up apiclients
- Update actions gorelease,misclint

### Fixed

- Improve client tests

## [0.0.16] - 2024-11-07

### Added

- Add project configurable def. visibility,ci
- Add tests, modularize gitlib,bin

### Changed

- Update chainguard/glibc-dynamic:latest-dev
- Update go 1.23.3, misc bumps
- Update anchore/sbom-action action to v0.17.7 (#103)

## [0.0.15] - 2024-11-03

### Added

- Add section about logging
- Add auth to fetch op
- Add default domain
- Add paging support to gh,gl

### Changed

- Improve general logging and traceing
- Move description to projectinfo
- Rename metainfo to projectinfo
- Improve iteration
- Update anchore/sbom-action action to v0.17.6
- Update actions/dependency-review-action action to v4.4.0

### Fixed

- Improve validation with docs

## [0.0.14] - 2024-10-29

### Fixed

- Rename proxycommand, improve docs,valid

## [0.0.13] - 2024-10-28

### Added

- Add initial architecture description
- Add syncrun per target options

### Changed

- Refactor sync
- Update actions/dependency-review-action action to v4.3.5
- Bump deps
- Update actions/setup-go action to v5.1.0 (#86)
- Update actions/checkout action to v4.2.2
- Update anchore/sbom-action action to v0.17.5

## [0.0.12] - 2024-10-22

### Changed

- Update github/codeql-action action to v3.27.0 (#83)

### Fixed

- Use ns-id for groups

## [0.0.11] - 2024-10-21

### Added

- Add custom ssh command, git binary support

### Changed

- Change slsa verifier action to use tag
- Update slsa-framework/slsa-verifier digest to 70f3c9a
- Pin cgr.dev/chainguard/glibc-dynamic docker tag to da26696
- Update anchore/sbom-action action to v0.17.4
- Update ghcr.io/siderolabs/conform docker digest to e824f01
- Update github/codeql-action action to v3.26.13
- Update github/codeql-action digest to 4dd1613 (#69)
- Bump github,gitlab libs

## [0.0.10] - 2024-10-14

### Changed

- Migrate renovate config (#66)
- Update actions/dependency-review-action digest to 3e334b7 (#65)
- Replace dependatbot with renovate
- Improve visibility mapping

### Removed

- Remove invalid conf line

## [0.0.9] - 2024-10-14

### Added

- Add openssf badge best prac

### Changed

- Update ml to 8.1.0
- Update go 1.23.2 and misc friends
- Use in-memory repo

### Fixed

- Set default branch after push
- Bump the actions-dependencies group across 1 directory with 5 updates

## [0.0.8] - 2024-09-22

### Added

- Add general configurable description
- Add certificate dir support, prop restruct

### Changed

- Restruct the configuration

### Fixed

- Update deps and docs

## [0.0.7] - 2024-09-18

### Added

- Add proxy support

## [0.0.6] - 2024-09-18

### Added

- Add fork config option

### Fixed

- Fix workaround wf permission
- Restore http support, initial gitea

## [0.0.4] - 2024-09-17

### Fixed

- Restore custom domain and https choice

### Removed

- Remove go-git-provider

## [0.0.2] - 2024-09-15

### Changed

- Initial support for ssh auth, refactorings

### Fixed

- Bump golang 1.23.1
- Bump github/codeql-action

## [0.0.1] - 2024-09-11

### Changed

- Move roadmap to issues
- Initial commit

[0.0.20]: https://github.com/itiquette/git-provider-sync/compare/v0.0.19..v0.0.20
[0.0.19]: https://github.com/itiquette/git-provider-sync/compare/v0.0.18..v0.0.19
[0.0.18]: https://github.com/itiquette/git-provider-sync/compare/v0.0.17..v0.0.18
[0.0.17]: https://github.com/itiquette/git-provider-sync/compare/v0.0.16..v0.0.17
[0.0.16]: https://github.com/itiquette/git-provider-sync/compare/v0.0.15..v0.0.16
[0.0.15]: https://github.com/itiquette/git-provider-sync/compare/v0.0.14..v0.0.15
[0.0.14]: https://github.com/itiquette/git-provider-sync/compare/v0.0.13..v0.0.14
[0.0.13]: https://github.com/itiquette/git-provider-sync/compare/v0.0.12..v0.0.13
[0.0.12]: https://github.com/itiquette/git-provider-sync/compare/v0.0.11..v0.0.12
[0.0.11]: https://github.com/itiquette/git-provider-sync/compare/v0.0.10..v0.0.11
[0.0.10]: https://github.com/itiquette/git-provider-sync/compare/v0.0.9..v0.0.10
[0.0.9]: https://github.com/itiquette/git-provider-sync/compare/v0.0.8..v0.0.9
[0.0.8]: https://github.com/itiquette/git-provider-sync/compare/v0.0.7..v0.0.8
[0.0.7]: https://github.com/itiquette/git-provider-sync/compare/v0.0.6..v0.0.7
[0.0.6]: https://github.com/itiquette/git-provider-sync/compare/v0.0.4..v0.0.6
[0.0.4]: https://github.com/itiquette/git-provider-sync/compare/v0.0.2..v0.0.4
[0.0.2]: https://github.com/itiquette/git-provider-sync/compare/v0.0.1..v0.0.2

<!-- generated by git-cliff -->
