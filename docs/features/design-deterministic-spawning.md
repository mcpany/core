# Design Doc: Deterministic Agent Spawning & State Branching

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As agent swarms evolve from linear execution to complex, multi-level hierarchies (Parent -> Manager -> Worker), the lack of state consistency during sub-agent creation has become a major reliability bottleneck. The "OpenClaw 2026.2.17" update introduced deterministic spawning as a solution. MCP Any must implement a standardized "Session Branching" interface that allows parent agents to spawn sub-agents with a cryptographically verified and immutable "snapshot" of the current environment and tool state.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized API for spawning sub-agents with deterministic session IDs.
    * Implement "State Snapshots" (immutable clones of the session Blackboard) for each branch.
    * Ensure parent agents can "Reconcile" state changes from sub-agents via a merge policy.
    * Support Merkle-Proof style verification of the session lineage.
* **Non-Goals:**
    * Managing the LLM's internal reasoning for spawning (handled by the agent framework).
    * Providing long-term archival of all session branches (branches should be short-lived).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator (e.g., OpenClaw, CrewAI)
* **Primary Goal:** Spawn three parallel sub-agents to analyze different codebases without them interfering with each other's state, then merge the findings.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a session with MCP Any.
    2. Parent calls `mcpany_spawn_session(branch_name="analyzer-1")`.
    3. MCP Any creates an immutable snapshot of the current Blackboard and returns a new session ID: `parent-id.analyzer-1`.
    4. Sub-agent 1 uses the new ID to perform its task; its changes are local to the `analyzer-1` branch.
    5. Parent agent calls `mcpany_merge_session(branch_id="parent-id.analyzer-1")` to integrate the results back into the root session.

## 4. Design & Architecture
* **System Flow:**
    `Root Session -> Snapshot (COW) -> Branch Session -> State Modification -> Merge Protocol -> Root Session`
* **APIs / Interfaces:**
    * `POST /session/{id}/branch`: Creates a new deterministic branch.
    * `POST /session/{id}/merge`: Merges state from a branch back into the parent.
    * `GET /session/{id}/lineage`: Returns a Merkle-tree representation of the session history.
* **Data Storage/State:**
    * Uses "Copy-on-Write" (COW) logic for the SQLite Blackboard to ensure efficient branching without full database replication.

## 5. Alternatives Considered
* **Stateless Spawning**: Letting agents pass state via LLM prompts. *Rejected* due to context window limits and risk of hallucination/loss.
* **Full Database Replication**: Creating a new physical DB for every branch. *Rejected* as it doesn't scale for high-concurrency swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Branches inherit a "Restricted Clone" of the parent's capability tokens. Sub-agents cannot escalate permissions beyond their parent's scope.
* **Observability:** The UI will display a "Branching Timeline" (Marble Diagram) showing the full lineage of the swarm.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
