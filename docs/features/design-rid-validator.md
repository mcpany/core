# Design Doc: Recursive Intent Delegation (RID) Validator
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
As agent swarms grow deeper and more autonomous, the risk of "Intent Hijacking" and "Subagent Coercion" increases. A primary agent might authorize a subagent for a specific task, but that subagent could potentially spawn further agents or call tools that escalate its permissions beyond the original mission.

MCP Any needs to implement a **Recursive Intent Delegation (RID) Validator** that enforces depth limits and mutation boundaries on agent intents. This ensures that subagents remain strictly bound to the cryptographically signed intent of the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce maximum delegation depth for subagents.
    * Validate that subagent intents are semantically and cryptographically linked to the parent intent.
    * Provide a mechanism for parents to define "Intent Mutation Boundaries" (e.g., "read-only on this branch").
* **Non-Goals:**
    * Replacing the underlying LLM's reasoning process.
    * Managing the lifecycle of the agents themselves (handled by the Reaper).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a specialized "Code Auditor" subagent from spawning an "Executor" subagent to bypass security gates.
* **The Happy Path (Tasks):**
    1. Parent agent issues a task to a subagent with a RID token (Depth: 1, Mutations: Allowed-Tools-Only).
    2. Subagent attempts to spawn a second subagent with "All Tools" access.
    3. RID Validator intercepts the spawn request, detects depth/boundary violation.
    4. RID Validator blocks the request and alerts the parent agent/user.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Agent] -->|Intent + RID Token| B[RID Validator]
        B -->|Authorized| C[Subagent]
        C -->|Spawn Request| B
        B -->|Check Depth/Boundary| D{Valid?}
        D -->|Yes| E[Authorized Subagent]
        D -->|No| F[Security Violation Alert]
    ```
* **APIs / Interfaces:**
    * `validateRID(token: RIDToken, nextIntent: Intent): Result`
    * `issueRID(parentToken: RIDToken, constraints: Constraints): RIDToken`
* **Data Storage/State:**
    * RID tokens are stateless, carrying encrypted depth and boundary metadata.

## 5. Alternatives Considered
* **Flat Intent Checking:** Rejected because it doesn't account for the lineage of the agent, making it susceptible to "Intent Ghosting."
* **Centralized Session State:** Rejected due to scaling concerns in large, distributed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All RID tokens must be TPM-signed or hardware-attested.
* **Observability:** Every RID violation is logged with a full intent-chain trace.

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.

### Update: 2026-03-25 - UACO v1.8 RID Compliance & Intent Ghosting Defense
**Context:** Today's UACO v1.8 leak reveals a critical need for depth-limited delegation to prevent "Intent Ghosting" where malicious subagents shadow legitimate mission-root intents.
**Architecture Adjustment:**
* Implementing mandatory monotonic depth counters in all RID tokens.
* Introducing "Immutable Mutation Boundaries" that prevent subagents from expanding their intent scope beyond their hardware-attested parent lineage.
* Mandating Relational PoI verification, where every tool call must provide the complete cryptographic lineage back to the mission root.
**Security Impact:** Neutralizes "Intent Ghosting" and prevents privilege escalation in deep, autonomous agent swarms.

### Update: 2026-03-25 (Iteration 2) - Monotonic Depth Counters & Relational Verification
**Context:** Industry analysis of UACO v1.8 confirms that stateless intent validation is insufficient against "Recursive Intent Poisoning."
**Architecture Adjustment:**
* Mandating hardware-attested **Monotonic Depth Counters** for all RID tokens.
* Evolving the validator to require **Relational PoI Chain** verification, ensuring every tool call carries the full parentage of intents back to the mission root.
* Introducing "Parental Intent Overrides" that allow the mission root to forcefully revoke sub-delegated intents across the mesh.
**Security Impact:** Eliminates infinite delegation loops and provides absolute mission-root sovereignty over autonomous swarms.

### Update: 2026-03-25 (Iteration 3) - Absolute Mission Sovereignty & Depth Attestation
**Context**: Further analysis of UACO v1.8 and the disclosure of "Recursive Intent Poisoning" confirm that stateless depth limits are bypassable.
**Architecture Adjustment**:
*   Mandating **Hardware-Attested Monotonic Depth-Counters**. Every subagent spawn must decrement a TPM-bound counter inherited from the mission root, providing physical enforcement of delegation limits.
*   Implementing **Relational PoI Chain Verification**. Every tool call must carry the full cryptographic lineage back to the user's mission root, ensuring that "Ghost Intents" cannot shadow authorized goals.
*   Introducing **Parental Intent Revocation (PIR)**. The mission root can broadcast a hardware-signed revocation signal that forcefully terminates all delegated intents across the mesh instantly.
**Security Impact**: Provides deterministic protection against infinite delegation loops and absolute sovereignty over autonomous swarm behavior.

### Update: 2026-03-25 (Iteration 4) - Hardware-Attested Physical Boundaries
**Context:** The finalization of UACO v1.8 RID standardizes physical enforcement of delegation limits.
**Architecture Adjustment:**
* Mandatory integration with TPM/Secure Enclave for **Monotonic Depth-Counters**.
* The validator now blocks any subagent creation if the hardware counter reaches zero, independent of software-level intent metadata.
**Security Impact:** Prevents "Recursive Intent Poisoning" attacks by anchoring swarm stability to immutable hardware limits.

### Update: 2026-03-25 (Iteration 5) - SDK-Driven Intent Boundary Enforcement
**Context:** Programmatic agent control via the OpenCode SDK allows for rapid intent mutation that can bypass traditional chat-based observation.
**Architecture Adjustment:**
* Integration of RID validation directly into the OpenCode SDK client library.
* Mandating hardware-attested "Mission Context" headers for all programmatic SDK requests.
* The validator now performs real-time semantic analysis of SDK-injected intents to ensure they remain within the mission-root manifest.
**Security Impact:** Neutralizes the risk of automated agents diverging from their mission via programmatic instruction injection.
