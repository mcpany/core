# Design Doc: Multi-Registry Discovery Broker (MRDB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent ecosystems evolve from single-registry toolsets to complex, multi-framework swarms, the need for a unified discovery layer has become critical. Frameworks like Gemini CLI have already pivoted to multi-registry architectures to aggregate tools from disparate sources. However, without a centralized broker, agents suffer from namespace collisions, fragmented capability maps, and discovery-phase security vulnerabilities.

MCP Any needs to solve this by acting as the authoritative "Registry of Registries." The Multi-Registry Discovery Broker (MRDB) will unify tool discovery across MCP, gRPC, UACO, and framework-specific registries (e.g., OpenClaw, Claude Code) into a single, secure, and searchable bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Unify tool discovery across heterogeneous registries.
    * Enforce absolute namespace isolation using hardware-attested "Registry Shards."
    * Provide a single, searchable capability bus for all connected agents.
    * Support "Auth-before-Discovery" patterns to hide sensitive tool schemas.
* **Non-Goals:**
    * Developing a new tool execution protocol (MRDB is for discovery, not invocation).
    * Providing a persistent global database of all tools (discovery is session/mission bound).

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Orchestrator
* **Primary Goal:** Aggregate database tools from a local MCP server and code-analysis tools from a remote OpenClaw registry without namespace collisions or unauthorized discovery.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes MCP Any with multiple registry endpoints (Local MCP, Remote OpenClaw).
    2. MRDB establishes hardware-attested tunnels to each registry.
    3. Specialist subagent requests tool discovery within a specific "Mission Scope."
    4. MRDB authenticates the subagent and returns a filtered, namespaced view of aggregated tools.
    5. Subagent selects a namespaced tool (e.g., `openclaw::analyze_code`) without awareness of the underlying registry complexity.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Subagent] -->|Authenticated Discovery Request| MRDB[Multi-Registry Discovery Broker]
        MRDB -->|Query| Shard1[Registry Shard: Local MCP]
        MRDB -->|Query| Shard2[Registry Shard: OpenClaw Remote]
        MRDB -->|Query| Shard3[Registry Shard: UACO]
        Shard1 -->|Schema| MRDB
        Shard2 -->|Schema| MRDB
        Shard3 -->|Schema| MRDB
        MRDB -->|Namespaced Aggregation| Agent
    ```
* **APIs / Interfaces:**
    * `GET /v1/discovery/search`: Unified search across all shards.
    * `POST /v1/discovery/registries`: Register a new heterogeneous tool source.
* **Data Storage/State:**
    * Ephemeral, mission-bound Registry Map stored in the hardware-locked memory enclave.

## 5. Alternatives Considered
* **Flat Registry**: Rejected due to namespace collision risks and the inability to enforce granular security policies across different framework origins.
* **Agent-Side Aggregation**: Rejected because it places the security burden on the agent and fails to provide a centralized audit point for tool discovery.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All discovery requests require hardware-attested mission tokens. Capability cards are masked until a mission-bound handshake is completed.
* **Observability:** Discovery-time logs will be integrated into the Action-Chain Sovereignty Monitor (ACSM) to track the lineage of tool discovery.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
