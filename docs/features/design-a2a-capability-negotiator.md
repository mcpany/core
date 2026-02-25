# Design Doc: A2A Capability Negotiator

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As the agent ecosystem grows, the challenge shifts from "executing a tool" to "finding the right agent with the right capability." MCP Any's A2A Bridge provides the transport, but agents still need a way to dynamically negotiate their needs. The A2A Capability Negotiator acts as a high-level discovery middleware that allows agents to query the swarm for specific intents (e.g., "Who can analyze this CSV?") and receive a set of optimized A2A tool endpoints.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide an intent-based discovery API for A2A-connected agents.
    *   Enable agents to "advertise" their capabilities (intents) to the MCP Any mesh.
    *   Filter and rank candidate agents based on performance metadata (latency, success rate).
    *   Support "Capability Handshakes" where two agents confirm protocol compatibility before task delegation.
*   **Non-Goals:**
    *   Direct agent-to-agent task execution (this is handled by the A2A Bridge).
    *   Replacing the standard MCP `tools/list` (this complements it for agentic intents).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent System Architect.
*   **Primary Goal:** Allow a "Planner Agent" to dynamically find and bind to a "Data Scientist Agent" within a federated mesh.
*   **The Happy Path (Tasks):**
    1.  The Data Scientist Agent registers with MCP Any, advertising the `intent:data_analysis` capability.
    2.  The Planner Agent calls the `negotiate_capability` tool with `query="perform statistical analysis on financial data"`.
    3.  The Capability Negotiator performs a similarity search across advertised intents and ranks the Data Scientist Agent as the top match.
    4.  The Planner Agent receives a specialized A2A tool reference and proceeds with delegation via the A2A Bridge.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Registration**: Agents send a "Capability Manifest" to MCP Any.
    - **Intent Indexing**: MCP Any stores manifests in a vector-enabled segment of the `Shared KV Store`.
    - **Negotiation Logic**: The `CapabilityNegotiatorMiddleware` intercepts negotiation requests, performs semantic matching, and returns A2A tool descriptors.
*   **APIs / Interfaces:**
    - `POST /a2a/negotiate`: Semantic query endpoint for agent discovery.
    - `PUT /a2a/capabilities`: Endpoint for agents to update their advertised capabilities.
*   **Data Storage/State:** Intent manifests and performance telemetry are persisted in SQLite.

## 5. Alternatives Considered
*   **Static Configuration**: Manually mapping every agent. *Rejected* because it doesn't scale in dynamic federated environments.
*   **Broadcast Discovery**: Agents shouting their needs to all peers. *Rejected* due to high network overhead and security concerns.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Agents can only discover capabilities they have "Discovery Scopes" for. Permission to discover does not imply permission to execute.
*   **Observability:** Log all negotiation events to track how the swarm "thinks" and organizes itself.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
