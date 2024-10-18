# SPDX-FileCopyrightText: Josef Andersson
#
# SPDX-License-Identifier: CC0-1.0

FROM cgr.dev/chainguard/glibc-dynamic:latest-dev@sha256:7c9315864a85195116d712cc163a0ec5bc048b9979dd143aeb9bada01d7a9639
ARG TARGETOS TARGETARCH
ARG DIRPATH=""

COPY ${DIRPATH}gitprovidersync-${TARGETOS}-${TARGETARCH} /usr/bin/gitprovidersync
ENTRYPOINT ["/usr/bin/gitprovidersync"]
