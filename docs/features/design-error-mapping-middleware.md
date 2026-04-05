# Design Doc: Universal Error Mapping (UEM) Middleware
**Status:** Draft
**Created:** 2026-11-02

## 1. Context and Scope
As MCP Any scales to support diverse upstream interfaces (REST, gRPC, CLI, GraphQL), handling disparate error formats has become a critical bottleneck for AI agents. When an agent receives an arbitrary HTTP 500 or a cryptic CLI exit code, its ability to autonomously recover is severely degraded. This "Error-Drift" leads to hallucinatory retry loops and prompt pollution.

The Universal Error Mapping (UEM) Middleware acts as a translation layer between upstream adapters and the agent client. It normalizes arbitrary failures into standardized, actionable `mcp.Error` payloads, ensuring agents receive consistent signals for recovery.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept arbitrary errors from all upstream Adapters (HTTP, gRPC, CLI, GraphQL).
    * Map diverse failure codes to a unified `mcp.Error` schema.
    * Provide "Actionable Suggestions" within the error payload to guide agent recovery.
    * Reduce prompt pollution caused by verbose or inconsistent upstream logs.
* **Non-Goals:**
    * Automatically retrying failed requests (handled by the Smart Retry Middleware).
    * Fixing underlying upstream bugs.
    * Managing hardware-level transport failures.

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Specialist Agent
* **Primary Goal:** Successfully recover from a transient database connection failure reported via a gRPC upstream.
* **The Happy Path (Tasks):**
    1. Agent invokes a tool via the MCP Any Gateway.
    2. The gRPC Adapter encounters a `Unavailable` error from the upstream service.
    3. UEM Middleware intercepts the gRPC error.
    4. UEM maps the gRPC `Unavailable` code to a standard `mcp.Error` with code `UPSTREAM_UNAVAILABLE`.
    5. UEM appends a `recoveryHint`: "Service is transiently down. Suggest retry after 2s."
    6. Agent receives the normalized error and follows the hint instead of hallucinating.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Gateway[MCP Any Gateway] -->|Request| UEM[UEM Middleware]
        UEM -->|Execute| Adapter[Upstream Adapter]
        Adapter -->|Raw Error| UEM
        UEM -->|Normalized mcp.Error| Gateway
        Gateway -->|Standardized Payload| Agent[AI Agent]
    ```
* **APIs / Interfaces:**
    * `uem.MapError(rawErr interface{}) *mcp.Error`: The core mapping function.
* **Data Storage/State:**
    * **Mapping Registry**: A static configuration (or dynamic YAML) defining the mapping rules between upstream codes and MCP errors.

## 5. Alternatives Considered
* **Client-Side Normalization**: Rejected because it requires every agent framework (OpenClaw, AutoGen) to implement complex mapping logic.
* **Adapter-Specific Error Handling**: Rejected because it leads to duplicate logic across 10+ adapters.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** UEM must ensure that sensitive upstream error details (e.g., stack traces, DB connection strings) are redacted during mapping.
* **Observability:** Logs all mapping events to the "Actionable Observability Feed" for human auditing.

## 7. Evolutionary Changelog
* **2026-11-02:** Initial Document Creation.
