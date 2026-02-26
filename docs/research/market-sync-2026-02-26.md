# Market Sync: 2026-02-26

## Ecosystem Updates

### OpenClaw 2026.2.3: Security Hardening
- **Insight**: OpenClaw has reached 150k+ stars and is focusing heavily on "Safer, More Reliable Agents." Key updates include sandboxed file handling and prompt protection to mitigate the high risk of hijack via prompt injection.
- **Impact**: MCP Any must ensure its policy firewall can detect and block OpenClaw-specific injection patterns when proxying tools.

### Claude Code 2.1.50: Tool Search & Isolation
- **Insight**: Anthropic released Claude Code 2.1.50 with improved MCP tool discovery (tool search) and "worktree isolation" for agents. They also added a `claude agents` command to list configured agents.
- **Impact**: MCP Any's "Lazy-MCP" discovery must be compatible with Claude's tool search patterns. The "isolation: worktree" concept should be supported in our session management.

### Gemini CLI v0.29.0: Extension Discovery & Admin Controls
- **Insight**: Google introduced an extension registry and allowlisting for MCP server configurations. Sub-agents now have default execution limits.
- **Impact**: MCP Any can act as the "Admin Control" layer for Gemini, providing a unified allowlist across different agent frameworks.

### Rising Threat: AI Swarm (Hivenet) Attacks
- **Insight**: Security research (CrowdStrike, Kiteworks) highlights the emergence of "AI Swarm Attacks" where coordinated autonomous agents infiltrate systems simultaneously.
- **Impact**: MCP Any's role as a Zero-Trust gateway is more critical than ever. We need to implement multi-agent anomaly detection in the Policy Firewall.

## Autonomous Agent Pain Points
- **Inter-Framework Friction**: Integrating OpenClaw agents with Claude Code or Gemini CLI agents remains manual and error-prone.
- **Tool Schema Pollution**: Large toolsets still cause context window bloat despite improvements in individual frameworks.
- **Lack of Cost/Latency Visibility**: Agents often select expensive or slow tools because performance metadata is missing from the MCP schema.

## Security Vulnerabilities
- **CVE-2025-59536 / CVE-2026-21852**: RCE and API token exfiltration in Claude Code via malicious project-level settings.
- **Clinejection 2.0**: Supply chain attacks targeting MCP server registries and community-contributed tools.
- **A2A Identity Spoofing**: In multi-agent swarms, subagents can be tricked into accepting tasks from unauthorized "parent" agents.
