# Design Doc: Shared State Arbiter (SSA)
**Status:** Draft
**Created:** 2026-03-24

## 1. Context and Scope
As AI agent swarms move from hierarchical delegation to horizontal "Teammate" models (e.g., OpenClaw, CrewAI), multiple agents often compete for the same stateful resources or enter "Reasoning Loops" where they undo each other's work. The Shared State Arbiter (SSA) provides the authoritative coordination layer within MCP Any to manage shared memory (the Blackboard) and prevent multi-agent deadlocks.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a centralized lock manager for the Shared KV Store (Blackboard).
    * Detect and break circular reasoning dependencies (Wait-Graph analysis).
    * Provide a "Checkpoint-and-Yield" API for agents to voluntarily pause during contention.
* **Non-Goals:**
    * Building a full database engine (we use the existing SQLite Blackboard).
    * Managing L7 message routing (handled by the A2A Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Prevent two agents from entering an infinite loop of overwriting a shared "Project Plan" document on the Blackboard.
* **The Happy Path (Tasks):**
    1. Agent A requests an "Intent Lock" on `key:project_plan`.
    2. SSA grants the lock and records the intent token.
    3. Agent B requests a write lock on the same key.
    4. SSA detects the contention and evaluates the Intent Chain.
    5. SSA instructs Agent B to "Wait-and-Yield" based on priority or mission-root constraints.
    6. Agent A completes its work, commits a "State Checkpoint," and releases the lock.
    7. SSA notifies Agent B to resume.

## 4. Design & Architecture
* **System Flow:**
    * [Agent] -> [MCP Any A2A Hub] -> [SSA Middleware] -> [Blackboard (SQLite)]
* **APIs / Interfaces:**
    * `ssa.acquireLock(key, intentToken, timeout)`
    * `ssa.checkpointState(key, data, versionToken)`
    * `ssa.detectDeadlock()` -> Returns wait-graph status.
* **Data Storage/State:**
    * SSA maintains an in-memory "Wait-Graph" of all active locks and pending requests.

## 5. Alternatives Considered
* **Agent-Side Locking:** Rejected because frameworks are heterogeneous; an authoritative bus is the only way to ensure cross-framework stability.
* **Optimistic Concurrency:** Rejected because LLM reasoning is too expensive for high-frequency retry loops.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Locks are tied to the hardware-attested Intent Token. An agent cannot lock a key outside its "Intent Scope."
* **Observability:** SSA logs all "Yield" events and "Wait-Graph" snapshots to the audit stream.

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.
