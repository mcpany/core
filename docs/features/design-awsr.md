# Design Doc: Attention-Weighted Shard Router (AWSR)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As AI agent swarms evolve toward horizontal, high-density teammate coordination (e.g., Claude Code Agent Teams), the latency of retrieving and synchronizing task-bound context shards has become a primary performance bottleneck. Current sharding models (AMS, SMS) are reactive, retrieving state only when an agent explicitly requests it. This leads to "Cognitive Stall" during teammate rotations.

The Attention-Weighted Shard Router (AWSR) is an authoritative context orchestration service designed to proactively route and pre-fetch context shards based on real-time attention-weight telemetry. By predicting which state fragments will be needed next, AWSR minimizes retrieval latency and ensures that mission-critical context is always in the agent's high-speed "hot-path."

## 2. Goals & Non-Goals
* **Goals:**
    * Implement proactive context shard pre-fetching based on attention-weight telemetry.
    * Reduce "Cognitive Stall" during horizontal teammate rotations and handoffs.
    * Provide a standardized "Attention-Weighted" routing API for heterogeneous frameworks.
    * Optimize resource allocation by prioritizing high-weight shards in high-speed buffers.
* **Non-Goals:**
    * Modifying the internal attention mechanisms of LLMs.
    * Replacing the underlying shard storage (handled by AMS/SMS).
    * Managing the semantic integrity of the shards (handled by ARI).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Teammate (e.g., OpenClaw specialist)
* **Primary Goal:** Access required context shards with sub-millisecond latency during a high-frequency task handoff.
* **The Happy Path (Tasks):**
    1. Parent Agent delegates a task to a specialist teammate.
    2. Parent Agent emits "Attention-Weight" telemetry signaling the importance of specific state fragments.
    3. AWSR intercepts the telemetry and identifies the target shards.
    4. AWSR proactively "routes" and pre-fetches these shards from cold storage into a zero-copy shared memory buffer.
    5. The specialist teammate spawns and requests the shards.
    6. AWSR serves the shards instantly from the high-speed buffer.
    7. The teammate begins reasoning without waiting for I/O, eliminating cognitive stall.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Attention Telemetry] --> B[AWSR Orchestrator]
        B --> C[Predictive Router]
        C --> D[Shard Pre-fetcher]
        D --> E[Zero-Copy Hot-Path Buffer]
        F[Teammate Request] --> G[AWSR Gateway]
        G --> E
        E --> H[Instant State Ingestion]
        I[Shard Storage - AMS/SMS] --> D
    ```
* **APIs / Interfaces:**
    * `awsr.PushAttention(weights, missionID)`: Ingests attention-weight telemetry.
    * `awsr.GetShard(shardID) -> Buffer`: Retrieves a shard, prioritized by its attention weight.
* **Data Storage/State:**
    * **Attention Heatmap:** Real-time map of mission-root attention distribution.
    * **Zero-Copy Shard Buffer:** High-speed shared memory region for "hot" context fragments.

## 5. Alternatives Considered
* **Global Prefetching:** Rejected due to memory exhaustion; swarms generate too much state to pre-fetch everything.
* **Reactive-Only Routing:** Current baseline; rejected due to the 150ms+ "Streaming Tax" observed in granular meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention telemetry must be cryptographically bound to the mission-root to prevent "Attention-Poisoning" by malicious subagents.
* **Observability:** Integrated with the "Shard-Aware Performance Heatmap" to visualize pre-fetch hit/miss rates.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
