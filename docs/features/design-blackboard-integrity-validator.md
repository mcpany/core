# Design Doc: Blackboard Integrity Validator
**Status:** Draft
**Created:** 2026-03-20

## 1. Context and Scope
The `Shared KV Store` (Blackboard) is a critical component for state persistence in multi-agent swarms. However, as revealed by recent research into "State Injection" and "Context Mirroring," simple row-level security (RLS) is no longer sufficient to protect the integrity of the agent's reasoning path. A compromised subagent could attempt to "smuggle" unauthorized state or "shadow" existing task cards. MCP Any needs a cryptographic validation layer that ensures every change to the Blackboard is accompanied by a verifiable proof of the agent's current "Intent Scope" and its parentage.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Verifiable Lineage" for all Blackboard operations.
    * Every `SET` operation must include a cryptographic signature of the current `Intent Scope` and `Agent Identity`.
    * Provide a "State Audit Trail" that resists manipulation, even if a subagent gains elevated permissions.
    * Enable "Intent-Bound Isolation" where data is cryptographically sealed to a specific mission root.
* **Non-Goals:**
    * Replacing the underlying SQLite storage (remains the primary persistence layer).
    * Providing real-time synchronization between disparate MCP Any nodes (handled by Federated Policy Sync).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor Agent in a horizontal swarm.
* **Primary Goal:** Verify that a "Code Generator" subagent's temporary state was indeed authorized by the primary mission root.
* **The Happy Path (Tasks):**
    1. The Code Generator subagent attempts to write a `plan.json` to the Blackboard.
    2. The Blackboard Integrity Validator intercepts the request and requires a `Proof-of-Intent` (PoI) token.
    3. The subagent provides a token signed by the parent agent.
    4. The Validator verifies the signature against the hardware-attested mission root.
    5. The write is committed with a `lineage_hash` column populated.
    6. The Security Auditor agent later reads the `plan.json` and verifies the `lineage_hash` to ensure it hasn't been tampered with.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Blackboard Tool` -> `Integrity Validator` -> `SQLite (Lineage-Aware)`
* **APIs / Interfaces:**
    * `Blackboard.set(key, value, poi_token)`
    * `Blackboard.verify(key)` returns the full cryptographic lineage of a given key.
* **Data Storage/State:**
    * `blackboard.db`: Added `lineage_hash` and `parent_signature` columns to the primary KV table.

## 5. Alternatives Considered
* **Hash-Chaining all writes**: Rejected as too performance-intensive for high-frequency state updates.
* **Merkle Trees for state**: Considered for a future "Mesh-Aware" version but deemed out of scope for the initial local implementation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The `poi_token` must be session-bound and hardware-attested to prevent replay attacks.
* **Observability:** Lineage verification failures are logged as high-priority security alerts.

## 7. Evolutionary Changelog
* **2026-03-20:** Initial Document Creation.
