# Market Sync: 2026-03-09

## Ecosystem Updates

### Gemini CLI (v0.30.0+)
- **Project-Level Policies**: Support for scoped policies that apply only to specific projects.
- **MCP Server Wildcards**: Simplified configuration for multiple MCP servers.
- **Tool Annotation Matching**: New policy engine capability to match tools based on metadata/annotations.
- **Interactive Shell Tool Calling**: Capability to execute tools that require interactive terminal sessions.

### Claude Code Security Disclosures
- **CVE-2025-59536**: RCE via malicious Hooks in `.claude/settings.json`.
- **CVE-2026-21852**: API Key exfiltration via crafted `ANTHROPIC_BASE_URL` in untrusted repositories.
- **The "Silent Helpfulness" Risk**: AI agents making small, passing-test changes that introduce long-term security backdoors.

### Agent Swarm Trends
- **OWASP Top 10 for Agentic Security (2026)**: Focus on "Indirect Prompt Injection" and "Tool Ecosystem Abuse."
- **Execution Boundary Hardening**: Movement towards "Zero Trust" execution where every tool call is verified against a high-level user intent.

## Unique Findings & Pain Points
1. **Hook Vulnerabilities**: Local hooks (pre-message, post-response) are becoming a primary RCE vector for developers working on untrusted repositories.
2. **Interactive Tooling**: Increasing demand for agents to handle interactive CLI tools (e.g., `git add -p`, `npm init`).
3. **Semantic Governance**: Names-based policies are failing; policies need to understand *what* a tool does (via annotations) to be effective.
