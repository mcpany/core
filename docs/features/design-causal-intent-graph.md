# Design Doc: Causal Intent Graph (Blackboard v2)
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
In large agent swarms (>5 agents), "State Fragmentation" often occurs where specialized agents lose track of the high-level goal (the "Why") behind their specific tasks. This leads to redundant work, hallucinations, and security over-authorizations. The Causal Intent Graph evolves the Shared KV Store (Blackboard) into a persistent log of causal relationships between agent actions, ensuring "Intent Traceability" across the entire swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Track causal links (Parent-Child) between agent actions and state changes.
    * Persist high-level "Intent" metadata alongside every Blackboard entry.
    * Provide an "Intent Trace" API for agents to query the reasoning behind a piece of state.
    * Enable "Intent-Based Garbage Collection" where state is purged once its high-level goal is achieved.
* **Non-Goals:**
    * Storing full conversation history (handled by the LLM/Agent framework).
    * Providing real-time streaming of state changes (handled by the notification middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator.
* **Primary Goal:** Audit *why* a subagent performed a specific file deletion.
* **The Happy Path (Tasks):**
    1. Parent Agent sets a high-level goal: "Refactor Database Schema."
    2. Subagent A (DB Analyst) creates a "Migration Plan" in the Blackboard, linked to the "Refactor" intent.
    3. Subagent B (Code Generator) reads the plan and generates a delete command for an old table, linking it to Subagent A's plan.
    4. Auditor Agent queries the Causal Intent Graph for the delete command and receives a full trace back to the "Refactor Database Schema" goal.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Blackboard v2 Tool` -> `Graph Database / SQLite-LSM`
    1. **Context Enrichment**: Every `set` operation requires an `IntentID` and optional `ParentNodeID`.
    2. **Causal Mapping**: MCP Any maintains a directed acyclic graph (DAG) of these entries.
    3. **Query Interface**: Agents can use `get_intent_trace(key)` to retrieve the causal path.
* **APIs / Interfaces:**
    * `Blackboard.set(key, value, intent_id, parent_id?)`
    * `Blackboard.get_trace(key)` -> `[]IntentNode`
* **Data Storage/State:**
    * `intent_graph.db`: SQLite with a recursive CTE for graph traversal or a dedicated graph extension.

## 5. Alternatives Considered
* **Vector Embeddings for State**: Too non-deterministic for causal auditing.
* **Centralized Orchestrator State**: Doesn't scale well for decentralized swarms; the Blackboard provides a decoupled alternative.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Intent-based isolation ensures that an agent can only access nodes within its causal lineage.
* **Observability**: A visual "Intent Map" in the MCP Any UI to help developers debug swarm logic.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
