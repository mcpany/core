# Design Doc: Intent-Gated Delegation (IGD)
**Status:** Draft
**Created:** 2026-04-03

## 1. Context and Scope
As AI agent swarms grow in complexity, the risk of "Intent Drift" and "Ghost Reasoning" becomes a critical failure point. Specialist subagents, once branched into speculative tasks, may continue to operate and access tools even after their specific line of reasoning has been invalidated or superseded by the parent agent. Current delegation models (MCP, UAB) treat delegation as a one-time permission grant rather than a continuous, mission-anchored relationship.

Intent-Gated Delegation (IGD) solves this by mandating that every tool call from a subagent is accompanied by a cryptographic proof that the current task directly serves an authorized, active intent branch of the parent mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a validator that interdicts tool calls if subagent intent is not verified.
    * Provide a mechanism for parents to "Seal" and "Prune" intent branches in real-time.
    * Mandate hardware-attested lineage for all task proposals (UACO).
* **Non-Goals:**
    * Providing natural language reasoning for *why* an intent serves a parent (this is semantic, not cryptographic).
    * Replacing existing A2A authentication; IGD sits *on top* of it.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Prevent a specialized "Code Reviewer" agent from executing a "File Delete" tool if that action was not explicitly branched from a verified "Refactor" intent.
* **The Happy Path (Tasks):**
    1. Parent Agent initiates a "Code Cleanup" mission.
    2. Parent branches a "Redundant Code Removal" intent and delegates to a subagent.
    3. MCP Any issues an `IGD-Token` bound to the branch.
    4. Subagent proposes a `delete_file` task.
    5. MCP Any's IGD Validator checks the `delete_file` call against the `IGD-Token` and the active mission manifest.
    6. If the call aligns with the "Removal" intent, it is authorized.

## 4. Design & Architecture
* **System Flow:**
    * **Intent Registry**: Tracks the hierarchy of `Mission-Root -> Intent-Branch -> Sub-Intent`.
    * **IGD Middleware**: Intercepts `tool/call` requests. It extracts the `x-mcp-intent-proof` header.
    * **Lineage Verifier**: Checks the signature of the proof against the Parent Agent's session key.
* **APIs / Interfaces:**
    * `x-mcp-intent-proof`: A signed JWT containing `{ root_id, intent_id, task_hash, timestamp }`.
* **Data Storage/State:**
    * Active intents are stored in a high-speed KV store (e.g., Redis or in-memory) to minimize latency impact.

## 5. Alternatives Considered
* **Static Permission Lists**: Rejected because they cannot handle the dynamic nature of speculative reasoning branches.
* **Manual HITL for every sub-call**: Rejected as it causes "Approval Fatigue" and breaks autonomous velocity.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If a subagent is compromised, IGD prevents it from "Escaping" its specific task-intent to perform lateral movements in the swarm.
* **Observability:** Audit logs will record every interdicted tool call with the specific "Intent Mismatch" reason.

## 7. Evolutionary Changelog
* **2026-04-03:** Initial Document Creation.
