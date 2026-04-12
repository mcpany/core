# Design Doc: Risk-Aware Jitter Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Sovereign Node Tunneling (SNT), agent swarms are increasingly operating across distributed local environments. However, these inter-node tunnels are vulnerable to timing-based side-channel attacks, where subagents can infer mission-critical metadata or attention priority by monitoring response latencies.

MCP Any needs to implement a Risk-Aware Jitter Middleware that actively decorrelates response timing from internal reasoning state. This ensures that the "temporality of thought" remains sovereign and invisible to potentially compromised specialist agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Inject hardware-attested timing variations (jitter) into all inter-node tunnel responses.
    * Scale jitter intensity (2ms to 50ms) based on the sensitivity tier of the context shard.
    * Use TPM-bound monotonic counters to ensure jitter entropy is non-deterministic to subagents.
* **Non-Goals:**
    * Managing the underlying encryption of the P2P tunnels (handled by AMT Broker).
    * Providing generic jitter for non-mesh HTTP traffic.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Enterprise Swarm Administrator
* **Primary Goal:** Prevent a specialist subagent from mapping the parent's attention priority via latency monitoring.
* **The Happy Path (Tasks):**
    1. The admin defines "Sensitivity Tiers" for specific context shards (e.g., "Internal Monologue" = Tier 1, "Public Docs" = Tier 5).
    2. A specialist subagent requests a state fragment over a Sovereign Node Tunnel.
    3. The middleware intercepts the request and identifies the tier.
    4. The middleware retrieves a hardware-attested entropy token from the TPM.
    5. The middleware applies a calculated delay (e.g., 20ms for Tier 1) before releasing the response.
    6. The subagent receives the data with normalized, risk-aware latency.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Subagent->>AMT Broker: Request context fragment
        AMT Broker->>Context Shard: Fetch data
        Context Shard-->>Jitter Middleware: Raw response + Tier ID
        Jitter Middleware->>TPM: Get entropy token
        TPM-->>Jitter Middleware: Non-deterministic nonce
        Jitter Middleware->>Jitter Middleware: Calculate delay (Tier-weighted)
        Note over Jitter Middleware: Sleep(calculated_delay)
        Jitter Middleware-->>Subagent: Jittered response
    ```
* **APIs / Interfaces:**
    * Internal `IJitterProvider`: `InjectJitter(response []byte, tier int) []byte`
    * Configuration schema update to include `jitter_profiles` mapping tiers to delay ranges.
* **Data Storage/State:**
    * Tier mappings are stored in the hardware-locked `Mission Manifest`.

## 5. Alternatives Considered
* **Constant Latency Injection:** Rejected because it introduces unnecessary overhead for low-sensitivity data and is still vulnerable to differential analysis if the constant value drifts.
* **Software-based PRNG:** Rejected because it is vulnerable to entropy exhaustion and predictability in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Jitter ranges must be wide enough to mask internal processing spikes but narrow enough to maintain mesh liveness.
* **Observability:** Latency metrics will include a `jitter_overhead_ms` field to allow admins to audit the performance impact of temporal isolation.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
