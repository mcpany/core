# Design Doc: Teammate Identity Attestation (TIA)
**Status:** Draft
**Created:** 2026-05-30

## 1. Context and Scope
As swarms move from hierarchical to horizontal (mesh) coordination, the risk of "Teammate Impersonation" has emerged. A compromised specialist agent can spoof headers in the shared mailbox to send unauthorized instructions to a more privileged sibling. MCP Any needs a hardware-attested identity layer that binds every inter-agent message to a verified mission role.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement per-message cryptographic signing for all teammate-to-teammate (T2T) communication.
    * Mandate hardware-attested (TPM/Secure Enclave) identity tokens for session-bound agency.
    * Provide automated "Identity Quarantine" for messages with failed attestation.
* **Non-Goals:**
    * Managing human user identities (Identity Provider/IdP role).
    * Providing long-lived agent identities (identities are mission-bound).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor
* **Primary Goal:** Prevent a "Refactor Agent" from being coerced by a compromised "Test Agent" into deleting production code.
* **The Happy Path (Tasks):**
    1. "Test Agent" attempts to send a `delete_file` command to the "Refactor Agent" via the shared mailbox.
    2. TIA Enforcer intercepts the message and requests a hardware-attested signature from the "Test Agent."
    3. The signature is validated against the mission root; TIA detects that the "Test Agent" role is not authorized for this instruction.
    4. TIA blocks the message, quarantines the "Test Agent," and alerts the auditor.

## 4. Design & Architecture
* **System Flow:**
    * Every T2T message must include a `x-mcp-tia-token`.
    * The token contains a TPM-signed hash of the message content and the agent's hardware-bound identity.
    * The T2T Encryption Bridge verifies the token before delivering the message to the target teammate's mailbox.
* **APIs / Interfaces:**
    * `tia.v1.SignMessage(message_body)`: Returns a hardware-attested signature.
    * `tia.v1.VerifyMessage(message_body, signature)`: Validates the lineage and authority of the sender.
* **Data Storage/State:**
    * Mission-bound public keys are stored in the secure TIA registry.
    * Attestation logs are pushed to the immutable PR integrity gate (APRIG).

## 5. Alternatives Considered
* **Shared Symmetric Keys:** Rejected because a single compromised agent would reveal the key for the entire swarm.
* **Software-only JWTs:** Rejected because they are susceptible to token theft and replay if the local environment is partially compromised.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** TIA is the foundation for horizontal Zero Trust; no message is trusted without hardware-bound proof of origin.
* **Observability:** Visualizes the "Chain of Command" in the UI, showing the verified lineage of every teammate instruction.

## 7. Evolutionary Changelog
* **2026-05-30:** Initial Document Creation.
