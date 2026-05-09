# Design Doc: Hardware-Enclave Path Attestation (HEPA) Provider

**Status:** Draft
**Created:** 2026-05-06

## 1. Context and Scope
The "ClawJacked" (CVE-2026-25253) exploit and recent "Symlink-to-Inode Racing" (SIR) vulnerabilities have shown that path-based validation is no longer sufficient for securing local AI agent environments. Attackers can swap configuration files between the time of validation and the time of execution (TOCTOU). HEPA (Hardware-Enclave Path Attestation) addresses this by utilizing Secure Enclaves (TPM/SEP) to provide hardware-bound path validation and Inode pinning from the moment of the initial file open.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound (TPM/SEP) validation for project-local configuration paths.
    * Enforce "Inode Pinning" to prevent TOCTOU symlink races.
    * Provide "Deterministic Absence Proofs" (DAP) for restricted configuration files.
    * Integrate with the "Trust Lease Broker" for low-latency re-attestation.
* **Non-Goals:**
    * Implementing a full hardware security module (HSM) in software.
    * Replacing OS-level filesystem permissions.
    * Managing remote cloud-based secrets.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a malicious repository from hijacking the local AI agent via project-local configuration hooks.
* **The Happy Path (Tasks):**
    1. The user opens a new repository in their IDE.
    2. MCP Any's Pre-Flight Validator triggers the HEPA Provider.
    3. HEPA retrieves the hardware Inode for `.claude/settings.json` and pins it using the Secure Enclave.
    4. HEPA generates a signed "Environment Manifest" including the pinned Inode and DAPs for missing sensitive files.
    5. The AI agent boots, and its configuration loader verifies the HEPA manifest against the hardware root of trust.
    6. Any attempt to swap the `.claude/settings.json` file via symlink-racing is automatically detected and blocked at the hardware level.

## 4. Design & Architecture
* **System Flow:**
    * The HEPA Provider sits between the agent and the OS filesystem.
    * It utilizes `O_PATH` file handles and `fstat` to retrieve Inodes before any data is read.
    * Inodes are registered with the Secure Enclave, which issues a session-bound "Hardware Pin."
* **APIs / Interfaces:**
    * `AttestPath(path)`: Returns a hardware-signed manifest for the given path.
    * `VerifyPin(file_descriptor)`: Validates that the current FD matches the hardware-pinned Inode.
    * `GenerateDAP(path)`: Provides a signed proof that a specific path does not exist.
* **Data Storage/State:**
    * Transient, hardware-bound state maintained within the Secure Enclave and a session-bound "Pin Registry."

## 5. Alternatives Considered
* **Polling-Based File Watchers**: Rejected due to race conditions (TOCTOU) and high CPU overhead.
* **Standard SHA-256 Hashing**: Rejected because it does not prevent file-swapping between the hash check and the actual file read.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: HEPA provides a hardware-level guarantee of environment integrity, neutralizing configuration-injection escapes.
* **Observability**: The HEPA Security Monitor in the UI will visualize pinned Inodes and DAP status for the active workspace.

## 7. Evolutionary Changelog
* **2026-05-06:** Initial Document Creation.
