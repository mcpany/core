# Design Doc: A2A Interop Bridge (Pseudo-MCP)
**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
As the AI agent ecosystem fragments into various specialized frameworks (OpenClaw, CrewAI, AutoGen, etc.), a significant interoperability gap has emerged. Agents within one framework cannot easily discover or call agents in another, leading to "framework silos." MCP Any is positioned to bridge this gap by acting as a universal bus.

The **A2A Interop Bridge** (Pseudo-MCP) will allow any compliant agent framework to be exposed as a standard MCP server. This means an LLM using Claude Code or Gemini CLI can "call" a CrewAI swarm or an OpenClaw subagent as if it were a standard MCP tool.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize the discovery of remote agents as MCP tools.
    * Implement a "Pseudo-MCP" wrapper that translates JSON-RPC tool calls into A2A protocol messages.
    * Support bi-directional state exchange between the calling agent and the bridge-exposed agent.
    * Enforce Zero-Trust security policies on inter-agent calls.
* **Non-Goals:**
    * Building a new agent framework from scratch.
    * Direct modification of 3rd party agent framework internals (bridge should be external or use stable APIs).
    * Universal translation of all framework-specific features (focus on tool-like task execution).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent System Architect
* **Primary Goal:** Enable a Claude-based "Manager Agent" to delegate a specialized coding task to an OpenClaw "Coder Agent" via MCP Any.
* **The Happy Path (Tasks):**
    1. Architect configures an OpenClaw framework endpoint in MCP Any as an `a2a_service`.
    2. MCP Any discovers the OpenClaw subagents and registers them as tools (e.g., `openclaw_fix_bug`).
    3. The Manager Agent queries MCP Any for available tools and sees `openclaw_fix_bug`.
    4. The Manager Agent calls `openclaw_fix_bug` with the repository context.
    5. MCP Any's A2A Bridge translates this into an OpenClaw task execution request.
    6. OpenClaw executes the task and returns the result to MCP Any.
    7. MCP Any passes the result back to the Manager Agent as a standard tool response.

## 4. Design & Architecture
* **System Flow:**
    `[Manager Agent] --(JSON-RPC / MCP)--> [MCP Any Gateway] --(Internal Dispatch)--> [A2A Interop Bridge] --(A2A Protocol / gRPC)--> [External Agent Framework (OpenClaw/CrewAI)]`
* **APIs / Interfaces:**
    * `A2ABridgeInterface`: Defines `DiscoverAgents()`, `InvokeAgent()`, and `StreamState()`.
    * `PseudoMCPWrapper`: Implements the standard MCP `ListTools` and `CallTool` methods by wrapping the `A2ABridgeInterface`.
* **Data Storage/State:**
    * Uses the `Shared KV Store` (Blackboard) to synchronize context between the caller and the bridged agent.
    * Session-bound metadata tracks the "Agent-to-Agent" handoff lifecycle.

## 5. Alternatives Considered
* **Direct Integration**: Hardcoding support for each framework into MCP Any. *Rejected* due to maintenance overhead and lack of scalability.
* **Standardizing a new A2A Protocol**: Trying to force everyone onto a new protocol. *Rejected* because the industry is already converging on MCP-like patterns; leveraging Pseudo-MCP is more pragmatic.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**:
    * Every A2A call must carry an "Intent-Scoped Token."
    * The `Policy Firewall` must validate that the Manager Agent has permission to delegate to the specific OpenClaw subagent.
    * Input/Output sanitization to prevent prompt injection across framework boundaries.
* **Observability**:
    * Integrated into the `Agent Chain Tracer (A2A)` UI.
    * Distributed tracing (OpenTelemetry) to track latency across the framework hop.

## 7. Evolutionary Changelog
* **2026-02-26**: Initial Document Creation.
