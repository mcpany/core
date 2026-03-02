# Market Sync: 2026-03-02

## Ecosystem Updates

### Gemini CLI (v0.31.0)
- **Tool Annotation Matching**: Gemini CLI now supports matching tools based on annotations. This suggests a shift towards metadata-driven tool selection and policy enforcement.
- **Project-Level Policies**: Introduced support for granular policies at the project level, including MCP server wildcards.
- **Experimental Browser Agent**: New capabilities for web interaction, increasing the surface area for autonomous tool use.

### Claude Code
- **Headless Mode**: Official support for non-interactive execution (`claude -p "..." --output-format stream-json`). This highlights a growing need for agents to operate in CI/CD pipelines without human intervention.
- **Auto-Accept Patterns**: The `-y` or `--dangerously-skip-permissions` flags indicate a user preference for "Trusted local dev" environments, which must be balanced with safety in "Universal Agent" contexts.

### OpenClaw & Agent Swarms
- **Autonomous Agent Pain Points**: Scaling tool discovery remains a bottleneck. Agents are struggling with "Context Pollution" when 100+ tools are exposed simultaneously.
- **Security Vulnerabilities**: Recent "Clinejection" patterns show that even local-only agents are vulnerable to indirect prompt injection via malicious tool schemas or data.

## Unique Findings for MCP Any
- **The "Headless Attestation" Gap**: While Claude Code offers a "dangerously-skip" flag, a Universal Gateway like MCP Any must provide a *secure* way for headless agents to attest their identity without compromising the "Safe-by-Default" local-only posture.
- **Annotation-Driven Routing**: There is a clear opportunity to implement a middleware that routes tool calls based on Gemini-style annotations, allowing for more dynamic and context-aware policy enforcement.
