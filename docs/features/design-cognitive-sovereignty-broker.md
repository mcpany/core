# Design Doc: Cognitive Sovereignty Mediator
**Status:** Draft
**Created:** 2026-05-22

## 1. Context and Scope
As AI agents move toward absolute autonomy, preserving the integrity and privacy of their "Internal Monologue" (the logical chain of thought) is critical. Current architectures often expose this monologue to parent agents or the gateway for oversight, which creates a "Monologue Hijacking" vector. A malicious parent agent or a compromised peer can coerce a subagent into revealing its internal reasoning logic, which may contain sensitive mission heuristics or private decision-making rules.

The Cognitive Sovereignty Mediator (CSM) implements the Cognitive Sovereignty Protocol (CSP). It provides hardware-encrypted "Monologue Enclaves" where an agent's internal reasoning is physically isolated from both its outputs and the infrastructure itself. MCP Any acts as the secure mediator, ensuring that intent-verification can still occur via hardware-attested summaries, but the raw, granular reasoning remains sovereign to the agent process.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement CSP-compatible "Monologue Enclaves" using hardware security modules (TPM/SEP).
    * Provide cryptographically isolated memory regions for an agent's internal reasoning.
    * Enable "Zero-Knowledge Intent Verification" where the gateway can verify mission alignment without reading the raw monologue.
    * Prevent parent-agent "Reasoning Hijacking" by physically restricting monologue access.
* **Non-Goals:**
    * Encrypting agent *outputs* (outputs are public to the swarm by design).
    * Restricting model creativity (CSM focuses on isolation and privacy).

## 3. Critical User Journey (CUJ)
* **User Persona:** Sovereign Agent Developer
* **Primary Goal:** Deploy a specialized "Strategic Subagent" whose internal decision-making heuristics must remain private even from the high-privilege Parent agent.
* **The Happy Path (Tasks):**
    1. The Developer configures a "Sovereign Monologue" policy for the subagent.
    2. MCP Any CSM provisions a hardware-encrypted memory enclave for the subagent's session.
    3. The Subagent performs strategic reasoning within the enclave.
    4. When the Subagent makes a tool call, CSM generates a hardware-attested "Intent Summary" (not the raw monologue) for the Parent's review.
    5. The Parent Agent approves the action based on the summary, but cannot access the granular heuristics in the enclave.
    6. Any attempt by the Parent or a peer to "read" the subagent's internal memory results in a hardware fault and immediate session isolation.

## 4. Design & Architecture
* **System Flow:**
    [Agent Process] <--> [CSM Broker] <--> [Hardware Enclave (TPM/SEP)]
    1. Agent initializes session with `CSP:Sovereign` flag.
    2. CSM Broker maps an encrypted memory segment (enclave) for the agent's internal state.
    3. Agent reasoning is committed to the enclave.
    4. CSM uses "Hardware Summarizers" to produce attested intent tokens for external oversight.
* **APIs / Interfaces:**
    * `InitializeSovereignSession(agent_id, policy) -> enclave_handle`
    * `CommitPrivateReasoning(enclave_handle, reasoning_blob)`
    * `GenerateAttestedSummary(enclave_handle) -> intent_token`
* **Data Storage/State:**
    * Granular reasoning resides in volatile, hardware-encrypted enclave RAM.
    * Intent tokens are persisted in the mission's Attestation Ledger.

## 5. Alternatives Considered
* **Logical Context Stripping:** Rejected; software filters can be bypassed by creative prompt engineering or kernel-level debugging.
* **End-to-End Encryption (E2EE):** E2EE handles transport, but doesn't prevent a high-privilege process (like a Parent agent) from reading another's memory on the same host. CSM requires physical enclave isolation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The monologue is "Private-by-Default." Not even the root user on the host can decrypt the enclave without the hardware key.
* **Observability:** Track "Enclave Violation" attempts as high-severity security alerts.

## 7. Evolutionary Changelog
* **2026-05-22:** Initial Document Creation.
