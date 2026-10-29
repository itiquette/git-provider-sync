# SPDX-FileCopyrightText: 2025 itiquette/git-provider-sync
#
# SPDX-License-Identifier: CC0-1.0

FROM cgr.dev/chainguard/glibc-dynamic:latest-dev@sha256:6fb1a88c68a9898d70760d1670581aa70717426c332d09f5dc81116fe7bc8e99
ARG TARGETOS TARGETARCH
ARG DIRPATH=""

COPY ${DIRPATH}gitprovidersync-${TARGETOS}-${TARGETARCH} /usr/bin/gitprovidersync
ENTRYPOINT ["/usr/bin/gitprovidersync"]
