# Design Doc: Predictive Shard Placement (PSP) Engine
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent meshes become increasingly distributed across edge devices and multi-cloud environments, the latency introduced by fetching remote context shards (Mean Time to Coordinate, MTTC) has become the primary performance bottleneck. Static re-sharding is insufficient for dynamic swarms where agent attention shifts rapidly between specialist tasks.

The Predictive Shard Placement (PSP) Engine solves this by speculatively migrating context shards to physical nodes before they are requested, based on real-time analysis of agent attention heatmaps and mission-root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce average MTTC for distributed teammate meshes by 40%+.
    * Implement real-time attention-heatmap tracking across all mesh nodes.
    * Enable speculative, non-blocking shard migration mediated by the DMR Hub.
    * Mandate hardware-re-attestation (MTA) for all migration events to neutralize CVE-2026-99101.
* **Non-Goals:**
    * Replacing the underlying storage layer (UEG).
    * Managing CPU/GPU resource scheduling (focused purely on state proximity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Minimize latency for a specialist subagent running on an edge device that needs to access mission-root state stored on a central node.
* **The Happy Path (Tasks):**
    1. The primary agent signals a shift in intent toward a task requiring the edge-based specialist.
    2. The PSP Engine detects the attention shift and identifies the required context shards.
    3. The PSP Engine triggers a speculative migration of shards to the edge node.
    4. The edge node performs Migration-Time Attestation (MTA) to verify memory boundaries.
    5. The specialist subagent executes the tool call with sub-millisecond local state access.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Intent] --> B[Attention Heatmap Tracker]
        B --> C[PSP Prediction Engine]
        C --> D[DMR Shard Migrator]
        D --> E[Destination Node MTA]
        E --> F[Local UEG Shard Cache]
    ```
* **APIs / Interfaces:**
    * `POST /v1/mesh/psp/predict`: Internal endpoint for attention-driven migration triggers.
    * `GET /v1/mesh/heatmap`: Telemetry endpoint for visualizing agent attention density.
* **Data Storage/State:**
    * Shard locations are tracked in the global Mesh Directory.
    * Heatmap data is stored in the Universal Episodic Graph (UEG).

## 5. Alternatives Considered
* **Reactive Sharding:** Rejected due to 500ms+ latency spikes during cold-start fetches.
* **Global Replication:** Rejected due to prohibitive bandwidth costs and synchronization overhead in 100+ node meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory Migration-Time Attestation (MTA) ensures that speculatively moved shards cannot be accessed by unauthorized processes on the destination node.
* **Observability:** Track "Prediction Hit Rate" (speculative migration utility) and MTTC reduction metrics.

## 7. Evolutionary Changelog
* **[2026-07-25]:** Initial Document Creation.
