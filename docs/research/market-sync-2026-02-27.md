# Market Sync: 2026-02-27

## 1. Ecosystem Updates

### OpenClaw
- **Gateway Security**: OpenClaw's architecture relies on a Gateway process that connects channels and tools. Recent audits emphasize keeping gateways local-only until fully trusted.
- **Skill Safety**: "Skills" in OpenClaw are executable code, not just metadata. This has led to the discovery of malicious skills on ClawHub, highlighting a significant supply-chain risk.
- **Actionable Tooling**: Built-in deep security audits (`openclaw security audit --deep`) are now standard practice for verifying skill integrity.

### Claude Code (Anthropic)
- **MCP Tool Search**: Officially rolled out to handle "context pollution." Instead of loading all tool definitions, Claude now searches for and loads tools on-demand.
- **Vulnerability Discovery**: Claude Opus 4.6 has discovered 500+ high-severity zero-day vulnerabilities in open-source software, shifting the landscape toward AI-speed discovery and patching.
- **Security Patches**: Critical vulnerabilities (CVE-2025-59536, CVE-2026-21852) were recently patched. These exploited MCP server configurations and Hooks to achieve RCE and credential exfiltration.

### Gemini CLI (Google)
- **v0.30.0 Release**: Introduced an initial SDK for custom skills and `SessionContext` for SDK tool calls.
- **Policy Engine**: A new `--policy` flag and "strict seatbelt profiles" have been introduced, deprecating simpler allowed-tool lists in favor of a robust policy engine.
- **Terminal Integration**: Improved terminal suspension (Ctrl-Z) and Vim support, making it more robust for local developer workflows.

## 2. Autonomous Agent Pain Points
- **Context Pollution**: Large MCP server implementations (50+ tools) are exceeding LLM context windows, necessitating lazy-loading/search architectures.
- **Supply Chain Integrity**: The ease of installing third-party MCP servers/skills is being exploited for RCE and token exfiltration.
- **Cross-Environment State**: Bridging the gap between remote cloud sandboxes (like Anthropic's) and local filesystems/tools remains a high-friction area for developers.

## 3. Security & Vulnerability Trends
- **MCP Hook Exploits**: Attackers are targeting the "Hooks" and configuration mechanisms of MCP servers to execute arbitrary shell commands when a user clones/opens a malicious repository.
- **Identity & Attestation**: There is a growing need for cryptographic attestation of MCP servers to ensure they haven't been tampered with.
