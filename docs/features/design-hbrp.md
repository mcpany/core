# Design Doc: Hardware-Bound Reasoning Provenance (HBRP) Provider
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
In deep agent swarms (e.g., A -> B -> C delegations), the "Identity" of the root mission can become "diluted." Each subsequent delegation hop increases the risk that the specialist agent loses its connection to the primary user objective, or that a compromised subagent in the chain can "shadow" the mission root. This leads to "Intent Drift" and makes auditing the lineage of high-risk tool calls nearly impossible in heterogeneous meshes.

The Hardware-Bound Reasoning Provenance (HBRP) Provider is a security service designed to mandate hardware-attested lineage signatures for all multi-hop agent delegations. Every reasoning step in the chain carries a signature that includes a sliding-window lineage of the previous three hops, ensuring absolute mission-root sovereignty throughout the delegation path.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-attested reasoning provenance for multi-hop delegations.
    * Mandate "Sliding-Window Lineage" signatures for all agent reasoning fragments.
    * Provide a non-repudiable "Chain of Reason" for all high-risk tool calls.
    * Neutralize "Identity Dilution" in deep agent swarms.
* **Non-Goals:**
    * Storing the full history of all agent reasoning traces.
    * Replacing the structural validation provided by the SRM Provider.
    * Managing the transport of the provenance tokens themselves (handled by UAB/A2A).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor
* **Primary Goal:** Verify the complete, hardware-attested lineage of a high-risk tool call (e.g., `run_shell_command`) performed by a 4th-level subagent.
* **The Happy Path (Tasks):**
    1. Parent Agent A delegates a task to Subagent B.
    2. Subagent B reasoning fragment is signed with a TPM-bound provenance token (A -> B).
    3. Subagent B delegates to Subagent C.
    4. Subagent C reasoning fragment is signed with a merged provenance token (A -> B -> C).
    5. Subagent C calls a high-risk tool via MCP Any.
    6. HBRP Provider intercepts the call and validates the (A -> B -> C) lineage.
    7. The provider confirms the mission-root (A) and the authorized delegation path.
    8. The tool call is authorized, and the provenance is recorded in the audit log.
    9. If an unauthorized "Shadow Agent" (X) attempts to call the tool, the lack of a valid (A -> B -> C) lineage results in immediate interdiction.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Reasoning Fragment] --> B[HBRP Signer]
        B --> C[Hardware Enclave - TPM/SEP]
        C --> D[Attested Provenance Token]
        E[Tool Call Request] --> F[HBRP Validator]
        F --> G[Lineage Consistency Check]
        G --> H{Authorized?}
        H -- Yes --> I[Tool Execution]
        H -- No --> J[Interdiction & Audit]
        K[Mission-Root Manifest] --> G
    ```
* **APIs / Interfaces:**
    * `hbrp.SignReasoning(fragment, parentLineage) -> ProvenanceToken`: Signs a reasoning fragment with hardware attestation.
    * `hbrp.ValidateLineage(token, missionRoot) -> bool`: Validates the lineage of a reasoning path.
* **Data Storage/State:**
    * **HBRP Registry:** Temporary registry of active hardware-attested delegation paths.
    * **Sliding-Window Lineage Buffer:** Circular buffer for the last three hops of a reasoning chain.

## 5. Alternatives Considered
* **Full-Chain Signing:** Rejected due to token bloat and performance overhead; 3-hop sliding windows provide sufficient lineage for most swarms.
* **Session-Only Tokens:** Current baseline; rejected due to the risk of "Identity Dilution" in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HBRP tokens must be hardware-bound to prevent "Token Smuggling" by compromised subagents.
* **Observability:** Integrated with the "Mission-Resident Lineage Tracker" for visual auditing of the "Chain of Reason."

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
