# Design Doc: Attention-Aware Shard Orchestrator (AASO)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms evolve into multi-node meshes, the physical distance between where an agent reasons and where its context state (shards) resides becomes a primary performance bottleneck. Current "Static Sharding" models lead to high MTTC (Mean Time to Coordinate) when specialist agents are distributed across different devices or cloud instances.

AASO addresses this by dynamically migrating context fragments to the physical node currently hosting the most "attentive" reasoning subagent. It transforms the Universal Agent Bus from a passive storage layer into an active, location-aware state mediator.

## 2. Goals & Non-Goals
* **Goals:**
    * Minimize MTTC by co-locating context shards with active reasoning processes.
    * Provide sub-100ms shard migration across multi-node meshes.
    * Implement "Attention Affinity" scoring based on real-time subagent monologues.
* **Non-Goals:**
    * Global state replication (AASO focuses on migration, not duplication).
    * Long-term archival storage management.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Agent Mesh Architect
* **Primary Goal:** Reduce "Cognitive Stall" when an OpenClaw specialist on Node A delegates to a Claude Code teammate on Node B.
* **The Happy Path (Tasks):**
    1. Agent A initiates a sub-task delegation to Agent B (on a remote node).
    2. AASO detects the shift in "Reasoning Focus" via the inter-node coordination bus.
    3. AASO calculates the "Attention Affinity" for the relevant context shards.
    4. AASO proactively triggers an encrypted P2P migration of the shard to Node B.
    5. Agent B resumes reasoning with local context access, bypassing the 200ms+ network roundtrip.

## 4. Design & Architecture
* **System Flow:**
    * [Subagent Monologue] -> [Attention Analyzer] -> [Affinity Score] -> [Migration Trigger] -> [P2P Encrypted Shard Transfer] -> [Local Shard Mount].
* **APIs / Interfaces:**
    * `POST /v1/mesh/shards/migrate`: Initiates a hardware-attested shard migration.
    * `GET /v1/mesh/affinity`: Retrieves real-time attention scores for active sessions.
* **Data Storage/State:**
    * Utilizes the **Universal Multimodal Memory Bus (UMMB)** for shard metadata and the **Attested Mesh Tunneling (AMT)** for secure transport.

## 5. Alternatives Considered
* **Global Replication:** Rejected due to the prohibitive cost of synchronizing 1M+ token context windows across all nodes in real-time.
* **Centralized State Hub:** Rejected to avoid single-point-of-failure and to maintain local-first sovereignty.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All migrations are cryptographically bound to a hardware-attested mission root. Destination nodes must prove mission-bound authority before a shard is unsealed.
* **Observability:** Integrated with the **Mesh Resilience Dashboard** to visualize shard location and migration latency.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
