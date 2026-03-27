# Design Doc: Reasoning-Aware Memory Segmentation (RAMS) Isolation Hub

**Status:** Draft
**Created:** 2026-05-06

## 1. Context and Scope
As AI agent swarms become more complex and autonomous, the "Blackboard" (Shared KV Store) has become a primary attack surface and source of operational instability. Specialized subagents often suffer from "Memory Smearing," where their local intents and specialized knowledge are overwritten or polluted by sibling agents sharing the same flat memory space. RAMS (Reasoning-Aware Memory Segmentation) aims to solve this by providing cryptographically isolated memory shards for subagents, ensuring that state is tied to specific reasoning intents and cannot be accessed or modified by unauthorized peers.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement cryptographically isolated "Intent-Sealed Shards" for subagents.
    * Provide a mandatory "Intent-Bound" access control layer for all Blackboard operations.
    * Enable secure, parental-authorized state handoffs between specialized agents.
    * Maintain sub-millisecond latency for shard mounting and memory access.
* **Non-Goals:**
    * Replacing the underlying SQLite storage for the Blackboard.
    * Managing the internal reasoning logic of the agents themselves.
    * Providing long-term persistent storage for non-agentic data.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Share secure, task-specific context between 3 specialized agents without risking state pollution or exfiltration.
* **The Happy Path (Tasks):**
    1. Parent agent spawns 3 subagents (Researcher, Coder, Reviewer) via the UAB.
    2. MCP Any automatically creates an "Intent-Sealed Shard" for each subagent, bound to their unique task ID.
    3. The Researcher writes findings to its isolated shard; these findings are invisible to the Coder by default.
    4. Upon task completion, the Researcher requests a "Verified Handoff" to the Coder.
    5. MCP Any validates the parental intent and merges the specific Researcher findings into the Coder's shard.
    6. The Coder proceeds with its task, shielded from any "Ghost Fragments" or residual state from the Researcher's internal reasoning.

## 4. Design & Architecture
* **System Flow:**
    * Subagents communicate with the Blackboard via the RAMS Isolation Hub middleware.
    * Every request must carry a cryptographically signed "Intent Token" (UACO-compliant).
    * The Hub maps the Token to a specific, isolated memory shard (virtualized SQLite table or namespace).
* **APIs / Interfaces:**
    * `MountShard(intent_token)`: Creates or retrieves a sealed shard for the given intent.
    * `CommitShard(intent_token, target_intent_token)`: Performs an atomic, authorized state handoff.
    * `PruneShard(intent_token)`: Forcefully purges a shard and its associated state upon task termination.
* **Data Storage/State:**
    * Utilizes the existing SQLite-based Blackboard but enforces row-level security and table-level isolation based on the Intent ID.

## 5. Alternatives Considered
* **Flat State with ACLs**: Rejected because path-based ACLs are vulnerable to "Intent Mirroring" and do not provide the cryptographic isolation required for high-trust swarms.
* **Separate SQLite Files per Agent**: Rejected due to the high I/O overhead and complexity of managing cross-agent state handoffs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All shard access is gated by UACO-signed tokens. "Intent-Sealed" shards ensure that even if a subagent is compromised, it cannot "smear" its state into the parent or siblings.
* **Observability**: The RAMS Isolation Monitor in the UI will provide real-time visualization of shard mounts, memory usage, and handoff events.

## 7. Evolutionary Changelog
* **2026-05-07:** Update: Intent-Sealed Shard Leak Mitigation.
    * **Context**: Research into enterprise production swarms revealed a potential "State Bleeding" vector where shard metadata could be leaked during rapid handoffs.
    * **Adjustment**: Implemented "Metadata Scrubbing" during the `CommitShard` lifecycle.
    * **Security Impact**: Neutralizes "Subdomain Hijacking" on localhost where shards could be identified via metadata side-channels.
* **2026-05-06:** Initial Document Creation.
