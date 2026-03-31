# Design Doc: Universal Episodic Graph (UEG) Memory Broker
**Status:** Draft
**Created:** 2026-07-13

## 1. Context and Scope
As AI agent swarms evolve from linear sessions to high-density horizontal meshes (e.g., Claude Code Agent Teams), the legacy "Blackboard" (Shared KV Store) model has become a primary bottleneck. Key-value state management fails to capture the rich, episodic relationships between parallel sub-missions, leading to "Context Amnesia" and "Memory Smearing" where agents lose track of the reasoning lineage that led to a specific state mutation.

The Universal Episodic Graph (UEG) Memory Broker is the authoritative evolution of MCP Any's state layer. It replaces flat KV storage with a hardware-attested graph database that cryptographically links reasoning fragments, mission intents, and environment snapshots into a cohesive "Episodic Fabric."

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a graph-based storage engine for agent state and reasoning traces.
    * Provide hardware-attested (TPM/SEP) lineage for every state mutation.
    * Support "Intent-Pinned" queries, allowing agents to retrieve state relevant to specific mission branches.
    * Neutralize "Memory Smearing" via cryptographically isolated sub-graphs.
* **Non-Goals:**
    * Implementing a general-purpose Knowledge Graph (this is focused on episodic execution state).
    * Replacing long-term vector memory (UEG handles active mission state, not global RAG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Lead Swarm Architect
* **Primary Goal:** Audit a multi-agent workflow to determine why a specific security policy was bypassed by a subagent.
* **The Happy Path (Tasks):**
    1. The Architect opens the UEG Inspector UI.
    2. They select the "Policy Bypass" event node.
    3. The UEG traverses the "Reasoning Lineage" edges back through three parallel sub-intents.
    4. The UI displays the hardware-attested internal monologues of the subagents that collectively drifted from the mission root.
    5. The Architect identifies the specific "Fragment Grafting" event where a poisoned tool output influenced the reasoning chain.

## 4. Design & Architecture
* **System Flow:**
    * Agents submit state updates as "Graph Fragments" containing the new state, the parent node ID, and a hardware-attested reasoning trace.
    * The UEG Broker validates the attestation and "pins" the fragment to the authorized mission branch.
    * Parallel teammates query the graph using "Context Lens" filters that only expose fragments consistent with their hardware-attested mission scope.
* **APIs / Interfaces:**
    * `AppendFragment(node_id, state_blob, attestation_token)`
    * `QueryByIntent(mission_root_id, depth_limit)`
    * `PruneBranch(intent_id)` (Triggers the Active Subagent Reaper)
* **Data Storage/State:**
    * Persistence: Embedded SQLite with Graph Extensions (utilizing Recursive Common Table Expressions).
    * Identity: Fragments are keyed by TPM-signed Lineage IDs.

## 5. Alternatives Considered
* **Vector-Only Memory:** Rejected because vector retrieval lacks the deterministic lineage required for security attestation and state-machine consistency.
* **Blockchain-based State:** Rejected due to prohibitive MTTC (Mean Time to Coordinate) in high-frequency local swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** UEG implements "Fragment Isolation." Fragments are cryptographically locked to the hardware identity of the issuing subagent and the intent of the parent.
* **Observability:** UEG provides a native "Lineage Trace" for OpenTelemetry, allowing standard APM tools to visualize the swarm's cognitive path.

## 7. Evolutionary Changelog
* **2026-07-13:** Initial Document Creation.
