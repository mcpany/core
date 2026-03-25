# Design Doc: Adaptive Jitter Profiler (AJP)
**Status:** Draft
**Created:** 2026-07-06

## 1. Context and Scope
To neutralize "Enclave-Timing Side-Channel" attacks (CVE-2026-62001), MCP Any implemented mandatory monotonic jitter injection into all memory-broker responses. While effective for security, this has introduced a static 15-20ms "Coordination Tax" per handoff. In high-density meshes with 20+ teammates, this tax accumulates into significant "Reasoning Lag," hindering real-time performance.

The Adaptive Jitter Profiler (AJP) aims to move from a "one-size-fits-all" jitter model to a risk-aware strategy that scales timing variations based on the trust level and sensitivity of the accessed shard.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically scale jitter windows (2ms to 25ms) based on shard trust levels.
    * Provide sub-5ms latency for high-trust local teammate handoffs.
    * Maintain 20ms+ isolation for cross-framework or unauthenticated discovery requests.
* **Non-Goals:**
    * Completely removing jitter (Side-channel immunity is a hard requirement).
    * Modifying kernel-level timing primitives (AJP operates at the middleware layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 5 local agents for a high-frequency data processing task with <200ms total latency.
* **The Happy Path (Tasks):**
    1. Agent A and Agent B establish a "Pre-Attested Trust Bridge" within MCP Any.
    2. Agent A requests a context shard from the DME Broker.
    3. The AJP retrieves the "Risk Profile" for the Agent A/Agent B session.
    4. Since both are local and pre-attested, AJP selects the "Fast-Path" jitter profile (3ms).
    5. The response is returned with minimal latency, maintaining swarm responsiveness.

## 4. Design & Architecture
* **System Flow:**
    `[Request] -> [DME Broker] -> [AJP Lookup] -> [Profile Selection] -> [Jitter Injection] -> [Response]`
* **APIs / Interfaces:**
    * Internal `GetJitterProfile(requester_id, shard_id) -> JitterDuration`
    * Configuration `ajp.profiles: { "fast": "2-5ms", "standard": "15-25ms" }`
* **Data Storage/State:**
    * Risk Profiles are cached in memory-mapped regions for zero-latency lookup.

## 5. Alternatives Considered
* **Static Profile Selection**: Using a CLI flag to set jitter globally. Rejected because it forces a binary choice between security and performance across the entire mesh.
* **Network-Only Jitter**: Only applying jitter to remote requests. Rejected because local side-channels (cache timing) are equally dangerous in multi-tenant environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Profile Downgrade" attacks must be prevented by mandating hardware attestation for "Fast-Path" eligibility.
* **Observability:** Real-time jitter metrics will be displayed in the "Side-Channel Timing Heatmap."

## 7. Evolutionary Changelog
* **2026-07-06:** Initial Document Creation.
