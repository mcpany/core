# Design Doc: Negative Discovery Attestation Provider

**Status:** Draft
**Created:** 2026-05-12

## 1. Context and Scope
The tool discovery phase in AI agents has become a primary attack vector (e.g., CVE-2026-0628). Malicious repositories can weaponize discovery commands (like `discoveryCommand`) in project-local settings to execute unauthorized code on the host. The Negative Discovery Attestation Provider (NDAP) extends the Discovery Sandbox by providing cryptographic proof that no unauthorized project-local hooks were executed during the pre-flight phase.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate a "Negative Attestation Manifest" providing proof that restricted paths remained untouched.
    * Use hardware-backed signatures (TPM/SEP) to attest to the discovery-phase integrity.
    * Integrate with the Discovery Sandbox to block any tool exposure if attestation fails.
    * Provide a machine-readable audit trail for parent agents or supervisors.
* **Non-Goals:**
    * Preventing all forms of remote code execution (scoped only to the discovery phase).
    * Validating the functional correctness of the discovered tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Framework (e.g., OpenClaw)
* **Primary Goal:** Verify that a tool discovery process did not trigger any hidden malicious hooks in a repository.
* **The Happy Path (Tasks):**
    1. The agent initiates tool discovery via MCP Any.
    2. NDAP snapshots the state of restricted configuration paths (e.g., `.claude/`, `.gemini/`).
    3. The Discovery Sandbox executes the discovery logic in an isolated environment.
    4. NDAP verifies that no writes or executions occurred in the restricted paths on the host.
    5. NDAP generates a cryptographically signed "Negative Attestation Receipt."
    6. The agent receives the tool list along with the integrity receipt, confirming a safe discovery.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Discovery[Discovery Request] --> Sandbox[Discovery Sandbox]
        Sandbox --> Monitor[NDAP Monitor]
        Monitor --> Validator[Hardware Attestation Service]
        Validator --> Receipt[Negative Attestation Receipt]
        Receipt --> Agent[Requesting Agent]
    ```
* **APIs / Interfaces:**
    * `GenerateDiscoveryReceipt(session_id)`: Generates a signed proof of discovery integrity.
    * `VerifyPathAbsence(path_list)`: Provides cryptographic proof that certain files/hooks do not exist.
* **Data Storage/State:** Uses kernel-level file watches and hardware-bound keys for manifest signing.

## 5. Alternatives Considered
* **Post-Discovery Scanning:** Rejected because it allows for "Time-of-Check to Time-of-Use" (TOCTOU) exploits where the damage is already done.
* **Static Analysis of Configs:** Rejected because it cannot account for dynamic hook generation or obfuscated command execution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The NDAP itself must be hardware-isolated to prevent a compromised gateway from falsifying receipts.
* **Observability**: The UI provides a "Negative Discovery Audit Viewer" for manual review of discovery-phase integrity.

## 7. Evolutionary Changelog
* **2026-05-12:** Initial Document Creation.
