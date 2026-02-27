# Market Sync: 2026-02-27

## Ecosystem Updates

### Gemini CLI v0.30 (Released 2026-02-25)
*   **Policy Engine First**: Introduced the `--policy` flag for user-defined policies and strict seatbelt profiles. Deprecated `--allowed-tools` in favor of a full Policy Engine.
*   **SessionContext SDK**: New SDK enabling `SessionContext` for tool calls, allowing for better state management within custom skills.
*   **Admin Controls**: Administrators can now allowlist specific MCP server configurations, signaling a shift toward enterprise-managed tool access.

### Claude Code: Remote Control (Released 2026-02-24)
*   **Local-First Cloud Bridging**: Anthropic introduced "Remote Control," allowing users to start a session locally and control it via mobile/browser without moving code to the cloud.
*   **Agent Teams**: Claude Code now supports running multiple coding agents in parallel with shared history and repository instructions.

### GitHub Copilot: Multi-Model Agency (Released 2026-02-26)
*   **Agent Control Plane**: Generally available for Copilot Business, providing centralized enablement and audit logging for multiple agents (Claude, Codex, Copilot).
*   **Shared Memory/Context**: Agents now operate on a shared platform with unified governance and "Copilot Memory."

## Security & Vulnerability Trends

### OWASP Top 10 for Agentic Applications (2026)
*   **ASI01: Agent Goal Hijack**: Manipulating agent decision pathways through indirect injection (e.g., hidden payloads in documents/emails).
*   **ASI07: Insecure Inter-Agent Communication**: Weak authentication or semantic validation between agents leading to unauthorized commands or spoofing.
*   **ASI03: Identity and Privilege Abuse**: Agents operating with over-privileged access or lack of unique, scoped identities.

## Strategic Implications for MCP Any
*   **Policy Engine Alignment**: MCP Any's Policy Firewall must match or exceed the capabilities of Gemini's new Policy Engine, specifically supporting Rego/CEL for complex "seatbelt" profiles.
*   **Identity as a Perimeter**: With the rise of "Agent Teams" and multi-agent coordination, tool calls must be bound not just to a "Session" but to a verified "Agent Identity" (Attestation).
*   **Remote-to-Local Secure Proxy**: The "Remote Control" pattern highlights the need for MCP Any to act as a secure gateway that bridges remote UI/Agents to local tools without exposing the host filesystem directly.
