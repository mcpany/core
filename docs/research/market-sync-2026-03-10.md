# Market Sync: 2026-03-10

## Ecosystem Shifts & Market Ingestion

### OpenClaw Security Crisis
* **Critical Vulnerabilities**: OpenClaw (formerly Clawdbot) is facing a multi-vector security crisis.
    * **CVE-2026-25253**: Remote Code Execution (RCE) via malicious hooks in project-local configs.
    * **CVE-2026-28453**: Path Traversal vulnerability in archive extraction.
* **Mitigation Strategy**: The community is moving towards `resolvePathWithinRoot` functions and path canonicalization. There's a strong push for isolated, sandboxed environments for tool execution.

### Agent Framework Dominance
* **Claude Code**: Now the most-used AI coding tool, overtaking GitHub Copilot and Cursor. 75% of developers at smaller companies use it as their primary tool.
* **MCP Integration**: Claude Code and Gemini CLI are heavily pushing MCP (Model Context Protocol). Gemini CLI has introduced an experimental policy engine for fine-grained tool call control.

### Autonomous Agent Pain Points
* **Confused Deputy Problem**: Attackers tricking trusted agents into executing malicious commands (e.g., via prompt injection or malicious config hooks).
* **Supply Chain Risks**: Malicious skills in marketplaces are becoming a major threat.
* **NIST Priorities**: NIST has stepped in to set security priorities for AI agents, focusing on standardized security layers.

## Unique Findings for MCP Any
* The "Local-Only by Default" stance of MCP Any is a significant competitive advantage given the "8,000 Exposed Servers" crisis.
* There is a desperate need for a "Path-Invariant" middleware that guarantees tools cannot escape their intended directory boundaries, regardless of the agent framework using them.
* "Intent-Aware" permissions are becoming the gold standard to prevent the Confused Deputy problem.
