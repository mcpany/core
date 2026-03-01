# Market Sync: 2026-03-01

## 1. Ecosystem Updates

### OpenClaw: The "ClawJacked" Crisis
- **Incident**: Researchers at Oasis Security disclosed **ClawJacked (CVE-2026-25253)**, a vulnerability chain allowing websites to hijack local AI agents.
- **Root Cause**: Misplaced trust in "local" connections. Attackers can use local HTTP tunneling to execute commands on the host.
- **Marketplace Risk**: Over 1,000 malicious "skills" were detected in the OpenClaw community marketplace, highlighting a massive supply chain risk for agentic tools.
- **Response**: OpenClaw 2026.2.23 introduced hardening, but the incident has shifted the industry focus toward "Local-Only by Default" and "Attested Tooling."

### Gemini CLI: v0.31.0 Policy Engine Evolution
- **New Release**: Google released Gemini CLI v0.31.0 (2026-02-27).
- **Policy Engine**: Major updates to the policy engine, now supporting **project-level policies**, **MCP server wildcards**, and **tool annotation matching**.
- **Agent Features**: Support for Gemini 3.1 Pro Preview and a new experimental browser agent.
- **Significance**: Gemini is doubling down on granular, attribute-based access control (ABAC) for tools, moving away from simple whitelisting.

### Claude Code: Configuration-Based RCE
- **Vulnerability**: Check Point Research detailed **CVE-2025-59536** and **CVE-2026-21852**, where malicious `.claude/settings.json` files in a repository could lead to Remote Code Execution (RCE) or API token exfiltration (via `ANTHROPIC_BASE_URL` hijacking).
- **Industry Impact**: This reinforces the danger of "Configuration-as-Code" without strict integrity checks. Even tool-specific config directories (like `.claude/` or `.vscode/`) are now primary attack vectors.

## 2. Autonomous Agent Pain Points
- **Reliability vs. Features**: Community sentiment (Reddit/GitHub) shows a pivot from "cool demos" to "production reliability." Users are demanding "fewer decision points and clearer failure modes."
- **Execution Governance**: High demand for permission boundaries and failure containment in autonomous workflows.
- **Context Pollution**: Large toolsets are still causing significant context window fatigue, increasing interest in lazy-loading and search-based tool discovery.

## 3. Strategic Implications for MCP Any
- **Hardening the Perimeter**: We must accelerate the "Safe-by-Default" hardening. Local connections are no longer inherently trusted.
- **Config Integrity**: MCP Any needs a mechanism to verify that configuration files haven't been tampered with or placed by a malicious third party (e.g., in a cloned repo).
- **Granular Scoping**: Aligning with Gemini CLI's project-level policies; MCP Any should support "Project-Aware" scoping where tools are only enabled if the agent is working within a verified project root.
