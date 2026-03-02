# Design Doc: Programmatic Edge-Logic Bridge
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
With the rise of "Programmatic Tool Calling" (PTC) in models like Claude 4.6, agents are beginning to generate logic to coordinate multiple tool calls. However, executing this logic on the client-side (the agent) introduces high latency due to multiple round-trips. The "Edge-Logic Bridge" allows MCP Any to host and execute these small, verified logic snippets locally, closer to the tools.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a sandboxed environment (Wasm or restricted Go/Python) for orchestrating tool calls.
    * Reduce inter-agent latency by up to 80% for complex multi-tool workflows.
    * Expose a standard "Logic-Adapter" interface for agents to deploy and invoke edge functions.
* **Non-Goals:**
    * Building a general-purpose FaaS (Function-as-a-Service) platform.
    * Supporting long-running or resource-intensive background jobs.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Developer
* **Primary Goal:** Aggregate results from 5 different research tools into a single summarized response without 5 LLM round-trips.
* **The Happy Path (Tasks):**
    1. The agent identifies a pattern that requires calling Tool A, B, and C sequentially.
    2. The agent sends a `deploy_edge_logic` call to MCP Any with a logic snippet.
    3. MCP Any verifies the snippet against safety policies.
    4. The agent calls the newly deployed "Aggregate Tool."
    5. MCP Any executes the logic locally, calling A, B, and C, and returns the combined result.

## 4. Design & Architecture
* **System Flow:**
    - `Logic-Adapter` (Go) wraps a Wasm runtime (e.g., Wazero).
    - `Tool-Interconnect`: A virtual bus that allows Wasm functions to call other registered MCP tools via internal gRPC/JSON-RPC calls without network overhead.
* **APIs / Interfaces:**
    - `mcp.deploy_logic(name: string, runtime: string, code: string)`
    - `mcp.invoke_logic(name: string, args: object)`
* **Data Storage/State:**
    - Logic definitions stored in the local configuration or a persistent SQLite-backed registry.

## 5. Alternatives Considered
* **Client-side Orchestration:** Too slow (latency).
* **Hardcoded Aggregators:** Too rigid; doesn't support the dynamic nature of LLM reasoning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Strict memory and CPU limits for edge logic. No filesystem or network access unless explicitly granted via capability tokens.
* **Observability:** Detailed tracing of internal tool calls made by edge logic.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
