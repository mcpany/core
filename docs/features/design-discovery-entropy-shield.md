# Design Doc: Discovery Entropy Shield (DES)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
The disclosure of the "Metadata Echoing" side-channel vulnerability has exposed a fundamental weakness in current agent discovery buses. By monitoring the micro-timing variations of metadata propagation (e.g., the exact millisecond a tool card is revealed), a malicious subagent can infer restricted mission-root constraints without ever directly accessing them.

The **Discovery Entropy Shield (DES)** is a critical security middleware designed to neutralize timing side-channels in the discovery phase. It ensures that discovery-time metadata is propagated with deterministic, hardware-attested jitter, making micro-timing analysis impossible for adversarial subagents.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement deterministic timing normalization for all tool discovery broadcasts.
    * Inject hardware-attested monotonic jitter (2ms - 50ms) into metadata receipts.
    * Decouple tool schema revelation from transport-layer handshakes.
    * Neutralize "Metadata Echoing" probes by saturating the discovery entropy.
* **Non-Goals:**
    * Eliminating all network latency (DES intentionally adds small, controlled latency).
    * Restricting legitimate tool discovery for authorized agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent from "mapping" the host environment's hidden constraints by timing the tool-discovery process.
* **The Happy Path (Tasks):**
    1. A subagent initiates a broadcast discovery request for "database" tools.
    2. MCP Any identifies the request and triggers the DES middleware.
    3. The gateway fetches matching tool cards but holds them in a "Smoothing Buffer."
    4. DES calculates a hardware-attested jitter profile based on the current system load.
    5. Tool schemas are released to the subagent at deterministic intervals, normalized to the jitter profile.
    6. The subagent receives the tools without being able to distinguish timing variations caused by hidden mission-root filtering.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Request] --> B[Discovery Hub]
        B --> C{DES Middleware}
        C --> D[Smoothing Buffer]
        D --> E[Jitter Injection]
        E --> F[Normalized Response]
    ```
* **APIs / Interfaces:**
    * Integrated directly into the existing `/v1/discovery` bus.
    * `x-mcpany-entropy-level`: Internal header for configuring jitter intensity.
* **Data Storage/State:**
    * Jitter seeds are derived from the hardware TPM's true random number generator (TRNG) to ensure they cannot be predicted by sibling processes.

## 5. Alternatives Considered
* **Static Delay:** Rejected because subagents could eventually "average out" a static delay to find the underlying signal.
* **Full Noise Injection:** Rejected as it would add too much overhead and degrade MTTC for high-frequency swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The DES logic itself must execute in a task-bound ephemeral sandbox to prevent metadata leakage from the shield.
* **Observability:** Entropy scores and jitter distributions are visualized in the "Attention Entropy Heatmap."

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
