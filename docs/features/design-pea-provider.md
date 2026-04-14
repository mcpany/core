# Design Doc: Pre-Flight Environment Attestation (PEA) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of "Panel Hijacking" vulnerabilities (e.g., CVE-2026-0628 in Gemini) and the rise of malicious CLI configuration hooks confirm that simply sandboxing the agent process is no longer sufficient. Attackers are now targeting the *host environment* (browser extensions, shell aliases, CLI metadata) to exfiltrate secrets before the agent even begins to execute.

The Pre-Flight Environment Attestation (PEA) Provider is designed to provide a hardware-bound "Seal of Integrity" for the complete agent execution environment. It ensures that no unauthorized modifications exist in the browser or CLI path before sensitive capabilities are exposed to the reasoning engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform hardware-attested (TPM) scans of the browser and CLI environment before agent boot.
    * Detect and block unauthorized browser extensions or modified CLI metadata.
    * Generate a signed "Environment Manifest" that is required for high-risk tool execution.
    * Provide a standardized "PEA Signal" for the A2A Messaging Hub to verify remote node health.
* **Non-Goals:**
    * Providing anti-virus or malware removal for the host OS.
    * Managing user authentication (handled by LOWA).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Enterprise Agent User
* **Primary Goal:** Securely execute a filesystem-modifying agent without risk of hidden browser-based exfiltration.
* **The Happy Path (Tasks):**
    1. The user initiates a high-risk agent mission.
    2. The PEA Provider triggers a "Pre-Flight Scan" of the local environment.
    3. PEA verifies that the browser origin is locked and no blacklisted extensions are active.
    4. PEA checks the CLI path for unauthorized shim scripts or aliases.
    5. PEA generates a hardware-attested manifest and binds it to the mission session.
    6. The agent executes; every tool call is validated against this PEA manifest.
    7. If the environment is compromised mid-session, PEA revokes the attestation and the mission is forcefully terminated.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Boot] -> [PEA Provider (Scan)] -> [TPM Attestation] -> [Signed Manifest] -> [Mission Start]`
* **APIs / Interfaces:**
    * `pea.ScanEnvironment() -> (Manifest, error)`: Initiates the pre-flight scan.
    * `pea.ValidateSession(sessionID) -> bool`: Verifies that a session's environment remains untampered.
* **Data Storage/State:**
    * **Known-Good Baseline:** A TPM-signed registry of authorized environment states.
    * **Active Session Manifests:** Short-lived state binding sessions to specific hardware hashes.

## 5. Alternatives Considered
* **Reactive File Watches:** Rejected as they cannot detect extension-based hijacking which occurs within the browser memory space.
* **Purely Software-Based Fingerprinting:** Rejected as it can be easily spoofed by malicious hooks. PEA requires hardware-bound (TPM) root of trust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PEA is the primary prerequisite for the "Sovereign Discovery Proxy" (SDP). No tool is exposed unless PEA is verified.
* **Observability:** Real-time attestation status is visible in the "Environment Attestation Report" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
