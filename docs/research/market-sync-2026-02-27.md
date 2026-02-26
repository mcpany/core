# Market Sync: 2026-02-27

## Today's Unique Findings

### 1. Claude Code Security & Capability Updates
*   **Vulnerabilities**: Anthropic addressed CVE-2025-59536 and CVE-2026-21852 in version 2.0.65. These flaws allowed malicious project configurations to execute arbitrary commands or steal credentials. This reinforces the need for MCP Any's **Policy Firewall**.
*   **Background Agents**: Claude Code introduced support for background agents and task-dependency management. This indicates a shift towards long-running, asynchronous tool execution.
*   **Remote Build Control**: New features for "remote-control" of external builds enabling local serving suggest a need for more robust environment bridging.

### 2. Gemini CLI v0.30.0 Release
*   **Policy Engine**: Introduced a new `--policy` flag for user-defined policies and "strict seatbelt profiles." Deprecated `--allowed-tools` in favor of this sophisticated policy engine.
*   **SessionContext**: Added `SessionContext` for SDK tool calls, allowing for better state management during multi-turn interactions.
*   **Custom Skills SDK**: Initial SDK package enabling dynamic system instructions and custom skills.

### 3. OpenClaw Security Crisis
*   The "OpenClaw Phenomenon" has led to a major security crisis due to the risks of autonomous shell execution and file system access. This has increased the market demand for "Zero Trust" execution environments and machine-checkable security contracts.

### 4. Emergence of A2A Protocol
*   The **Agent-to-Agent (A2A) Protocol** is gaining traction as the standard for inter-agent communication, complementing MCP. Standardized messaging (ACP) and distributed collaboration are key themes for 2026 multi-agent orchestration.

## Impact on MCP Any
*   **P0 Necessity**: The Gemini/Claude shifts toward "Policy Engines" validate our plan for a Rego/CEL based Policy Firewall.
*   **Async Support**: We must prioritize middleware that can handle background tool execution and status polling.
*   **A2A Bridging**: MCP Any must serve as the primary bridge between MCP-native tools and A2A-native agents.
