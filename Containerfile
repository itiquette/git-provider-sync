# SPDX-FileCopyrightText: Josef Andersson
#
# SPDX-License-Identifier: CC0-1.0

FROM cgr.dev/chainguard/glibc-dynamic:latest-dev@sha256:da2669610231b98dc8a064c8f3d9855605e95deaee42c8011842c37fd4ca6efd
ARG TARGETOS TARGETARCH
ARG DIRPATH=""

COPY ${DIRPATH}gitprovidersync-${TARGETOS}-${TARGETARCH} /usr/bin/gitprovidersync
ENTRYPOINT ["/usr/bin/gitprovidersync"]
