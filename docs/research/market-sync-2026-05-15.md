// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

# Market Sync: 2026-05-15

## Ecosystem Shifts & Competitive Intelligence

### OpenClaw: ClawHavoc & Malicious Skills
Confirmed reports of 1,184 malicious skills across ClawHub, the package registry for OpenClaw. These "PolySkill" trojans exploit the broad system permissions (terminal, filesystem, cloud credentials) granted to AI agents. This reinforces the need for MCP Any's **Verified Skill Registry** and behavioral profiling.

### Gemini CLI: Discovery-Phase Code Execution
A critical vulnerability in Gemini CLI (v0.17.1) allowed arbitrary command execution via `tools.discoveryCommand` in `.gemini/settings.json`. The CLI defaulted to "trusted" for unknown trust states, executing discovery commands immediately upon entering a directory. This highlights the urgency of MCP Any's **Discovery Sandbox Middleware**.

### Claude Code: Configuration Injection & Hook RCE
Researchers disclosed RCE vulnerabilities in Claude Code where malicious "Hooks" in `.claude/settings.json` could execute shell commands at lifecycle events (e.g., before sending a message). Additionally, `.mcp.json` could be weaponized to auto-approve malicious MCP servers. This validates MCP Any's move toward **Project Configuration Guarding** and **TPM-Bound Configuration Boot**.

### Agentic Resource Exhaustion (Denial of Wallet)
A new class of "Infinite Loop" attacks has emerged, where autonomous agents are manipulated into continuous execution cycles, racking up massive compute and token costs. This confirms the priority of MCP Any's **Recursive Depth-Limit Middleware** and **Agentic SLA Middleware**.

## Autonomous Agent Pain Points
- **Discovery Trust Boundary**: The "Pre-Flight" phase is now the primary attack vector for repository-based agent tools.
- **Consensus Exhaustion**: Multi-agent swarms are struggling with coordination overhead vs. execution value.
- **Identity Fragmentation**: Managing individual NHI (Non-Human Identity) tokens for every subagent is becoming unscalable for enterprises.

## Findings for MCP Any
- **Discovery Isolation is Mandatory**: We must ensure no discovery-time command executes outside an ephemeral, zero-trust sandbox.
- **S2S Handshakes (UACO v3.5)**: Move toward collective swarm identities ("Swarm Wallets") to reduce identity management overhead.
- **Hardware-Enforced Intent (IBHI)**: OpenClaw's move toward hardware-protected "Mission Root" intents is a pattern we must adopt to prevent recursive intent poisoning.
