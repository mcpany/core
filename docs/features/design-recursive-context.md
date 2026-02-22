# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agents evolve from single-instance tools to multi-agent swarms (e.g., OpenClaw, CrewAI), the need for recursive tool calls has increased. However, when Agent A calls a tool that is actually Agent B (via MCP), there is no standardized way to pass the execution context (user preferences, session ID, security boundaries) down the chain.

MCP Any needs to solve this by providing a standardized header-based protocol for context inheritance, ensuring that subagents remain "in character" and compliant with the original user's intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize a set of `X-MCP-Context-*` headers for passing metadata.
    * Implement middleware to automatically inject/propagate these headers.
    * Support session-based state inheritance.
* **Non-Goals:**
    * Solving the general LLM "Character Drift" problem (this is a transport/infrastructure solution).
    * Replacing the MCP protocol itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (e.g., OpenClaw User)
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars or re-authenticating at every step.
* **The Happy Path (Tasks):**
    1. User configures a root agent with a set of "Global Preferences" (e.g., `tone: professional`).
    2. Root agent calls a "Researcher" subagent via MCP Any.
    3. MCP Any automatically injects the `X-MCP-Context-Tone: professional` header into the request to the Researcher.
    4. Researcher subagent receives the context and adjusts its output accordingly without explicit instruction in the prompt.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Client->>MCP Any: Tool Call (with context in metadata)
        MCP Any->>Middleware: Context Extraction
        Middleware->>Registry: Resolve Upstream
        Registry->>Upstream: Forward Request + Context Headers
        Upstream-->>Registry: Response
        Registry-->>MCP Any: Result
        MCP Any-->>Client: Tool Result
    ```
* **APIs / Interfaces:**
    * Enhancement to `Upstream` interface to support a `Metadata` map.
    * New `ContextMiddleware` in `server/pkg/middleware`.
* **Data Storage/State:**
    * Context is transient and lives within the lifecycle of the request chain.

## 5. Alternatives Considered
* **Prompt Injection**: Manually appending context to the prompt. *Rejected* because it consumes tokens and is prone to being ignored by the model.
* **Session Persistence in DB**: Storing context in a central DB. *Rejected* for initial phase due to complexity and latency, but considered for future "Shared State" feature.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** We must prevent "Context Escalation" where a subagent tries to override inherited security constraints. Headers should be signed or validated by the Policy Firewall.
* **Observability:** Trace IDs must be propagated alongside context headers to allow for full-chain debugging.

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
