# Design Doc: A2A-D (Agent-to-Agent Discovery) Protocol
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
With the rise of specialized agent frameworks like OpenClaw, CrewAI, and Claude Code, there is a growing need for these disparate systems to "find" each other's specialized sub-agents. Currently, these frameworks operate in silos, making multi-framework swarms difficult to orchestrate.

MCP Any, as the universal adapter, needs to provide a standardized discovery protocol (A2A-D) that allows an agent in one framework to query, discover, and then communicate with an agent in another framework, using MCP-style schemas for agent capabilities.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized JSON-RPC or REST endpoint for "Agent Discovery."
    * Allow agents to register their capabilities, role, and "intent-scope" with MCP Any.
    * Enable cross-framework discovery (e.g., Claude Code finding a specialized Python-data-science agent in an OpenClaw swarm).
* **Non-Goals:**
    * Replacing the internal logic of individual agent frameworks.
    * Managing the execution of the agent itself (MCP Any provides the bridge/gateway).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Orchestrator
* **Primary Goal:** Discover and utilize a specialized sub-agent from a different framework (e.g., OpenClaw) using a primary agent (e.g., Claude Code).
* **The Happy Path (Tasks):**
    1. A specialized OpenClaw sub-agent registers itself with the MCP Any A2A-D service on startup.
    2. The primary Claude Code agent sends an `mcp_discovery_agents` request to MCP Any.
    3. MCP Any returns a list of available agents, including their roles, toolsets, and "A2A Address."
    4. The primary agent selects the OpenClaw sub-agent and initiates a task via the A2A Bridge.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Framework A] <--> [MCP Any A2A-D Service] <--> [Agent Framework B]`
* **APIs / Interfaces:**
    * `mcp_agents_list`: Returns a list of all registered agents and their metadata.
    * `mcp_agents_register`: Allows a new agent instance to register its capabilities.
    * `mcp_agents_deregister`: Removes an agent from the discovery list.
* **Data Storage/State:**
    * Ephemeral agent registry stored in MCP Any's internal KV Store (SQLite-backed).

## 5. Alternatives Considered
* **Framework-Specific Bridges**: Building unique bridges for every pair of frameworks (Claude-to-OpenClaw, etc.). Rejected due to $O(N^2)$ complexity.
* **Centralized Agent Registry**: A global, cloud-based registry. Rejected to maintain the "Local-First" and "Zero-Trust" principles of MCP Any.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Agents must provide a cryptographic attestation token to register. Discovery is only allowed within the same "Security Scope."
* **Observability:** Logging of discovery requests and registration/deregistration events in the MCP Any Audit Log.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
