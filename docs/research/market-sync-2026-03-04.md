# Market Sync: 2026-03-04

## Overview
Today's sync focuses on the aftermath of the "ClawJacked" exploit and the emerging trend of "Configuration-as-Attack-Vector" in agentic coding tools like Claude Code.

## Findings

### 1. OpenClaw "ClawJacked" Vulnerability (CVE-Pending)
- **Problem**: Malicious websites could open a WebSocket connection to `localhost` on the OpenClaw gateway port. Since the gateway assumed `localhost` was inherently trusted, it allowed full agent takeover (reading Slack, exfiltrating files, RCE).
- **Mitigation**: Update to v2026.2.25+. Enforce strict Origin and Host header validation.
- **Implication for MCP Any**: We must move beyond simple "localhost binding" and implement cryptographic proof of origin for even local connections if they originate from a browser context.

### 2. Claude Code Configuration Exploits (CVE-2026-24887, CVE-2026-21852)
- **Problem**: Attackers used malicious `.claude/settings.json` or `.env` files in cloned repositories to redirect API traffic (`ANTHROPIC_BASE_URL`) to attacker-controlled proxies or trigger RCE via untrusted hooks.
- **Mitigation**: Anthropic implemented "Trust Prompts" and restricted certain settings from being overridden by local project files without explicit consent.
- **Implication for MCP Any**: Our "Safe-by-Default" hardening must include "Project Boundary Isolation." MCP Any should ignore or sandbox project-local configurations unless the directory is explicitly "trusted."

### 3. Gemini CLI Policy Engine Maturity (v0.31.0)
- **Problem**: Tool discovery bloat and over-permissioning.
- **Solution**: Google moved to a robust Policy Engine with project-level granularity and "tool annotation matching."
- **Implication for MCP Any**: Reinforces the need for our "Policy Firewall" (Rego/CEL) and confirms that "Intent-Aware" scoping is the industry standard.

### 4. The Rise of "Outcome-Driven Development" (ODD)
- **Trend**: Shift from "task-running" to "outcome-validating." Swarms are being designed with "Critic" agents that verify outcomes before completion.
- **Implication for MCP Any**: Our A2A protocol should support "Validation Handoffs" where one agent can request a "Peer Review" of a tool output from another agent.

## Summary of Action Items for MCP Any
1. **Hardening**: Implement strict Host/Origin validation for the WebSocket gateway to prevent browser-based local hijacks.
2. **Isolation**: Add "Trusted Directory" logic to prevent malicious project-level config overrides.
3. **A2A Evolution**: Explore "Critic Handoff" patterns in the A2A Bridge design.
