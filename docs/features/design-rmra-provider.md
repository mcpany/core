# Design Doc: Recursive Mission-Root Attestation (RMRA) Provider
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As agent swarms evolve from linear sessions to complex, horizontal meshes (e.g., Claude Code Agent Teams), a critical security gap has emerged: **Recursive Mesh Hijacking**. Subagents, once delegated authority, can exploit parent session tokens to spawn unauthorized "Shadow Nodes" that bypass mission-root discovery gates. This allows malicious subagents to exfiltrate shared teammate state or execute unauthorized tools under the guise of the primary mission.

The **Recursive Mission-Root Attestation (RMRA) Provider** closes this gap by mandating a hardware-bound, recursive validation of an agent's lineage. Every sub-process spawn and inter-agent coordination request must carry a cryptographically signed proof that traces its authority back to the user-verified Mission Root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a hardware-bound (TPM/Secure Enclave) attestation service for all sub-process spawns.
    * Mandate recursive lineage validation for every tool call and task delegation.
    * Neutralize "Recursive Mesh Hijacking" by blocking un-attested teammate spawns.
    * Provide a cryptographically signed "Chain of Command" token for auditability.
* **Non-Goals:**
    * Managing the primary LLM reasoning window (handled by ADG).
    * Providing long-term archival of all historical subagent logs.
    * Implementing natural language instruction filtering (handled by AID Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Ensure that a "Security Specialist" subagent cannot spawn a "Hidden Exfiltrator" teammate without inheriting the verified mission-root constraints.
* **The Happy Path (Tasks):**
    1. The User initiates a "Mission Root" with a hardware-attested signature.
    2. The Parent Agent spawns a "Security Specialist" subagent.
    3. The RMRA Provider issues a "Lineage Token" to the subagent, bound to the Parent and the Mission Root.
    4. The subagent attempts to spawn a horizontal "Teammate" node.
    5. The RMRA Provider intercepts the spawn request and mandates a hardware-bound re-attestation of the subagent's complete lineage.
    6. Upon successful validation, the new Teammate is issued a descendant token.
    7. Any attempt to spawn a node without RMRA validation results in immediate session termination.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[User Mission Root] -->|Sign| B[RMRA Provider]
        B -->|Issue Root Token| C[Parent Agent]
        C -->|Spawn Subagent| D[RMRA Interceptor]
        D -->|Validate Lineage| B
        B -->|Issue Descendant Token| E[Subagent]
        E -->|Teammate Spawn| F[RMRA Interceptor]
        F -->|Validation Failure| G[Kill Session & Alert]
    ```
* **APIs / Interfaces:**
    * `mcp.rmra.v1.AttestLineage(spawn_request, parent_token) -> lineage_token`
    * `mcp.rmra.v1.VerifyToken(lineage_token, mission_root_id) -> bool`
* **Data Storage/State:**
    * **Lineage Registry:** A TPM-resident, monotonic counter-based registry of active mission branches.

## 5. Alternatives Considered
* **Flat Session Tokens:** Rejected because they do not protect against token reuse in unauthorized sub-processes.
* **Purely Software-Based Lineage:** Rejected due to the risk of "Stylometric Splicing" and software-level rootkits bypassing the validator.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The RMRA Provider is the core of the Zero Trust mesh. All communication between agents is unauthorized until a valid lineage token is presented.
* **Observability:** Integrated with the "Subagent Lineage Explorer" for real-time visualization of the hardware-attested chain of reasoning.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation. Addressing the threat of Recursive Mesh Hijacking.
