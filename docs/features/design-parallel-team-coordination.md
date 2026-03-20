# Design Doc: Parallel Team Coordination Hub

**Status:** Draft
**Created:** 2026-05-11

## 1. Context and Scope
With the release of Claude Code "Agent Teams," the industry is moving from sequential subagent execution to parallel multi-agent swarms. Currently, MCP Any handles individual agent sessions but lacks the high-speed synchronization layer required for teammates to work concurrently on the same mission. The Parallel Team Coordination Hub provides the message bus and state reconciliation logic necessary for high-performance agent teams.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a low-latency (sub-10ms) message bus for inter-teammate communication.
    * Implement "Snapshot-and-Merge" for the Shared KV Store (Blackboard) to handle parallel state mutations.
    * Support "Mission-Root" anchoring to ensure all teammates remain aligned with the primary objective.
    * Facilitate "Task Bidding" where teammates can claim and release tasks from a shared list.
* **Non-Goals:**
    * Orchestrating the LLM reasoning loops themselves (this is done by the framework, e.g., Claude Code).
    * Providing long-term persistent storage beyond the session lifecycle.

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator
* **Primary Goal:** Coordinate 3 parallel agents (Coder, Reviewer, Tester) to implement a feature without state corruption.
* **The Happy Path (Tasks):**
    1. The lead agent initializes a "Mission Team" via MCP Any.
    2. Parallel agents (Coder, Reviewer, Tester) connect to the Hub using a shared Mission Token.
    3. The Coder agent writes a draft to the Blackboard; the Hub snapshots the state.
    4. The Tester agent reads the draft and performs parallel verification.
    5. The Hub manages concurrent updates to the Blackboard, using conflict resolution to merge the Tester's results with the Coder's subsequent edits.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Lead[Lead Agent] --> Hub[Parallel Coordination Hub]
        Hub --> T1[Teammate 1]
        Hub --> T2[Teammate 2]
        Hub --> Blackboard[(Blackboard Snapshot-and-Merge)]
    ```
* **APIs / Interfaces:**
    * `InitializeTeam(mission_root_intent)`: Creates a new coordination scope.
    * `BroadcastMessage(sender_id, payload)`: High-speed teammate messaging.
    * `MergeState(agent_id, kv_deltas)`: Atomic application of parallel state changes.
* **Data Storage/State:** Uses an in-memory Graph of intents and a versioned SQLite Blackboard for reconciliation.

## 5. Alternatives Considered
* **Framework-Native Coordination**: Rejected because it creates siloed agents. A gateway-level hub allows OpenClaw agents and Claude Code agents to work in the same team.
* **Lock-Based State Management**: Rejected due to latency and deadlock risks in autonomous environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Teammates are only granted access to the mission-specific context shard. "Intent-Drift" detection prevents teammates from diverging from the Mission Root.
* **Observability**: Real-time "Marble Diagrams" in the UI show concurrent agent flows and merge events.

## 7. Evolutionary Changelog
* **2026-05-11:** Initial Document Creation.
* **2026-05-12:** Resolving Local Port Exposure.
    * **Context:** Today's market sync revealed a new exploit pattern in OpenClaw subagent routing involving local port hijacking.
    * **Architecture Adjustment:**
        * Deprecating TCP/UDP-based local teammate communication in Section 4.
        * Introducing **Isolated Named-Pipe Transport** (UNIX domain sockets) for inter-agent and inter-teammate coordination.
    * **Security Impact:** Mitigates unauthorized host-level file access and MitM attacks by rogue subagents by eliminating port exposure.
    * **2026-05-13:** Coordination Token Optimization.
        * **Context:** Claude Code documentation highlights substantial coordination overhead and token consumption in parallel agent teams.
        * **Architecture Adjustment:**
            * Introducing **Coordination Token Compression** in Section 4.
            * Implementing a "Deduplication Proxy" within the named-pipe transport to strip redundant context during high-frequency teammate message passing.
        * **Economic Impact:** Reduces token consumption for parallel swarms by up to 40%, making multi-agent coordination economically viable for large-scale enterprise tasks.
