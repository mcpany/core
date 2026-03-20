# Design Doc: Mesh-Aware Blackboard Adaptor
**Status:** Draft
**Created:** 2026-05-01

## 1. Context and Scope
As agent swarms evolve toward "Mesh-Aware" intelligence (OpenClaw v2026.4.1), the traditional flat Key-Value (KV) store for state management is no longer sufficient. Agents need to reconcile conflicting intents and share state as a cohesive cognitive graph. The Mesh-Aware Blackboard Adaptor transforms the existing SQLite-based Blackboard into a graph-based "Intent Mesh."

## 2. Goals & Non-Goals
* **Goals:**
    * Transform flat KV storage into a directed acyclic graph (DAG) of intents.
    * Implement native intent reconciliation algorithms to resolve state conflicts.
    * Support "Cognitive Anchoring" by pinning root intents within the mesh.
    * Provide sub-millisecond graph traversal for deep agent swarms.
* **Non-Goals:**
    * Replacing the underlying SQLite engine (it will be extended with graph-like relational schemas).
    * Managing non-agent state (e.g., UI preferences).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Reconcile two conflicting subagent intents (e.g., "Delete File" vs "Archive File") into a unified "Mesh-Bound" decision.
* **The Happy Path (Tasks):**
    1. Parent agent defines a "Root Intent" in the mesh.
    2. Subagent A proposes a "Branch Intent" (Delete).
    3. Subagent B proposes a "Branch Intent" (Archive).
    4. The Mesh-Aware Adaptor detects the conflict and triggers a "Reconciliation Request."
    5. The swarm performs a CQ (Contextual Quorum) vote, and the winning intent is merged into the root.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Blackboard Tool` -> `Mesh-Aware Adaptor` -> `Graph-Relational SQLite`
* **APIs / Interfaces:**
    * `Blackboard.pushIntent(intent_node)`
    * `Blackboard.reconcile(node_id_a, node_id_b)`
    * `Blackboard.getCognitiveGraph()`
* **Data Storage/State:**
    * `blackboard.db` extended with `edges` and `intent_metadata` tables.

## 5. Alternatives Considered
* **Vector Database for State**: Rejected due to latency and lack of deterministic relational logic required for intent reconciliation.
* **Centralized Logic Engine**: Rejected in favor of a distributed mesh model where agents participate in the reconciliation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Every intent node is cryptographically bound to its proposing agent's identity.
* **Observability**: A new "Mesh-Aware Intent Visualizer" in the UI will display the real-time DAG of swarm reasoning.

## 7. Evolutionary Changelog
* **2026-05-01:** Initial Document Creation.
