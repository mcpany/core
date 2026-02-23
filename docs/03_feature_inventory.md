# Feature Inventory: MCP Any

This is the rolling masterlist of priority features for MCP Any, categorized by strategic alignment and development status.

## High Priority: Universal Agent Bus (P0)

| Feature ID | Feature Name | Strategic Alignment | Status |
| :--- | :--- | :--- | :--- |
| **MCP-P0-001** | **Policy Firewall Engine** | Zero Trust Security | In Design |
| **MCP-P0-002** | **Recursive Context Protocol** | Swarm Context Inheritance | In Design |
| **MCP-P0-003** | **HITL Middleware** | Safety & User Approval | Planned |
| **MCP-P0-004** | **Shared Blackboard (KV Store)** | Agentic Mesh State Sync | Planned |

## Priority 1: Enterprise & Scalability (P1)

| Feature ID | Feature Name | Strategic Alignment | Status |
| :--- | :--- | :--- | :--- |
| MCP-P1-001 | Multi-Model Advisor | Collective Intelligence | Planned |
| MCP-P1-002 | Canary Tool Deployment | Operational Resilience | Planned |
| MCP-P1-003 | Advanced Tiered Caching | Performance | Planned |

## New Features: 2026-02-23 (Post-Market Sync)

### 1. Zero Trust Tool Sandboxing (P0)
*   **Description:** Ephemeral isolation for `command` upstreams.
*   **Reasoning:** Mitigates host-level file and environment variable leaks found in recent local execution exploits.

### 2. Recursive Context Propagation (P0)
*   **Description:** Automated inheritance of auth headers and session IDs for subagents.
*   **Reasoning:** Eliminates "Auth Fragmentation" in multi-agent swarms (OpenClaw/AutoGen).

### 3. Agentic Blackboard (P1)
*   **Description:** Secure, observable KV store for inter-agent communication.
*   **Reasoning:** Reduces context window bloat and enables complex multi-step reasoning across swarms.
