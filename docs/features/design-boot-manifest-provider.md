# Design Doc: Hardware-Attested Boot Manifest Provider
**Status:** Draft
**Created:** 2026-04-16

## 1. Context and Scope
With the shift toward "Deterministic Environment Integrity," agents require a verifiable proof that their execution environment (configurations, hooks, local files) is exactly as intended before they boot. The Hardware-Attested Boot Manifest Provider (HABMP) bridges the gap between the Pre-Flight Sandbox Validator and the local TPM. It generates the final, cryptographically bound "Golden Image" of the environment that the agent runtime must verify before execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Aggregate "Full-State Manifests" (including Inode-pins and Non-Existence proofs).
    * Coordinate with the `TPMSignerService` to bind the manifest to the local hardware identity.
    * Provide a standardized "Boot Token" that agent runtimes (like Claude Code or OpenClaw) can use to verify integrity.
    * Support "Resident Integrity Monitoring" by providing a baseline for periodic re-attestation.
* **Non-Goals:**
    * Providing the filesystem isolation (this is handled by the Sandbox Service).
    * Validating the *content* of the files (only their *integrity* and *binding*).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a swarm of 5 subagents is booting into a perfectly clean, hardware-locked environment that matches the parent's security posture.
* **The Happy Path (Tasks):**
    1. Parent Agent initiates a swarm boot.
    2. HABMP calls the Pre-Flight Validator to generate the environment manifest.
    3. HABMP retrieves the "Non-Existence Proofs" for sensitive configuration files.
    4. HABMP sends the aggregated manifest hash to the local TPM for signing.
    5. TPM returns a hardware-bound signature.
    6. HABMP bundles the signature and manifest into a "Hardware-Attested Boot Manifest."
    7. The agent runtimes verify the manifest and boot successfully.

## 4. Design & Architecture
* **System Flow:**
    `[Pre-Flight Validator] -> (Manifest) -> [HABMP] -> (Hardware Bind) -> [TPM] -> (Signature) -> [Boot Manifest]`
* **APIs / Interfaces:**
    * `ManifestProvider`: `GenerateBootManifest(sessionID string) (BootManifest, error)`
    * `AttestationEndpoint`: `VerifyManifest(manifest BootManifest) (bool, error)`
* **Data Storage/State:**
    * Temporary cache of active boot manifests in protected memory.

## 5. Alternatives Considered
* **Software-Only Manifests:** Rejected because they can be swapped by local processes with equal privileges.
* **Remote Attestation:** Rejected as a primary path to ensure offline-first reliability, though it may be supported as an extension.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The manifest must be bound to a unique `SessionID` and `HardwareNonce` to prevent reuse on different machines or sessions.
* **Observability:** Boot manifest generation and verification events are visualized in the "Deterministic Boot Dashboard."

## 7. Evolutionary Changelog
* **2026-04-16:** Initial Document Creation.
