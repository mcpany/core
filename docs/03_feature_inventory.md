<!--
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 -->

# Feature Inventory: MCP Any

## Evolution: [2026-05-06] Updates

### Proposed Additions
- **RAMS Shard Extension**: (P0) Advanced temporal memory isolation for the RAMS Hub, binding shard lifecycle to hardware-attested reasoning traces to mitigate Shadow Memory Exfiltration (SME).
- **Reasoning-Bound WebSocket (RBW) Controller**: (P0) Mandatory handshake protocol for persistent connections, requiring verifiable "Reasoning Proofs" before upgrading to high-privilege status.
- **Dynamic Capability Pruning (DCP) Middleware**: (P1) Just-in-time tool schema modification service that prunes irrelevant capabilities based on the active subagent task lifecycle.

### Priority Shifts
- **RAMS Hub**: (Re-affirmed P0) Now elevated with the requirement for "Temporal Isolation" to counter SME exploits.
- **Same-Origin Policy (SOP) Enforcer**: (Re-affirmed P0) Expanded to integrate with the RBW Controller for origin-locked reasoning proofs.

## Current Backlog (P0/P1)
- **Policy Firewall**: Rego/CEL based hooking for tool calls.
- **HITL Middleware**: Suspension protocol for user approval flows.
- **Recursive Context Protocol**: Standardized headers for subagent inheritance.
- **Shared KV Store**: Embedded SQLite "Blackboard" tool for agents.
