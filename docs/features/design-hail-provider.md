# Design Doc: Hardware-Attested Intent Lineage (HAIL) Provider
**Status:** Draft
**Created:** 2026-04-06

## 1. Context and Scope
The "Delegation Gap" is a critical bottleneck in autonomous swarms: when a subagent needs to perform a high-stakes task (like using a TPM-locked key), it often fails because the hardware security module (HSM) cannot verify the subagent's authority. This leads to frequent human-in-the-loop interruptions and "approval fatigue."

The HAIL Provider solves this by creating a cryptographically signed "Chain of Command." Every subagent request is bundled with a hardware-attested proof of its lineage, proving it is a direct descendant of a user-authorized mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate recursive, hardware-signed tokens for subagent delegations.
    * Neutralize "Intent Ghosting" by verifying the entire chain back to the mission root.
    * Reduce human-in-the-loop overhead for verified sub-tasks.
* **Non-Goals:**
    * HAIL will NOT store the raw content of reasoning (handled by SRM).
    * It will NOT replace existing mTLS or transport security.

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Automation Agent
* **Primary Goal:** Allow a specialized "Deploy Subagent" to access a hardware-locked production secret without a manual prompt.
* **The Happy Path (Tasks):**
    1. The user initiates a "Deployment Mission" and provides a hardware-bound root signature.
    2. The Parent Agent spawns a "Deploy Subagent" and issues a HAIL sub-token.
    3. The Subagent requests the secret from MCP Any.
    4. The HAIL Provider verifies the hardware signature and the parentage chain.
    5. MCP Any releases the secret fragment to the subagent because the lineage is verified.

## 4. Design & Architecture
* **System Flow:**
    [Mission Root] -> [Parent Token] -> [Subagent HAIL Token] -> [Hardware Vault]
* **APIs / Interfaces:**
    * `POST /api/v1/hail/mint`: Issue a sub-token for a child agent.
    * `POST /api/v1/hail/verify`: Validate a lineage chain for a tool call.
* **Data Storage/State:**
    * Stateless verification using the user's TPM public key.

## 5. Alternatives Considered
* **JWT-only delegation**: Rejected because software-only tokens can be exfiltrated and replayed by rogue sub-processes.
* **Centralized Approval Service**: Rejected due to high latency and single point of failure in air-gapped environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory monotonic counters to prevent replay of old lineage chains.
* **Observability:** Visual "Lineage Tree" in the UI for auditing task delegations.

## 7. Evolutionary Changelog
* **2026-04-06:** Initial Document Creation.
