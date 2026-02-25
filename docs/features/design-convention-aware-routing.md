# Design Doc: Convention-Aware Routing
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents (like Claude Code, OpenClaw, and AutoGen) increasingly take over complex software engineering tasks, a new problem has emerged: "Architectural Drift." Agents often generate code that is syntactically correct but violates the local architectural patterns, naming conventions, or library preferences of the specific codebase they are working on.

MCP Any is perfectly positioned to solve this by acting as a "Convention-Aware" gateway. By integrating with codebase intelligence tools (like Drift), MCP Any can automatically inject high-fidelity pattern metadata into tool calls, ensuring that the LLM has the necessary local context to produce "convention-perfect" code.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically detect local codebase patterns using an integrated intelligence provider.
    * Inject these patterns into the system prompt or tool descriptions dynamically.
    * Support multiple "Intelligence Providers" (e.g., Drift, custom scripts, LLM-based scanners).
    * Provide a scoring mechanism for how well a tool call aligns with local conventions.
* **Non-Goals:**
    * Replacing the agent's primary reasoning engine.
    * Automatically rewriting the agent's code (this is a discovery/context problem, not a post-processing problem).
    * Building a full codebase indexer from scratch (we will leverage existing tools).

## 3. Critical User Journey (CUJ)
* **User Persona:** Lead Systems Architect managing a distributed team of AI subagents.
* **Primary Goal:** Ensure all subagents follow the project's specific "Service-Repository" pattern and error-handling conventions without manual prompt engineering.
* **The Happy Path (Tasks):**
    1. The Architect enables the "Convention-Aware Routing" middleware in MCP Any.
    2. MCP Any runs a background scan (via Drift) to map the project's patterns (e.g., "all API calls must use the internal `FetchWrapper`").
    3. A subagent requests a tool to "Create a new API endpoint."
    4. MCP Any intercept the request, retrieves the "API Pattern" from the local index, and appends it to the tool's description.
    5. The LLM receives the tool description: "...use this tool to create an endpoint. NOTE: This project uses FetchWrapper for all calls."
    6. The agent generates code that follows the convention perfectly.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [MCP Any Gateway] -> [Convention Middleware] -> (Intelligence Provider / Drift) -> [Enriched Tool Call] -> LLM`
* **APIs / Interfaces:**
    * `GET /v1/patterns`: List currently detected codebase patterns.
    * `POST /v1/scan`: Trigger a fresh codebase intelligence scan.
    * Middleware Hook: `OnToolDiscovery` and `OnToolCall` hooks to inject metadata.
* **Data Storage/State:**
    * Local SQLite cache for storing mapped patterns and their confidence scores.

## 5. Alternatives Considered
* **Hardcoded Prompts**: Rejected because it doesn't scale across different projects or evolving architectures.
* **RAG-based Context**: Too slow and often retrieves irrelevant snippets. Convention-aware routing focuses on high-level *patterns* rather than specific code chunks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The intelligence provider must only have read-access to the codebase. It should not be able to execute code.
* **Observability:** Logs will track which patterns were injected and the "Convention Alignment Score" of the resulting tool outputs.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
