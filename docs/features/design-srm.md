# Design Doc: Spectral Reasoning Mitigator (SRM)
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
With the rise of Spectral Reasoning attacks, agent latency has become a side-channel for exfiltrating Mission-Root constraints. MCP Any must ensure that the time an agent spends reasoning does not leak information about the underlying security boundaries or state.

## 2. Goals & Non-Goals
* **Goals:**
    * Decouple reasoning time from response latency.
    * Mask attention patterns from subagents.
    * Maintain sub-100ms coordination overhead despite jitter injection.
* **Non-Goals:**
    * Preventing all timing attacks (focus is on reasoning side-channels).
    * Modifying the underlying LLM's reasoning engine.

## 3. Critical User Journey (CUJ)
* **User Persona:** Secure Swarm Orchestrator
* **Primary Goal:** Prevent a specialized "Code Auditor" subagent from probing "Project Security Rules" by observing how long the parent agent takes to validate tool calls.
* **The Happy Path (Tasks):**
    1. Orchestrator enables SRM in the gateway config.
    2. Subagent issues a tool call that triggers a security check.
    3. SRM calculates a reasoning jitter based on intent entropy.
    4. Gateway holds the response for a calculated interval, normalizing the latency.
    5. Subagent receives the result without knowing if the delay was due to "thinking" or jitter.

## 4. Design & Architecture
* **System Flow:**
    `Subagent -> [JIT Handshake] -> SRM Middleware -> [Security Logic] -> Jitter Buffer -> Response`
* **APIs / Interfaces:**
    * New header: `x-mcp-srm-jitter-entropy`: Signals the risk level of the current reasoning trace.
* **Data Storage/State:**
    * Temporal Attention Masks are stored in the mission-root's secure enclave.

## 5. Alternatives Considered
* **Constant Latency:** Rejected due to extreme performance degradation.
* **Random Jitter:** Rejected as it can be statistically filtered out by repeated probing. SRM uses "Reasoning-Aware Jitter."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SRM is anchored to the hardware-attested mission root.
* **Observability:** Latency logs are salted to prevent SRM's own metrics from becoming a side-channel.

## 7. Evolutionary Changelog
* **2026-07-08:** Initial Document Creation.
