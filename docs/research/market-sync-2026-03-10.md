# Market Sync: 2026-03-10

## Ecosystem Updates

### Claude Code
- **Binary Content Handling**: Now saves binary outputs (PDFs, Audio, etc.) to disk instead of injecting base64 into context. MCP Any should support this "sidecar" file pattern.
- **Security Concerns**: Reports of automatic `.env` and secret ingestion. Highlights the need for our "Project Configuration Guard."

### Gemini CLI (v0.32.0 / v0.31.0)
- **Generalist Agent**: Improved task delegation. MCP Any's A2A Bridge should align with this delegation model.
- **Policy Engine**: Now supports project-level policies and tool annotation matching. MCP Any should aim for interop with these policy formats.
- **Experimental Browser Agent**: New tool types to support.

### OpenClaw & Swarms
- **ClawHavoc Incident**: Mass supply chain poisoning of ClawHub (OpenClaw skill registry). 20% of packages found to be malicious.
- **Exposure Crisis**: 135,000 instances found exposed with insecure defaults.

## Key Findings & Pain Points
1. **Tool SSRF**: 36.7% of MCP servers are vulnerable to SSRF. MCP Any must provide an egress proxy layer.
2. **Supply Chain Trust**: The ClawHavoc incident proves that dynamic discovery is dangerous without deep static analysis or sandboxing.
3. **Privacy Leaks**: Agents aggressively reading local project files (`.env`, `.git`) without user-intent verification.

## Strategic Opportunities
- **Egress Proxying**: Position MCP Any as the "Tool Firewall" that prevents tools from making unauthorized network requests.
- **Attestation Registry**: Develop a "Verified by MCP Any" stamp for tools that pass static analysis.
- **Contextual Privacy**: Implement "Just-in-Time" file access permissions.
