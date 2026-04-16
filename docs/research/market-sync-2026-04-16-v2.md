# Market Sync: 2026-04-16 (Iteration 2)

## Ecosystem Shifts & Competitor Analysis

### GNAP: The Git-Native Agent Protocol
* **Update:** GNAP has gained traction as a serverless coordination alternative. It uses 4 JSON files in a git repo to manage agent state and tasking.
* **Impact:** Agents can coordinate without a centralized gateway, using standard `git push/pull` as the transport layer.
* **Opportunity:** MCP Any can implement a **GNAP Adapter**, allowing "Bus-resident" agents to interact with "Git-resident" agents seamlessly.

### Windsurf & Project-Level Memory
* **Update:** Windsurf's "Cascade" mode now features deep project-level memory that persists across 5 parallel agents.
* **Challenge:** This proprietary memory silo competes with MCP Any's Shared KV Store (Blackboard).
* **Gap:** MCP Any needs to provide a "Memory Export/Import" standard to bridge IDE-specific memory silos.

## Autonomous Agent Pain Points
* **"Air-Gapped" Coordination:** Agents in restricted environments cannot use WebSocket-based buses.
* **Versioned Intent:** Teams want to "roll back" not just state, but the entire history of agent intents.

## Security & Vulnerability Scan
* **Git-Injection:** Malicious PRs could inject GNAP task files that coerce local agents into unauthorized actions.
* **Memory Mirroring:** Insecure memory exports from IDEs could leak mission-root constraints into untrusted context windows.
