# Market Sync: 2026-03-02

## Ecosystem Updates

### OpenClaw (v2026.2.26)
*   **Secrets Workflow**: Introduced a centralized secrets auditing and reloading mechanism, moving away from fragmented `.env` files.
*   **Browser Control Reliability**: Significant improvements to the Chrome extension connection and multi-agent browser orchestration.
*   **Sub-Agent Spawning**: Introduced deterministic sub-agent spawning directly from chat, allowing for explicit delegation rather than relying purely on model decision.
*   **Context Expansion**: Support for up to 1 million tokens, emphasizing the need for efficient context management.

### Gemini CLI (v0.31.0)
*   **Experimental Browser Agent**: Native support for interacting with web pages.
*   **Policy Engine Enhancements**: Support for project-level policies and tool annotation matching. This signals a shift towards more granular, context-aware security.
*   **SDK & Custom Skills**: New SDK for dynamic system instructions and `SessionContext` for tool calls.

### Claude Code (Security Vulnerabilities)
*   **Sandbox Escapes (CVE-2026-25725)**: Vulnerability where the absence of a `settings.json` file at startup allowed malicious code inside the sandbox to create it and inject persistent hooks.
*   **Trust Boundary Violations**: Multiple CVEs (CVE-2025-59828, CVE-2025-59536) regarding arbitrary code execution when running in untrusted directories before the "trust" prompt is shown.
*   **API Key Exfiltration (CVE-2026-21852)**: Redirection of `ANTHROPIC_BASE_URL` via project-level config to attacker-controlled endpoints.

## Agentic Pain Points
*   **Supply Chain Trust**: Cloned repositories can contain malicious configuration files (`.claude/settings.json`, `.yarnrc.yml`) that execute code before user consent.
*   **Stateful Multi-Agent Handoffs**: While sub-agent spawning is becoming deterministic, maintaining state and residency between intermittent connections remains a challenge.
*   **Browser-Agent Fragmentation**: Every major player (OpenClaw, Gemini, Claude) is building their own browser-use adapter, creating a need for a universal, secure browser-tooling standard.

## Findings for MCP Any
*   **Need for "Project-Scoped" Isolation**: MCP Any should not just trust a local directory. It needs to implement a "Shadow Config" or "Attested Config" loader that validates project-level settings against a global safety policy.
*   **Standardized Browser-Tooling Adapter**: Opportunity to provide a unified, secure interface for browser automation that works across different agents.
