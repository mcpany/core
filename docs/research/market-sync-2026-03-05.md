# Market Sync: 2026-03-05

## Ecosystem Updates

### OpenClaw (ClawJacked Vulnerability)
- **Insight**: A critical vulnerability (dubbed "ClawJacked") was discovered in OpenClaw, allowing websites to hijack local agents by exploiting the lack of origin verification in its gateway.
- **Impact**: Re-affirms the necessity for MCP Any to implement strict "Safe-by-Default" bindings and cryptographic origin attestation. Ease-of-use (local dashboards) is creating massive security holes in the agent ecosystem.

### Claude Code & Marketplace
- **Insight**: Anthropic has launched an official MCP Marketplace. Tool discovery is now handled via `/plugin` commands.
- **Impact**: MCP Any's "Universal MCP Discovery Service" must be able to bridge these marketplaces and provide a unified view across Claude, Gemini, and local tools.

### Gemini CLI
- **Insight**: Added `/agents refresh` for dynamic tool discovery and enhanced skill activation.
- **Impact**: Tool discovery is becoming a first-class interactive command. MCP Any should support "Slash-Command Bridging" for all major CLI agents.

### Agentic Swarms & Inter-Agent Protocols (IAP)
- **Insight**: The industry is moving toward "Agentic Swarms" (MAS) where specialized agents (Architect, Specialist, Critic) communicate via machine-speed Inter-Agent Protocols.
- **Impact**: MCP Any must evolve from a "Model-to-Tool" bridge to an "Agent-to-Agent" (A2A) resident state hub.

## Autonomous Agent Pain Points
- **Origin Trust**: Distinguishing between a local developer's command and a "confused deputy" command from a browser.
- **Context Pollution**: As agents become more specialized, they need "Lazy-Discovery" to avoid overloading their context with irrelevant tool schemas.
- **Command Injection**: CVE-2026-2256 highlights the failure of regex-based filtering for shell tools.

## Security Trends
- **Zero-Trust for Agents**: Moving from simple API keys to "Intent-Aware" permissions.
- **Attested Tooling**: Demand for signed MCP servers to prevent "Clinejection" supply chain attacks.
