# Design Doc: OpenClaw ContextEngine Adapter
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the release of OpenClaw 2026.3.7, the framework moved to a pluggable "ContextEngine" architecture. This allows agents to delegate memory and state management to external providers. MCP Any, with its "Shared KV Store" (Blackboard), is perfectly positioned to provide a standardized, tool-aware memory layer that works across OpenClaw swarms and other MCP-compliant agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a native adapter for the OpenClaw `ContextEngine` interface.
    * Map MCP Any's "Blackboard" tools to OpenClaw's memory read/write operations.
    * Ensure session-bound state isolation between different agent swarms.
    * Support "Context-Budgeting" to prevent subagents from being overwhelmed by parent context.
* **Non-Goals:**
    * Replacing OpenClaw's internal reasoning logic.
    * Providing a vector database (focus on structured state and metadata).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Architect using OpenClaw.
* **Primary Goal:** Enable an Architect agent and a Specialist agent to share a consistent "Project State" without manual handoffs.
* **The Happy Path (Tasks):**
    1. User configures OpenClaw to use `mcp-any` as its ContextEngine.
    2. Architect agent writes a "Project Goal" to the context via MCP Any.
    3. Architect spawns a Specialist subagent.
    4. Specialist agent automatically pulls the "Project Goal" from its local ContextEngine (powered by MCP Any).
    5. MCP Any ensures the Specialist only sees state relevant to its assigned task (Intent-Scoped Isolation).

## 4. Design & Architecture
* **System Flow:**
    `OpenClaw Agent` -> `ContextEngine (MCP Any Adapter)` -> `MCP Any Blackboard (SQLite)`
    * **Bridge Layer**: A Python/TypeScript shim that implements the OpenClaw `IContextEngine` interface and makes JSON-RPC calls to MCP Any.
    * **Mapping**:
        * `ContextEngine.store(key, value)` -> `mcpany_blackboard_write(key, value)`
        * `ContextEngine.retrieve(query)` -> `mcpany_blackboard_read(query)`
* **APIs / Interfaces:**
    * `mcpany_context_bridge`: A specialized toolset exposed to agents that wraps the Blackboard API with OpenClaw-compatible semantics.
* **Data Storage/State:**
    * Uses the existing `Shared KV Store` (SQLite) with mandatory `agent_id` and `swarm_id` tagging.

## 5. Alternatives Considered
* **Direct Vector DB Integration**: Rejected because structured state (KV) is more reliable for coordination than purely semantic retrieval in multi-agent handoffs.
* **File-based State Sharing**: Rejected due to concurrency issues and lack of granular security (Zero Trust).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Implements "Row-Level Security" in the Blackboard. Agents can only access keys tagged with their `swarm_id`.
* **Observability**: Tool Execution Timeline will show context "Hits" and "Misses" to help developers optimize agent memory usage.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation. Mapping MCP Any Blackboard to OpenClaw ContextEngine.
