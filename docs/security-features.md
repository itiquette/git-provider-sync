<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Security Features

Build and release security for Git Provider Sync.

## SBOM (Software Bill of Materials)

File: `gitprovidersync_vX.Y.Z_linux_amd64.sbom.json`

Lists all components, libraries, and dependencies. Use SPDX or CycloneDX viewers to review.

## Checksums and Signatures

- `gitprovidersync_checksums_sha256.txt` - SHA256 checksums for all release files
- `gitprovidersync_checksums_sha256.txt.pem` - Public certificate
- `gitprovidersync_checksums_sha256.txt.sig` - Cosign signature

## SLSA Level 3 Attestation

File: `multiple.intoto.jsonl`

Build provenance with:
- Hermetic, scripted builds
- Version controlled source
- Generated provenance records

Verify with SLSA tools.

## Verification

### Basic

```bash
# Check SHA256
sha256sum -c gitprovidersync_checksums_sha256.txt
```

### Advanced

```bash
# Cosign verification
cosign verify-blob \
  --certificate gitprovidersync_checksums_sha256.txt.pem \
  --signature gitprovidersync_checksums_sha256.txt.sig \
  gitprovidersync_checksums_sha256.txt

# SLSA verification
slsa-verifier verify-artifact \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/itiquette/git-provider-sync \
  gitprovidersync_linux_amd64.tar.gz
```
