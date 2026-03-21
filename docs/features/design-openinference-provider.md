# Design Doc: OpenInference Provider
**Status:** Draft
**Created:** 2026-06-21

## 1. Context and Scope
As AI agent systems move into production, the need for standardized, vendor-agnostic observability has become critical. The **OpenInference** specification has emerged as the industry standard for tracing LLM-based applications, instrumenting multi-step tool calls, reasoning trajectories, and agentic loops.

MCP Any, as a universal gateway, is uniquely positioned to act as the authoritative instrumentation layer for all connected MCP servers and agent frameworks. By implementing native OpenInference support, MCP Any allows users to export high-fidelity traces to any compatible observability platform (W&B Weave, LangSmith, Arize Phoenix) without framework-specific modifications.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a native OpenInference instrumentation layer for the MCP protocol.
    * Support OTel-compliant exporting of agent trajectories to external sinks.
    * Provide "Mission-Root" aware tracing that links multi-hop delegations to a single root intent.
    * Support automatic instrumentation for heterogeneous swarms (Claude Code + OpenClaw).
* **Non-Goals:**
    * Building a full observability dashboard (we will export to existing platforms).
    * Implementing custom, non-standard tracing protocols.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Ops Engineer
* **Primary Goal:** Debug a failed multi-agent workflow that spanned across Claude Code and OpenClaw specialists.
* **The Happy Path (Tasks):**
    1. The engineer enables OpenInference exporting in the MCP Any configuration.
    2. A complex mission is initiated, involving a Claude lead agent delegating a sub-task to an OpenClaw specialist via MCP Any.
    3. MCP Any automatically generates OpenInference spans for every tool call and inter-agent message, binding them to the hardware-attested mission root.
    4. The mission fails during the OpenClaw phase.
    5. The engineer opens Arize Phoenix and views a unified, hierarchical trace showing the complete trajectory across frameworks.
    6. The engineer identifies the specific reasoning fragment that led to the failure.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Framework] -> [MCP Any Middleware] -> [OpenInference Instrumentation] -> [OTel Collector] -> [Observability Platform]`
* **APIs / Interfaces:**
    * Extension of the MCP JSON-RPC handlers to inject and extract OpenInference context headers.
    * Integration with `go.opentelemetry.io/otel`.
* **Data Storage/State:**
    * In-memory buffering of spans before export.
    * State for active trace contexts stored in the session manager.

## 5. Alternatives Considered
* **Framework-specific SDKs**: Rejected because it requires modifying every agent framework and doesn't provide a unified view at the gateway level.
* **Custom JSON Logs**: Rejected because it lacks the hierarchical structure and standard compatibility required for production observability.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Traces will be scrubbed of PII via the **PII-Sovereign Context Scrubber** before export.
* **Observability:** This *is* the observability feature. It will include internal metrics for span generation latency.

## 7. Evolutionary Changelog
* **2026-06-21:** Initial Document Creation.
