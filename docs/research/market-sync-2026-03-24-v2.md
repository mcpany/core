# Market Sync: 2026-03-24 (v2)

## Ecosystem Updates

### UACO v1.7: Proof-of-Intent (PoI) Standard
* **Context**: The Universal Agent Coordination Protocol (UACO) v1.7 introduces PoI, shifting security from static identity to relational intent.
* **Impact**: Tool calls must now be cryptographically bound to a "Signed Intent." This directly addresses "Context-Mirroring" attacks (CVE-2026-34015), where subagents are tricked into unauthorized actions using inherited parent context.

### OpenClaw v2.4: Binary State Handoff (BSH)
* **Context**: High-density agent swarms (10+ agents) are hitting "Token Storms"—latency and cost spikes due to JSON-based state transfer.
* **Impact**: OpenClaw v2.4 adopts BSH (Protobuf/gRPC) for inter-agent state. MCP Any must implement BSH to maintain performance in deep swarms.

### Claude Code: Configuration-as-Execution Hardening
* **Context**: Post-mortem of CVE-2025-59536 reveals that project-local configurations (e.g., `.claude/settings.json`) are major RCE vectors via "Binary Smuggling" in WASM-based hooks.
* **Impact**: "Ghost Shell" profiling of un-attested hooks is becoming a mandatory security baseline for agentic infrastructure.

### Gemini CLI: Lazy Tool Registration
* **Context**: The move toward 1000+ tool libraries has made upfront schema registration (push-based) inefficient.
* **Impact**: Gemini CLI's `discoverMcpTools()` implementation highlights a trend toward "lazy" (on-demand) tool discovery to prevent context window bloat.

## Summary of Findings
* **Tool Discovery**: Shifting from proactive pushing to reactive, similarity-based discovery.
* **Local Execution**: Mandatory "Ghost Shell" sandboxing for automated configuration hooks.
* **Inter-Agent Comms**: Moving from JSON-over-HTTP to Binary State Handoffs (BSH) to mitigate "Token Storms."
* **Security**: Zero Trust is evolving into "Intent-Bound" trust, where every action must prove its logical lineage to the mission root.
