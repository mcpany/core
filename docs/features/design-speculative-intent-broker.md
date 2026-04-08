# Design Doc: Speculative Intent Broker (SIB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms evolve from linear sessions to high-density parallel teammate meshes (e.g., Claude Code Agent Teams), the risk of "Consensus Drift" and "Speculative Collisions" has become a primary bottleneck. Today's market sync revealed that teammates often overwrite each other's "Thought Anchors" in shared scratchpads when exploring divergent reasoning paths. The Speculative Intent Broker (SIB) is designed to act as the authoritative "Consensus Stabilizer" for the Universal Agent Bus, resolving these conflicts before they pollute the shared mission state.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a hardware-attested "Thought Anchor" registry for speculative reasoning branches.
    * Provide real-time detection of conflicting mutations to the shared teammate scratchpad.
    * Orchestrate intent reconciliation quorums to resolve divergent sub-intents before commitment.
    * Neutralize reasoning loops caused by race conditions in parallel sub-branches.
* **Non-Goals:**
    * Replacing existing mailbox sharding (it works *with* AMS/PAMS).
    * Governing final tool execution (this is handled by the MLE and DAL layers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Enable 5+ parallel teammates to explore divergent solutions to a complex task without corrupting the shared "Mission Truth."
* **The Happy Path (Tasks):**
    1. Parent agent initiates a mission and spawns 3 specialist teammates.
    2. Teammate A and Teammate B both initiate speculative branches that involve modifying the same `.scratchpad` fragment.
    3. SIB intercepts the speculative writes and identifies a "Thought Collision."
    4. SIB creates a temporary "Speculative Isolation Shard" for each teammate.
    5. Teammates reach a cryptographically bound quorum on the "Winning Intent."
    6. SIB merges the winning branch into the shared mission state and prunes the stale speculative shards.

## 4. Design & Architecture
* **System Flow:**
    `Teammate` -> `Speculative Write` -> `SIB Conflict Detector` -> `Isolation Shard` -> `Intent Quorum` -> `Shared Scratchpad`
* **APIs / Interfaces:**
    * `IntentRegistry`: `RegisterSpeculation(missionID, teammateID, intentHash) (ShardID, error)`
    * `ConflictResolver`: `ResolveCollision(shardA, shardB) (WinningShardID, error)`
* **Data Storage/State:**
    * Ephemeral, TPM-bound "Speculative Isolation Shards" managed by the Memory Broker.

## 5. Alternatives Considered
* **Global Locking**: Rejected due to 5s+ "Cognitive Stall" latencies observed in production swarms.
* **Optimistic Merging (No SIB)**: Rejected due to the high risk of reasoning loops and instruction corruption.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All speculative intents must be signed with hardware-attested lineage tokens.
* **Observability**: Conflict resolution events and quorum strengths are logged to the Mesh Sovereignty Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
