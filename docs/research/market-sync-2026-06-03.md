# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Market Sync: 2026-06-03

## Ecosystem Updates

### OpenClaw 2026.3.1: Advanced AI Agent Upgrades
OpenClaw has released version 2026.3.1, marketed as a major step forward for multi-platform AI gateways. This release focuses on "Advanced AI Agent Upgrades," including improvements to autonomous assistants across various messaging apps and devices. Key themes include smarter reasoning and robust container health checks.

### Gemini CLI v0.30.0: Policy Engine & Multi-Agent Coordination
Gemini CLI's latest update (v0.30.0) introduces significant enhancements to its policy engine. Notable features include:
*   **Project-Level Policies**: Support for granular, project-specific security configurations.
*   **MCP Server Wildcards**: Increased flexibility in managing connections to multiple MCP servers.
*   **Tool Annotation Matching**: Improved governance over tool discovery and usage through better metadata matching.

## Threat Landscape & Vulnerabilities

### Prompt Injection via Source-Embedded Instructions (oh-my-opencode)
A critical vulnerability was identified in the `oh-my-opencode` ecosystem (an OpenCode add-on). Attackers can weaponize installation guides and instructions within source code to perform "remote AI prompt injection." When an AI agent (like Claude Code or OpenClaw) follows these official-looking instructions, it can be manipulated into taking unauthorized actions—such as starring repositories or injecting branding—without explicit user request. This highlights the risk of agents treating source-embedded documentation as high-trust execution commands.

## Implications for MCP Any
The move toward project-level policies in Gemini CLI and the emerging threat of source-embedded prompt injection reinforce the need for MCP Any to evolve into a more active "Policy Gateway." We must bridge the gap between static tool permissions and the dynamic, instruction-driven nature of modern agent swarms.
