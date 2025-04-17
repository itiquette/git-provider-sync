# SPDX-FileCopyrightText: 2025 itiquette/git-provider-sync
#
# SPDX-License-Identifier: CC0-1.0

FROM cgr.dev/chainguard/glibc-dynamic:latest-dev@sha256:2050597a0c525d863fee3b0eebce6d0a9b8c84ad4bc417fe1c4a11ad92e3d033
ARG TARGETOS TARGETARCH
ARG DIRPATH=""

COPY ${DIRPATH}gitprovidersync-${TARGETOS}-${TARGETARCH} /usr/bin/gitprovidersync
ENTRYPOINT ["/usr/bin/gitprovidersync"]
