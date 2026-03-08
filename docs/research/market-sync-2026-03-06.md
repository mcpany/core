# Market Sync: 2026-03-06

## Ecosystem Updates

### OpenClaw v2026.3.2 (Released 2026-03-03)
- **Native PDF Analysis**: Integrated support for Anthropic and Google PDF backends.
- **SecretRef Enhancements**: Credential references now cover 64 targets including runtime collectors and audit processes.
- **STT API**: New Speech-to-Text API for transcription.
- **Breaking API Changes**: `registerHttpHandler` replaced by `registerHttpRoute`, requiring explicit authentication declarations.
- **Security Hardening**:
    - WebSocket loopback hardening for gateways.
    - Symbolic link escape protection in skill workspaces.
    - Pre-authentication parsing for webhooks.

### Gemini CLI v0.32.0 (Released 2026-03-03)
- **Generalist Agent**: Enabled for improved task delegation and routing.
- **Model Steering**: Added workspace-level model steering.
- **Parallel Extension Loading**: Significant startup time improvements.
- **Policy Engine Updates**: Support for project-level policies, MCP server wildcards, and tool annotation matching.

### Claude Code GA (February 2026)
- **Generally Available Tools**: Code execution, web fetch, tool search, and memory tool are now GA.
- **MCP Tool Search**: Enabled by default; automatically defers tool descriptions to search if they exceed 10% of the context window.
- **Opus 4.6 Fast Mode**: Up to 2.5x faster token generation.

## Emerging Patterns & Pain Points
- **Autonomous Delegation**: The shift from single agents to "Generalist Agents" (Gemini) or "Swarms" (OpenClaw/Reddit) requires a universal delegation layer.
- **Dynamic Tool Discovery**: Claude's "MCP Tool Search" being default confirms that static tool schemas are a legacy bottleneck.
- **Hardened Multi-Tenant Security**: OpenClaw's move to explicit route authentication and symlink protection highlights the dangers of increasingly powerful agent workspaces.
