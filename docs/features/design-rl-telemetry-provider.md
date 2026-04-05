# Design Doc: RL Telemetry Provider
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
With the release of OpenClaw-RL v1, agents are moving from static reasoning to continuous reinforcement learning from feedback. However, there is a significant "Training Data Gap" for local and specialized agents who lack standardized infrastructure to collect and export reasoning traces and feedback tokens.

MCP Any needs to solve this by acting as a high-fidelity, privacy-preserving telemetry sink. This provider will capture the interaction loop between agents, tools, and users, normalizing it for ingestion by RL training pipelines.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize telemetry export formats (JSONL/Protobuf).
    * Provide privacy-preserving redaction of PII from reasoning traces.
    * Implement non-blocking, high-frequency data capture.
    * Support OpenClaw-RL v1 feedback token standards.
* **Non-Goals:**
    * Performing the actual model training or fine-tuning.
    * Storing telemetry indefinitely (offloaded to external sinks).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agentic Developer / RL Researcher
* **Primary Goal:** Export sanitized reasoning traces from a local agent swarm to a training pipeline.
* **The Happy Path (Tasks):**
    1. User enables `RL Telemetry Provider` in MCP Any configuration.
    2. Agent performs a multi-step task involving tool calls.
    3. MCP Any captures reasoning monologues and tool outputs in real-time.
    4. Privacy filters automatically redact detected PII.
    5. MCP Any streams the sanitized trace to the configured OpenClaw-RL endpoint.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> MCP Any (Telemetry Middleware) -> PII Scrubber -> Telemetry Sink -> External RL Pipeline`
* **APIs / Interfaces:**
    * `/v1/telemetry/stream`: WebSocket for streaming real-time traces.
    * `/v1/telemetry/export`: POST endpoint for batch export of session logs.
* **Data Storage/State:**
    * Ephemeral memory buffer for active sessions.
    * Periodic flushing to configured persistent storage or remote API.

## 5. Alternatives Considered
* **Local-only Logging:** Rejected because it doesn't provide the real-time feedback loop required for active learning.
* **Model-Provider Telemetry:** Rejected because it locks users into specific clouds and often lacks local context/tool execution data.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Telemetry endpoints must require hardware-attested tokens. Scoped access ensures only authorized auditors can view raw traces.
* **Observability:** Monitor telemetry throughput and scrubber latency to ensure no impact on agent reasoning performance.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation.
