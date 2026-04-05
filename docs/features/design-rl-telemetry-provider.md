# Design Doc: RL Telemetry Provider
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
The release of **OpenClaw-RL v1** signals a shift toward agents that learn from natural conversation and environment feedback in real-time. To support this, agent infrastructure must provide high-fidelity telemetry that captures not just the "what" (tool calls) but the "why" (reasoning traces).

However, exporting this data to training pipelines creates significant privacy risks. The RL Telemetry Provider aims to bridge this gap by offering standardized, privacy-preserving hooks for telemetry export.

## 2. Goals & Non-Goals
* **Goals:**
    * Export high-fidelity reasoning traces and feedback loops to RL training pipelines.
    * Implement real-time, semantic PII scrubbing for all exported traces.
    * Support standardized telemetry formats compatible with OpenClaw-RL and other framework-neutral optimizers.
    * Provide "Verifiable Reward" attestation (e.g., hash-chained success signals).
* **Non-Goals:**
    * Hosting the actual RL training logic or model weights.
    * Storing raw, un-scrubbed conversation history permanently.

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Framework Developer
* **Primary Goal:** Collect performance data from a fleet of local agents to optimize a custom "SQL Specialist" model.
* **The Happy Path (Tasks):**
    1. The developer configures the RL Telemetry Provider as a middleware in MCP Any.
    2. Agents perform SQL-related tasks; reasoning fragments and tool results are intercepted.
    3. The Provider scrubs database schema names and sensitive data from the traces.
    4. A "Verifiable Reward" (e.g., query successful, returned 10 rows) is attached to the trace.
    5. The anonymized, attested packet is exported asynchronously to the training sink.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Agent Reasoner] --> B[Telemetry Middleware]
        B --> C[PII Scrubber]
        C --> D[Reward Attestor]
        D --> E[Async Sink]
        E --> F[OpenClaw-RL Pipeline]
    ```
* **APIs / Interfaces:**
    * `telemetry.ExportTrace(reasoningPacket, rewardSignal)`: Submits a trace for processing and export.
    * `telemetry.SetScrubbingPolicy(rules)`: Configures the semantic sanitization layer.
* **Data Storage/State:**
    * **Buffer Cache:** Ephemeral, in-memory buffer for traces awaiting asynchronous export.

## 5. Alternatives Considered
* **Direct Agent Logging:** Rejected because agents often hallucinate telemetry or include PII that should never leave the local environment.
* **Centralized Cloud Logging:** Rejected due to data sovereignty concerns. Telemetry must be sanitized *locally* by MCP Any before propagation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The scrubber uses the same "Content Governance" engine as the Prompt Path Protection layer.
* **Observability:** Telemetry export rates and scrubbing efficacy are visualized in the UI.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation focusing on OpenClaw-RL compatibility.
