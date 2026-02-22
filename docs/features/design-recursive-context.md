# Design Doc: Recursive Context Protocol (RCP)
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agents move from single-task execution to complex swarms, they often call subagents. Currently, these calls lose the original user's context, authentication tokens, and trace IDs. MCP Any needs to provide a standardized way to propagate this metadata across nested agent calls to ensure security, observability, and intentionality.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize header propagation for `X-MCP-Context-ID`, `X-MCP-Auth-Token`, and `X-MCP-Trace-ID`.
    * Provide a middleware that automatically injects these headers into upstream calls if present in the incoming request.
    * Enable subagents to inherit the "Security Profile" of the parent agent by default.
* **Non-Goals:**
    * Implementing a new authentication protocol (we use existing ones).
    * Managing the lifecycle of subagents (handled by orchestrators like OpenClaw).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator (e.g., OpenClaw)
* **Primary Goal:** Ensure a "Researcher" subagent has the same restricted access to a corporate database as the "Manager" agent that invoked it.
* **The Happy Path (Tasks):**
    1. Manager agent receives a request with an `X-MCP-Auth-Token`.
    2. Manager agent calls the Researcher subagent via MCP Any.
    3. MCP Any detects the `RCP` headers and automatically propagates them to the Researcher subagent's service.
    4. Researcher subagent successfully authenticates using the inherited token.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Client->>MCP-Any: Request + RCP Headers
        MCP-Any->>Middleware: Extract & Store Context
        MCP-Any->>Upstream: Forward Request + RCP Headers
        Upstream->>MCP-Any: Response
        MCP-Any->>Client: Response
    ```
* **APIs / Interfaces:**
    * Middleware `RecursiveContextMiddleware` implemented in the core pipeline.
    * Configuration flag `enable_rcp: true` per service/upstream.
* **Data Storage/State:**
    * Context is stored in the request's `context.Context` during the execution lifecycle.

## 5. Alternatives Considered
* **Manual Header Injection**: Rejected because it places too much burden on the agent developer and is error-prone.
* **Global Session State**: Rejected due to scalability concerns and the risk of session bleed in multi-tenant environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** We must ensure that a malicious subagent cannot "escalate" its privileges beyond what the original token allows.
* **Observability:** Trace IDs must be logged at every hop to allow for waterfall visualization of agent swarms.

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
