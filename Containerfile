# SPDX-FileCopyrightText: 2025 itiquette/git-provider-sync
#
# SPDX-License-Identifier: CC0-1.0

FROM cgr.dev/chainguard/glibc-dynamic:latest-dev@sha256:b6b6ce613fd711146853b470ca61151898219647b40aa286be18d19cab6722d1
ARG TARGETOS TARGETARCH
ARG DIRPATH=""

COPY ${DIRPATH}gitprovidersync-${TARGETOS}-${TARGETARCH} /usr/bin/gitprovidersync
ENTRYPOINT ["/usr/bin/gitprovidersync"]
