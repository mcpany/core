# Design Doc: Intent-Aware Adaptive Jitter (IAAJ)
**Status:** Draft
**Created:** 2026-07-07

## 1. Context and Scope
Inter-agent coordination in high-density meshes is vulnerable to timing side-channel attacks, where subagents can map mission-root constraints by monitoring response latencies. While monotonic jitter (fixed 20ms) neutralizes these attacks, it introduces a significant "Coordination Tax" that slows down the entire swarm.

IAAJ evolves the jitter injection model to be risk-aware and intent-sensitive. By analyzing the real-time reasoning intent and the sensitivity of the context shard being accessed, MCP Any can dynamically scale jitter variations (from 2ms to 20ms). This optimizes performance for low-risk coordination (e.g., tool schema lookups) while maintaining maximum security for critical fragments (e.g., identity tokens).

## 2. Goals & Non-Goals
* **Goals:**
    * Implement dynamic jitter scaling based on context shard sensitivity.
    * Reduce inter-agent coordination latency by up to 40% for low-risk interactions.
    * Maintain side-channel immunity for PSS-compliant sensitive shards.
    * Integration with the Intent-Leakage Shield (ILS) to inform jitter profiles.
* **Non-Goals:**
    * Eliminating jitter entirely (security risk).
    * Predicting network-level latency (IAAJ operates at the application/middleware layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Teammate Swarm
* **Primary Goal:** Minimize coordination overhead during non-sensitive tool discovery while protecting root intent fragments.
* **The Happy Path (Tasks):**
    1. A specialist agent requests a cached tool schema from the Parallel Team Coordination Hub.
    2. The IAAJ Middleware identifies the request as "Low Sensitivity."
    3. The middleware injects a "Fast-Path" 2ms jitter into the response.
    4. The agent subsequently requests a Mission-Root intent shard.
    5. The IAAJ Middleware detects high-sensitivity intent and hardware-attested metadata.
    6. The middleware switches to a "High-Shield" 20ms jitter profile for that specific response.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Request] --> B[Sensitivity Classifier]
        B --> C{Risk Level?}
        C -- Low --> D[Fast-Path Jitter (2-5ms)]
        C -- High --> E[High-Shield Jitter (20ms)]
        D --> F[Response]
        E --> F
    ```
* **APIs / Interfaces:**
    * `iaaj.GetJitterProfile(intent Intent, shard ShardMetadata) -> JitterValue`
    * `X-MCP-Jitter-Profile`: Internal header for tracking jitter state.
* **Data Storage/State:**
    * **Sensitivity Mapping:** A local, hardware-attested registry mapping shard types to baseline risk levels.

## 5. Alternatives Considered
* **Static Adaptive Profiles:** Rejected because they don't account for "Intent Drift"—where a low-sensitivity shard might become critical depending on the reasoning context.
* **Kernel-Level Jitter:** Rejected due to lack of semantic awareness required for intent-based scaling.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Jitter profiles must be anchored to hardware-attested shard metadata. Subagents cannot influence their own jitter profiles.
* **Observability:** Performance gains and jitter distribution are visualized in the **Adaptive Jitter Control** dashboard.

## 7. Evolutionary Changelog
* **2026-07-07:** Initial Document Creation. Evolving from Adaptive Jitter Profiles (2026-07-06).
