# Design Doc: UACO v2.2 Intent Barrier Middleware
**Status:** Draft
**Created:** 2026-03-31

## 1. Context and Scope
With the introduction of OpenClaw v2.7 Sub-Intent Parallelization, agents can now branch a single mission into multiple parallel execution threads. However, this creates significant race conditions and non-deterministic states in the Shared KV Store (Blackboard). MCP Any needs a standardized way to synchronize these parallel intents to ensure swarm-wide consistency.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a synchronization barrier for parallel agent branches.
    * Provide "Snapshot-and-Merge" capabilities for the Blackboard.
    * Detect and resolve write conflicts between parallel sub-intents.
* **Non-Goals:**
    * Replacing the underlying LLM reasoning for branch merging.
    * Managing low-level OS thread synchronization.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Safely execute a code refactoring task in parallel across 5 modules without corrupting the global "dependency map" in the Blackboard.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a "Parallel Intent Session" via MCP Any.
    2. Parent agent branches 5 sub-intents, each receiving a "Branch Snapshot" of the Blackboard.
    3. Sub-agents execute tasks and write to their local branch state.
    4. Sub-agents hit an "Intent Barrier" call.
    5. MCP Any performs a "3-Way Merge" of the branch states.
    6. Parent agent is notified of the reconciled state and proceeds.

## 4. Design & Architecture
* **System Flow:**
    `[Parent Agent] -> [Branch Manager] -> [Snapshot API] -> [Sub-Agents]`
    `[Sub-Agents] -> [Intent Barrier] -> [Conflict Resolver] -> [Reconciled Blackboard]`
* **APIs / Interfaces:**
    * `intent/branch`: Create a new parallel intent branch.
    * `intent/barrier`: Wait for sibling branches and merge state.
    * `intent/resolve`: Manual conflict resolution hook for parent agents.
* **Data Storage/State:**
    Uses a "Copy-on-Write" (CoW) overlay for the SQLite Blackboard. Each branch writes to its own WAL (Write-Ahead Log) fragment.

## 5. Alternatives Considered
* **Global Locking:** Rejected due to performance bottlenecks in high-frequency swarms.
* **Agent-Side Merging:** Rejected because it increases token consumption and logic complexity for every agent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Branches are cryptographically bound to the parent's PoI token. A branch cannot "see" siblings unless explicitly authorized.
* **Observability:** Real-time "Branch Graph" visualization in the UI.

## 7. Evolutionary Changelog
* **2026-03-31:** Initial Document Creation.
* **2026-04-01:** Updated architecture to include "Reasoning-Bound Snapshot Divergence" resolution. Introduced a multi-stage merging process that reconciles divergent reasoning paths before committing state to the Blackboard.
