# Market Sync: 2026-03-02

## Ecosystem Shifts

### 1. OpenClaw (v2026.2.26)
- **Secrets Management**: Shifted from static configs to a dedicated secrets workflow with audit logs and dynamic reloading.
- **Browser Control**: Significant hardening of browser automation stability. Native support for multi-agent coordination during browser tasks.
- **Context Expansion**: Support for 1M token windows, yet emphasizing "Layered Multi-Agent" systems over single-agent monoliths.

### 2. Claude Code (v2.1.x)
- **Lazy Discovery by Default**: "MCP Tool Search" is now standard. Tools are deferred and discovered via search when descriptions exceed 10% of the context window.
- **Browser Integration**: Direct browser control via a Chrome extension is now a core capability.
- **LSP Tooling**: Deep integration with Language Server Protocol for precise code navigation.

### 3. Gemini CLI (v0.31.0)
- **Policy Engine Maturity**: Support for project-level policies and tool annotation matching.
- **Sequential Planning**: Formalized a 5-phase sequential planning workflow for complex tasks.
- **Browser Agent**: Introduced experimental browser interaction tools.

### 4. Security & Vulnerabilities
- **The "MCP Top 10"**: OWASP is formalizing the "MCP Top 10" vulnerabilities, highlighting tool poisoning, overprivileged access, and unauthenticated discovery.
- **Supply Chain Attacks**: Documented cases of "MCP Server Hijacking" where malicious issues/PRs inject hidden instructions into agent workflows.
- **Identity Crisis**: Rising need for "Agent-to-Agent" (A2A) identity attestation to prevent session smuggling between swarms.

## Strategic Observations
- **Browser is the new Terminal**: Every major agent framework is now treating the browser as a first-class execution environment, necessitating a standardized MCP interface for browser-use.
- **Context Efficiency > Context Size**: Despite 1M+ token windows, the market is moving toward "Lazy Discovery" to reduce latency and "Context Poisoning" risks.
- **Zero-Trust is Mandatory**: Unauthenticated MCP servers (38% of current registry) are being actively exploited. "Safe-by-Default" is no longer optional.
