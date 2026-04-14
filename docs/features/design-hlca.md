# Design Doc: Hardware-Locked Configuration Anchor (HLCA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents move from ephemeral chat sessions to long-term project residents, the security of local configuration files (e.g., `.claude/settings.json`, `.mcpany/config.yaml`) becomes critical. Vulnerabilities like CVE-2026-33068 demonstrate that malicious repositories can commit "Settings-as-Code" to silently bypass security dialogs or redirect tool execution.

HLCA provides a mechanism to anchor project-local configurations to a hardware-attested user session. It ensures that any security-critical configuration change requires explicit, hardware-bound user re-attestation.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically bind project-local settings to a TPM/Secure Enclave-signed user session.
    * Detect unauthorized modifications to security-critical configuration fields.
    * Provide a user-approval flow for "Trusting" a new or modified project workspace.
* **Non-Goals:**
    * Encrypting the configuration files on disk (handled by OS-level encryption).
    * Managing non-security-related settings (e.g., UI theme preferences).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a cloned repository from silently overriding local security policies.
* **The Happy Path (Tasks):**
    1. The user clones a new repository and starts MCP Any.
    2. HLCA detects an un-anchored project-local configuration.
    3. The system prompts the user to "Anchor Workspace."
    4. The user approves, and HLCA generates a TPM-signed manifest of the current settings.
    5. A future `git pull` modifies a security hook in the config.
    6. HLCA detects the mismatch during the next tool call and blocks execution until re-attested.

## 4. Design & Architecture
* **System Flow:**
    [Config Load] -> [HLCA Validator] -> [TPM Signature Check] -> [Execution Flow]
* **APIs / Interfaces:**
    * `POST /v1/config/anchor`: Sign and anchor the current workspace configuration.
    * `GET /v1/config/status`: Returns the anchoring status and last attestation timestamp.
* **Data Storage/State:**
    * Signed manifests are stored in the local MCP Any security database, keyed by project root.

## 5. Alternatives Considered
* **File-System Permissions Only:** Rejected as they can be bypassed by users with legitimate write access who might be tricked by a malicious script.
* **Global Allow-Lists:** Rejected as they lack the project-local granularity required for complex agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The anchoring keys must never leave the local hardware enclave.
* **Observability:** Anchoring events and violation alerts are displayed in the `HLCA Configuration Anchor Manager`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.

### Update: 2026-07-25 - Transition to State-Bound Sovereignty
**Context:** The maturation of Claude Code WHP confirms that hardware anchoring is insufficient if the environment state is mutable post-attestation.
**Architecture Adjustment:**
* Mandating **State-Bound Identity (SBI)** for all high-privilege HLCA sessions.
* Integrating "Workspace-Hash Pinning" (WHP) into the configuration load sequence.
**Security Impact:** Neutralizes "Rug Pull" attacks by ensuring leases are invalidated upon any unauthorized workspace modification.
