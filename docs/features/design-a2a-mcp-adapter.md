# Design Doc: A2A-to-MCP Adapter
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rapid adoption of the A2A (Agent-to-Agent) Protocol and the growth of decentralized agent ecosystems like OpenClaw's "Moltbook," there is a critical need for interoperability between disparate agent frameworks. Currently, an agent built on OpenClaw cannot easily utilize a specialized agent built on LangGraph as a tool.

MCP Any will solve this by implementing an A2A-to-MCP Adapter. This adapter will allow any A2A-compliant agent to be registered as an MCP server, exposing its capabilities as standard MCP tools. This effectively turns "Agents" into "Tools" that can be discovered and invoked by any MCP-compatible LLM or agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Wrap A2A-compliant endpoints as standard MCP tools.
    * Support standardized handoff and message passing via A2A.
    * Enable cross-framework agent collaboration (e.g., OpenClaw -> MCP Any -> LangGraph).
    * Provide a unified interface for discovery of A2A agents.
* **Non-Goals:**
    * Replacing the A2A protocol itself.
    * Implementing framework-specific logic (e.g., LangGraph internal state management).
    * Creating a new LLM reasoning engine.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Invoke a specialized "Legal Analysis" agent (A2A-native) from a general-purpose "Personal Assistant" agent (MCP-native).
* **The Happy Path (Tasks):**
    1. Orchestrator registers the Legal Agent's A2A endpoint in MCP Any.
    2. MCP Any handshakes with the A2A endpoint and discovers its available tasks.
    3. MCP Any exposes these tasks as standard MCP tools (e.g., `legal_analyze_contract`).
    4. The Personal Assistant agent discovers the `legal_analyze_contract` tool via MCP Any.
    5. The Personal Assistant calls the tool; MCP Any translates the JSON-RPC call into an A2A message.
    6. The Legal Agent processes the request and returns a response via A2A.
    7. MCP Any translates the A2A response back to MCP format and returns it to the Personal Assistant.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[MCP Client Agent] -->|MCP JSON-RPC| B[MCP Any Gateway]
        B -->|A2A Adapter| C[A2A Translation Layer]
        C -->|A2A Protocol| D[A2A Native Agent]
        D -->|A2A Response| C
        C -->|MCP Result| B
        B -->|MCP Response| A
    ```
* **APIs / Interfaces:**
    * **Discovery**: Translates A2A `get_capabilities` to MCP `list_tools`.
    * **Execution**: Translates MCP `call_tool` to A2A `send_message` / `task_handoff`.
    * **State**: Bridges A2A session tokens with MCP request IDs.
* **Data Storage/State:**
    * Minimal local state; primarily maps A2A session identifiers to MCP session contexts.

## 5. Alternatives Considered
* **Native A2A Support in Clients**: Rejected because it requires every agent framework to implement every other framework's protocol. MCP Any provides a "Hub-and-Spoke" model.
* **Direct Agent-to-Agent HTTP**: Rejected due to lack of standardized discovery and capability negotiation provided by MCP.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**:
    * All A2A handoffs are governed by the MCP Any Policy Firewall.
    * Intent-bound tokens restrict the scope of context passed to the A2A agent.
* **Observability**:
    * Full tracing of the MCP-to-A2A translation.
    * Latency metrics for A2A handoffs tracked in the Resource Telemetry middleware.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
