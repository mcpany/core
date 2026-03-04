# Market Sync: 2026-03-04

## Ecosystem Shifts & Market Ingestion

### 1. Claude Code Security Crisis (CVE-2025-59536, CVE-2026-21852)
- **Findings**: Critical vulnerabilities discovered where malicious repository-level configuration files could hijack the Claude Code agent.
- **Attack Vector**: `CVE-2025-59536` allowed arbitrary command execution via MCP server definitions in untrusted repositories. `CVE-2026-21852` enabled API key exfiltration by overriding the `ANTHROPIC_BASE_URL` before user consent prompts.
- **Impact**: Highlights the danger of "Configuration-as-Execution" where tools automatically trust project-local settings.

### 2. Gemini CLI: Policy Engine & Project-Level Governance
- **Updates**: Gemini CLI v0.31.0 introduced project-level policies and MCP server wildcards.
- **Strategic Shift**: Deprecation of simple flags like `--allowed-tools` in favor of a robust, CEL/Rego-like policy engine for granular control.
- **Tool Discovery**: Added support for `notifications/tools/list_changed`, allowing dynamic tool updates without CLI restarts.

### 3. Agent-to-Agent (A2A) Protocol Maturity
- **Findings**: The A2A protocol is emerging as the industry standard for inter-agent delegation.
- **Adoption**: New frameworks are treating A2A as a first-class citizen, moving beyond simple Model-to-Tool interactions.

### 4. Autonomous Agent Pain Points
- **Supply Chain Integrity**: Users are increasingly wary of "shadow" MCP servers and untrusted tool configurations.
- **Context Pollution**: The "8,000 Exposed Servers" incident has shifted focus toward lazy-discovery and high-performance similarity search for tools to avoid bloating LLM context.

## Today's Unique Findings
- **The "Trust Gap"**: There is a significant gap between "cloning a repo" and "trusting its agentic configuration." MCP Any must bridge this by isolating project-local tools until explicit attestation is provided.
- **Dynamic Scoping**: Agents need the ability to "narrow" their own permissions based on the specific project they are working on, rather than having global access.
