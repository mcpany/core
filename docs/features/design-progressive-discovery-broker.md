# Design Doc: Progressive Discovery Broker (PDB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent ecosystems scale, the number of available tools (MCP servers) is growing exponentially. Current architectures push all tool schemas to the model at once, leading to "Discovery Context Bloat." This saturates the LLM's context window with thousands of tokens of tool definitions before any reasoning occurs, reducing the effective space for actual mission logic and increasing latency.

MCP Any needs to solve this by implementing a metadata-first discovery pattern, where detailed tool schemas are only disclosed on-demand when a high-probability intent match is identified.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Progressive Disclosure" protocol for tool capabilities.
    * Reduce initial discovery context overhead by >80%.
    * Provide a Zero-Trust gateway for capability masking.
    * Maintain compatibility with legacy MCP clients.
* **Non-Goals:**
    * Building a full-text search engine for tools (handled by Lazy-MCP).
    * Executing tools on behalf of agents (handled by Execution Middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator with 50+ connected MCP servers.
* **Primary Goal:** Initialize an agent session without exhausting the 128k context window with tool schemas.
* **The Happy Path (Tasks):**
    1. Agent requests `/tools/list`.
    2. PDB returns a list of "Capability Cards" containing only the name and a 1-sentence description of each tool.
    3. Agent reasoning identifies a need for "Financial Analysis."
    4. Agent calls `activate_skill(tool_name="financial_ledger")`.
    5. PDB performs a mission-bound handshake and discloses the full JSON-RPC schema for the requested tool.
    6. Agent executes the tool with full schema awareness.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>PDB: list_tools()
        PDB-->>Agent: [Capability Cards (Metadata Only)]
        Agent->>Agent: Reasoning (Intent Match)
        Agent->>PDB: activate_skill(tool_name)
        PDB->>PolicyEngine: Verify Mission Token
        PolicyEngine-->>PDB: Approved
        PDB-->>Agent: Full Tool Schema (JSON-RPC)
    ```
* **APIs / Interfaces:**
    * `GET /discovery/cards`: Returns lightweight metadata.
    * `POST /discovery/activate`: Accepts a tool name and returns the full schema.
* **Data Storage/State:**
    * Capability Metadata Cache: In-memory store of tool summaries.
    * Secure Schema Vault: Encrypted store for full tool definitions.

## 5. Alternatives Considered
* **Client-Side Filtering:** Rejected because it doesn't solve the context bloat on the model's input side; the schemas are still transmitted over the wire.
* **Semantic Vector Search:** Complementary but doesn't replace the need for the PDB's disclosure control and security gating.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tool schemas can contain sensitive info (e.g., internal API endpoints). PDB ensures schemas are invisible to unauthorized subagents until a hardware-attested mission token is provided.
* **Observability:** Log disclosure events to track which agents are "learning" which skills and identify potential reconnaissance attempts.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
