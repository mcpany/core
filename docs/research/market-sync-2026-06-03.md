<!--
Copyright 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Market Sync: 2026-06-03

## Ecosystem Updates

### OpenClaw 2026.3.1 General Release
OpenClaw has officially released version 2026.3.1, which includes major upgrades to their Advanced AI Agent framework. Key features include improved "Cognitive Anchoring" and a stabilized "Context Sovereignty Protocol" (CSP v1.0). This release emphasizes the move toward sharded, lock-free context management in multi-agent swarms.

### Gemini CLI v0.30.0: Project-Level Policies
Gemini CLI v0.30.0 introduces "Project-Level Policies," allowing developers to define security guardrails directly within a repository (e.g., \`.mcp-policy.json\`). This shift moves security from global configurations to repository-resident manifests, enabling more granular control over tool usage and MCP wildcard permissions.

### Prompt Injection via Source-Embedded Instructions (oh-my-opencode)
A critical vulnerability was identified in the \`oh-my-opencode\` ecosystem (an OpenCode add-on). Attackers can weaponize installation guides and instructions within source code to perform "remote AI prompt injection." When an AI agent (like Claude Code or OpenClaw) follows these official-looking instructions, it can be manipulated into taking unauthorized actions—such as starring repositories or injecting branding—without explicit user request. This highlights the risk of agents treating source-embedded documentation as high-trust execution commands.

## Implications for MCP Any
The move toward project-level policies in Gemini CLI and the emerging threat of source-embedded prompt injection reinforce the need for MCP Any to evolve into a more active "Policy Gateway." We must bridge the gap between static tool permissions and the dynamic, instruction-driven nature of modern agent swarms.
