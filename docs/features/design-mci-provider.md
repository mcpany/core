# Design Doc: Mission-Constraint Inheritance (MCI) Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms grow in depth and complexity, maintaining the integrity of mission-critical constraints (e.g., "Do not access PII," "Maximum budget $10") across multiple layers of delegation has become a primary failure point. Current models rely on "Instruction Passing," which is vulnerable to "Reasoning Drift" and prompt injection. MCP Any needs an authoritative mechanism to ensure that constraints are inherited and enforced at the hardware (TPM/Secure Enclave) layer, independent of the agent's internal monologue.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-attested inheritance of mission constraints from parent to subagent.
    * Enforce constraints at the gateway/adapter layer using TPM-bound policy anchors.
    * Provide cryptographic proof of constraint compliance for every subagent tool call.
    * Neutralize autonomous divergence in deep delegation chains.
* **Non-Goals:**
    * Automatically resolving conflicting constraints (requires human-in-the-loop or supervisor arbitration).
    * Enforcing "Soft" stylistic preferences (focus is on "Hard" security/operational constraints).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Compliance Officer
* **Primary Goal:** Ensure that a 3rd-party subagent spawned by a primary developer agent cannot bypass "No-External-API" constraints.
* **The Happy Path (Tasks):**
    1. The supervisor agent initializes a mission with a set of hardware-bound constraints.
    2. The MCI Provider generates a "Constraint Anchor" signed by the host TPM.
    3. The primary agent delegates a sub-task, passing the MCI anchor.
    4. The subagent attempts to call an external API tool.
    5. The MCP Any gateway intercepts the call and verifies it against the hardware-bound MCI anchor.
    6. The gateway blocks the call as it violates the inherited constraint, even if the subagent's reasoning attempted to justify it.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        P[Parent Agent] -->|Set Constraints| MCI[MCI Provider]
        MCI -->|Sign Anchor| TPM[Hardware TPM]
        P -->|Spawn Subagent + Anchor| S[Subagent]
        S -->|Tool Call + Anchor| Gateway[MCP Any Gateway]
        Gateway -->|Verify Anchor| TPM
        TPM -->|Allowed| Exec[Execute Tool]
        TPM -->|Violated| Reject[Block & Log Violation]
    ```
* **APIs / Interfaces:**
    * `POST /v1/mci/anchor/create`: Create a signed constraint anchor for a mission root.
    * `POST /v1/mci/anchor/derive`: Derive a sub-anchor for a delegated task (scoped subset of constraints).
    * `GET /v1/mci/anchor/verify`: Verify a tool call against a constraint anchor.
* **Data Storage/State:**
    * Constraint definitions are stored in an encrypted vault; anchors are cryptographically linked to session-specific TPM keys.

## 5. Alternatives Considered
* **Context-Window Padding:** Rejected because it consumes tokens and can be ignored/overwritten by the model.
* **Supervisor-only Enforcement:** Rejected because it creates a bottleneck and cannot scale to deep, autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MCI anchors are immutable once signed. Any attempt to modify the anchor triggers an immediate mission-wide revocation.
* **Observability:** Constraint violations are logged with full reasoning-trace provenance via the RTP Auditor.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
