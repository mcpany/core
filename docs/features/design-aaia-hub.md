# Design Doc: Agent-to-Agent Identity Attestation (AAIA) Hub

**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The rise of "Lateral Swarm Social Engineering" has exposed a critical flaw in the Universal Agent Bus: implicit internal trust. Attackers are now compromising low-privilege specialist agents to impersonate parent agents or other high-trust teammates, initiating unauthorized actions (like fund transfers or code commits) that appear legitimate to the recipient.

MCP Any must solve this by becoming the authoritative **AAIA Hub**. This system will transition the inter-agent coordination bus from a "connected" state to a "Zero Trust" state, where every instruction is cryptographically signed and linked to a hardware-attested non-repudiable identity.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue session-bound, hardware-attested (TPM/SEP) identity fragments for every connected agent.
    * Mandate cryptographic signatures for all intra-swarm coordination messages (Mailboxes, Shared State).
    * Provide a real-time validation service that verifies the lineage and authority of an instruction before it is processed.
    * Implement "Identity Pinning" to prevent a compromised agent from rotating into a higher-privileged identity.
* **Non-Goals:**
    * Replacing existing framework-level coordination logic (e.g., Claude Code's task list).
    * Providing long-term persistent identity across different missions (AAIA is session-bound).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Auditor / Swarm Orchestrator
* **Primary Goal:** Prevent a compromised specialist agent (e.g., a "Code Reviewer" agent) from impersonating the "Deployment Lead" to trigger a production release.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a mission and requests AAIA identities for 3 subagents.
    2. AAIA Hub issues TPM-signed identity fragments to each subagent.
    3. Subagent A sends an instruction to Subagent B via the coordination bus, including its AAIA signature.
    4. AAIA Hub intercepts the message, validates the signature against the hardware root and mission manifest.
    5. Subagent B receives a "Validated" instruction and proceeds.
    6. (Threat Path) A compromised Subagent C attempts to send a "Deploy" command signed with Subagent A's ID; the AAIA Hub detects the signature mismatch/invalid lineage and quarantines the message.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent A->>AAIA Hub: Request Identity (TPM Attestation)
        AAIA Hub-->>Agent A: Identity Fragment (Session Token)
        Agent A->>Coordination Bus: Instruction + AAIA Signature
        Coordination Bus->>AAIA Hub: Validate Instruction
        AAIA Hub->>AAIA Hub: Verify Signature & Lineage
        AAIA Hub-->>Coordination Bus: Validation Result (Allow/Deny)
        Coordination Bus->>Agent B: Process instruction
    ```
* **APIs / Interfaces:**
    * `/v1/identity/issue`: Exchange TPM attestation for an AAIA session token.
    * `/v1/instruction/verify`: Validate a signed instruction fragment.
* **Data Storage/State:**
    * Identities are stored in a kernel-locked memory region, indexed by Session ID and Hardware ID.

## 5. Alternatives Considered
* **Framework-Native Auth:** Rejected because it creates "Identity Islands" and doesn't provide the hardware-bound guarantees required for Zero Trust in heterogeneous swarms.
* **Standard JWTs:** Rejected because they are susceptible to "Token Theft" and lateral reuse; AAIA requires hardware-bound signatures for every instruction.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** AAIA enforces the principle of least privilege at the instruction level, not just the session level.
* **Observability:** Every AAIA validation event is logged with full lineage, providing an immutable audit trail for inter-agent coordination.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
