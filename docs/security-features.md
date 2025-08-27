<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Advanced Build and Release Security Features

This document provides detailed information about the advanced security features included in Git Provider Sync releases.

## Overview

Git Provider Sync includes several security features to ensure the integrity and authenticity of your download.

## Security Features

### 1. SBOM (Software Bill of Materials)

**What it is**: An SBOM is like a detailed ingredient list for software. It lists all components, libraries, and dependencies used in Git Provider Sync.

**File format**: `gitprovidersync_vX.Y.Z_linux_amd64.sbom.json`

**Why it's important**:  
- Transparency: You can see exactly what's in the software
- Security: Helps identify if any components have known vulnerabilities
- Compliance: Useful for organizations that need to track software components

**How to use it**: You can review the SBOM using tools like SPDX or CycloneDX viewers. This is especially useful for security teams or curious users who want to know more about the software's composition.

### 2. Checksums and Signatures

**Checksums file**: `gitprovidersync_checksums_sha256.txt`

- Contains SHA256 checksums for all release files
- Use it to verify the integrity of your download

**Public certificate**: `gitprovidersync_checksums_sha256.txt.pem`

- Used to verify the signature of the checksums file

**Signature file**: `gitprovidersync_checksums_sha256.txt.sig`

- Used with Cosign for advanced verification

**Why they're important**:  
- Integrity: Ensures your download hasn't been tampered with
- Authenticity: Confirms the software comes from the legitimate source

**How to use them**:
1. Calculate the SHA256 checksum of your downloaded file
2. Compare it with the corresponding checksum in the `.txt` file
3. For advanced users: Use Cosign to verify the signature, ensuring the checksums file itself is authentic

### 3. SLSA (Supply-chain Levels for Software Artifacts) Level 3 Attestation

**What it is**: SLSA is a security framework to prevent tampering, improve integrity, and secure packages and infrastructure in your projects, businesses or enterprises.

**File**: `multiple.intoto.jsonl`

**Why it's important**:
- Build Integrity: Ensures the software was built in a secure and controlled environment
- Tamper Protection: Makes it extremely difficult for attackers to insert malicious code during the build process
- Traceability: Provides a verifiable record of how, when, and where the software was built

**What Level 3 means**:  
- The build process is fully scripted/automated and hermetic
- The source is version controlled and checked for reviews
- The build generates provenance explaining how the artifact was created

**How to use it**:  
- Advanced users can use SLSA verification tools to check the provenance and ensure it meets Level 3 requirements
- This is particularly important for enterprise environments or security-conscious users

## Understanding the Security Benefits

While these security features might seem complex, they work together to provide several layers of protection:

1. **Transparency**: The SBOM lets you see what's in the software
2. **Integrity**: Checksums ensure your download matches the original file
3. **Authenticity**: Signatures prove the software comes from the legitimate source
4. **Build Security**: SLSA attestation confirms the software was built securely

For most users, simply checking the checksum is a good start. For those with higher security requirements, utilizing all these features provides a comprehensive security approach.

## Verification Methods

### Basic Verification
1. Download the `gitprovidersync_checksums_sha256.txt` file
2. Use a checksum tool to calculate the SHA256 hash of your downloaded Git Provider Sync file
3. Compare your calculated hash with the one in the checksums file. They should match exactly

### Advanced Verification
1. Use Cosign to verify the signature of the checksums file
2. Use SLSA verification tools to check the build provenance
3. Review the SBOM to understand all components included in the software

Remember: You don't need to understand or use all these features to safely install and use Git Provider Sync.
