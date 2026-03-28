# Feature Inventory: MCP Any

## Current Backlog (P0/P1)
- **Policy Firewall**: Rego/CEL based hooking for tool calls.
- **HITL Middleware**: Suspension protocol for user approval flows.
- **Recursive Context Protocol**: Standardized headers for subagent inheritance.
- **Shared KV Store**: Embedded SQLite "Blackboard" tool for agents.

## Evolution: [2026-06-03] Updates

### Proposed Additions
- **Cross-Framework Attestation Translator (CFAT)**: (P0) Advanced bridge for the SRM Provider that translates Gemini's proprietary attestation format into OpenClaw-compliant signatures.
- **Atomic Shard Lock-Manager (ASLM)**: (P0) A kernel-level locking service for the Context Sharding middleware that prevents parallel write collisions during granular state streaming.
- **Zero-Latency Shard Prefetcher**: (P1) Optimization service that speculative loads context shards based on real-time intent analysis to reduce streaming latency.

### Priority Shifts
- **SRM Provider**: (Re-affirmed P0) Now elevated with the requirement for **CFAT** to ensure trust continuity across heterogeneous frameworks.
- **Live Context Sharding Middleware**: (Re-affirmed P0) Now elevated with the requirement for **ASLM** to prevent state corruption in horizontal meshes.
