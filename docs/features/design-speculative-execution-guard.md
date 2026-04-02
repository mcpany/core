# Design Doc: Speculative Execution Guard
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
To remain the high-performance agent bus, MCP Any must minimize the "Consensus Tax" imposed by multi-agent security quorums. Following patterns seen in Gemini CLI, the Speculative Execution Guard allows agents to speculatively execute low-risk tool calls while background attestation is pending. This requires a robust mechanism to manage "Shadow State" and ensure that speculative results are never permanently committed or leaked to high-trust reasoning loops until verified.

## 2. Goals & Non-Goals
* **Goals:**
    * Allow sub-millisecond execution of Read-Only tools during background quorums.
    * Implement a "Shadow State" buffer to isolate speculative results.
    * Provide atomic "Commit-or-Discard" logic based on attestation signals.
* **Non-Goals:**
    * Speculative execution for high-risk write operations (e.g., `sudo`, `rm -rf`).
    * Replacing the final security quorum; it only overlaps with execution time.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Speed Swarm Orchestrator
* **Primary Goal:** Execute a `grep` search across a massive codebase while the "Search Auditor" agent is still verifying the tool's parameters.
* **The Happy Path (Tasks):**
    1. Agent issues a `grep` tool call.
    2. Speculative Execution Guard identifies it as "Low Risk" (Read-Only).
    3. Guard initiates background attestation quorum.
    4. Guard simultaneously executes the tool and stores results in a **Shadow Buffer**.
    5. Agent continues reasoning against the speculative results.
    6. Quorum returns "Approved."
    7. Guard commits the results to the persistent Blackboard and marks the reasoning fragment as "Verified."

## 4. Design & Architecture
* **System Flow:**
    `[Agent Call] -> [Guard Check] -> (Attestation Loop) + (Speculative Execution)`
    `[Speculative Result] -> [Shadow State Buffer] -> [Agent Reasoning]`
    `[Attestation Signal] -> [Commit/Discard Logic] -> [Persistent State]`
* **APIs / Interfaces:**
    * `state/speculative/buffer`: Temporary storage for un-attested fragments.
    * `state/speculative/commit`: Merges shadow state into the global mission root.
* **Data Storage/State:**
    Utilizes a Copy-on-Write (CoW) overlay for the Shared KV Store (Blackboard), ensuring that speculative writes are physically isolated from siblings.

## 5. Alternatives Considered
* **Blocking Execution:** Rejected due to the "Consensus Fatigue" pain point where agents stall for 2s+ during every tool call.
* **Optimistic Commitment:** Rejected because it allows state corruption; "Shadow State" with rollback is safer.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative results are marked with a `speculative` flag and are invisible to high-trust auditors until committed.
* **Observability:** Metrics track the "Speculative Success Rate" and the delta between speculative start and final attestation.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
