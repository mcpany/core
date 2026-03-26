# Design Doc: Deterministic Attestation Gateway
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
The persistent threat from project-local configuration vulnerabilities (CVE-2025-59536) requires a move toward "Deterministic Boot" for AI agents. Current sandboxing is often bypassed if an agent can influence its environment before it is fully bound. The Deterministic Attestation Gateway provides a "Pre-Execution Attestation" service that ensures the project environment is in a known-good state before any agent code runs.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate a signed "Full-State Manifest" of the project environment (files, directories, env vars).
    * Provide cryptographic "Non-Existence Proofs" for sensitive paths (e.g., `.claude/settings.json`).
    * Integrate with the Pre-Flight Sandbox Validator to enforce environment immutability.
    * Support "Deterministic Boot" compliance for high-security agent frameworks.
* **Non-Goals:**
    * Validating the runtime behavior of the agent (handled by other middlewares).
    * Managing the underlying sandbox technology (it works with Docker, bubblewrap, etc.).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Deployer.
* **Primary Goal:** Verify the complete integrity of the project environment before allowing an agent to boot in a high-security zone.
* **The Happy Path (Tasks):**
    1. Deployer runs `mcpany attest --project ./secure-repo`.
    2. MCP Any performs a recursive scan of the project directory.
    3. MCP Any generates a manifest where every existing file is hashed and every missing file is recorded as "Absent".
    4. The manifest is cryptographically signed by MCP Any's internal authority.
    5. The agent is launched with the `--attested-manifest` flag.
    6. At startup, the agent's loader verifies the current environment against the signed manifest.
    7. If any deviation is found (e.g., a new `.env` file appeared), the boot is aborted.

## 4. Design & Architecture
* **System Flow:**
    `Scan` -> `Hash State` -> `Sign Manifest` -> `Agent Load` -> `Verify Manifest`
* **APIs / Interfaces:**
    * `AttestationService`: `Attest(path string) (SignedManifest, error)`
    * `IntegrityChecker`: `Verify(manifest SignedManifest) error`
* **Data Storage/State:**
    * Signed manifest files (JSON or Protobuf) stored locally or in a secure vault.

## 5. Alternatives Considered
* **Static Config Checkers**: Rejected because they don't handle TOCTOU (Time-of-Check to Time-of-Use) races where files are injected during the boot sequence.
* **OS-level Snapshots (ZFS/Btrfs)**: Rejected due to platform-specific dependencies and the requirement for root privileges.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Eliminates the "Implicit Trust" of the local filesystem during agent initialization.
* **Observability**: All attestation events and failures are logged with high-resolution timestamps.

## 7. Evolutionary Changelog
* **2026-04-11:** Initial Document Creation.
* **2026-04-12:** Updated to address CVE-2026-25725. Added specific focus on "Non-Existence Proofs" for missing configuration files to prevent "Empty File" injection escapes. Introduced the "Settings Injection Guard" as a recommended integration point for validating project-local settings against attested baselines.

### Update: 2026-04-13 - Introducing Environment Integrity Manifests
**Context:** The shift toward "Deterministic Boot" requires an immutable record of the project state before agent execution.
**Architecture Adjustment:**
* Section 4 now includes the "Environment Integrity Manifest" as a signed artifact of the attestation process.
* The gateway will now distribute these manifests to agent runtimes as a mandatory boot prerequisite.
**Security Impact:** Prevents runtime environment modification and ensures that the agent initializes in a cryptographically verified state.

### Update: 2026-04-14 - TPM-Bound Configuration Hardening
**Context:** The persistence of CVE-2026-25725 style configuration escapes in "cloned repository" scenarios proves that software-only manifests are insufficient.
**Architecture Adjustment:**
* **Hardware-Locked Boot:** Section 4 now includes a requirement for TPM-bound signatures on all project-local hooks and security-critical settings.
* **Binding Logic:** The gateway will now refuse to attest to a manifest if sensitive hooks are not cryptographically bound to the local machine's TPM, neutralizing the risk of "Payload Smuggling" in malicious repo clones.
**Security Impact:** Moves the trust boundary from the software manifest to the hardware, ensuring that security policies cannot be bypassed even if the manifest file itself is compromised or replaced.
