# Design Doc: Intent-Boundary Telemetry (IBT) Hub
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms grow in complexity, "Fragment-Ghosting" attacks allow subagents to execute tool-calls that are semantically detached from the mission-root. MCP Any needs a centralized telemetry aggregator that validates every tool-call against its intent-boundary in real-time.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified sink for OpenClaw-compatible telemetry.
    * Enforce mission-root alignment for all coordination fragments.
    * Support sub-millisecond latency for telemetry ingestion.
* **Non-Goals:**
    * Long-term storage of reasoning traces (handled by external sinks).
    * Direct inter-agent message passing.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Enterprise Architect
* **Primary Goal:** Verify that a subagent's code-generation tool-call is authorized by the specific JIRA-ticket intent.
* **The Happy Path (Tasks):**
    1. Subagent initiates `generate_code` call.
    2. IBT Hub intercepts the call and extracts the `Intent-Token`.
    3. Hub validates the token against the resident mission-root in the Entangled State Broker.
    4. Call is allowed to proceed if lineage is verified.

## 4. Design & Architecture
* **System Flow:** `Agent` -> `IBT Middleware` -> `IBT Hub` (Validation) -> `Upstream Tool`.
* **APIs / Interfaces:** `PostTelemetry(Fragment)`, `ValidateIntent(Token)`.
* **Data Storage/State:** In-memory Bloom filters for active intent validation; Redis for cross-node state.

## 5. Alternatives Considered
* **Distributed Validation:** Rejected due to the risk of "Logic Drift" between nodes. A centralized Hub ensures an authoritative mission-root.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All telemetry must be signed by the originating agent's hardware-bound key.
* **Observability:** Exports prometheus metrics for "Intent Drift" scores.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
