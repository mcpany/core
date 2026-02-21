# Feature Inventory: MCP Any

This is a rolling masterlist of priority features for MCP Any.

## Current Backlog

| Feature ID | Name | Priority | Status | Description |
|------------|------|----------|--------|-------------|
| FEAT-001 | HTTP Adapter | P0 | GA | Basic REST/JSON proxying. |
| FEAT-002 | gRPC Adapter | P0 | Beta | Dynamic gRPC invocation via reflection. |
| FEAT-003 | Command Adapter | P0 | Beta | Safe local CLI execution. |
| FEAT-004 | Policy Engine | P1 | Alpha | Rate limiting and basic access control. |

## New Features: 2026-02-21

| Feature ID | Name | Priority | Status | Description |
|------------|------|----------|--------|-------------|
| FEAT-005 | Inter-agent Context Bus | P0 | Proposed | A shared state layer for context inheritance (auth, scope) across agent swarms. |
| FEAT-006 | MCP-native Subagent Sandbox | P1 | Proposed | Automated Docker-bound isolation for tools invoked by subagents. |
| FEAT-007 | Dynamic Relevance Filtering | P2 | Proposed | Context-aware tool discovery to reduce agent context window bloat. |
