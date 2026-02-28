# Market Sync: 2026-02-28

## Ecosystem Shifts & Findings

### 1. The "ClawJacked" Crisis (CVE-2026-25253)
*   **Source**: Oasis Security / OpenClaw post-mortems.
*   **Finding**: A critical vulnerability chain in OpenClaw allowed websites to hijack local AI agents via unauthenticated local HTTP connections. The "misplaced trust in local" pattern is now a primary attack vector.
*   **Impact**: Mandates a shift from open local ports to authenticated transports (Named Pipes, Unix Domain Sockets with UID checks, or mTLS).

### 2. Claude Code Configuration Exploits (CVE-2025-59536)
*   **Source**: Check Point Research.
*   **Finding**: Remote Code Execution (RCE) achieved through malicious project-level configuration files (Hooks and MCP server definitions) that were executed without explicit user consent upon repository initialization.
*   **Impact**: MCP Any must treat "Project-Specific MCP Configs" as untrusted and implement a mandatory "Hook Sanitizer" and user-approval flow for any project-bound automation.

### 3. Gemini CLI v0.30.0 "Policy Engine" Pivot
*   **Source**: Google Gemini CLI Changelog.
*   **Finding**: Google is moving away from simple `--allowed-tools` flags in favor of a full "Policy Engine" with "Seatbelt" profiles and `SessionContext` for SDK calls.
*   **Impact**: MCP Any should align its policy architecture to be compatible with Gemini's declarative policy format to maintain its "Universal Adapter" status.

### 4. "Clinejection" & Supply Chain Integrity
*   **Source**: Snyk Research.
*   **Finding**: Indirect prompt injection in GitHub issue triage bots led to full supply chain compromise (unauthorized npm publishes).
*   **Impact**: Reinforces the need for **MCP Provenance Attestation** (P0) and **Intent-Aware Scoping** to prevent subagents from being hijacked into performing high-privilege actions (like publishing packages) via indirect prompts.

## Summary of Agent Pain Points
*   **Implicit Trust in 'Local'**: Users assume local services are safe; attackers are proving otherwise.
*   **Configuration as Code (CaC) Risk**: Project-specific tool definitions are being used for RCE.
*   **Prompt-Driven Privilege Escalation**: Agents are too easily tricked into using powerful tools for unintended purposes.
