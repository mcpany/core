# Market Sync Research: 2026-03-02

## 1. Ecosystem Updates

### OpenClaw: Localhost Hijacking Vulnerability
*   **Finding**: Oasis Security reported a major vulnerability in OpenClaw where malicious websites could open a WebSocket connection to `localhost` on the agent's port.
*   **Exploit**: Since browsers don't block cross-origin WebSocket connections to localhost, malicious JS could brute-force passwords (exempt from rate limits on loopback) and register as a trusted device.
*   **Impact**: Full agent hijacking from a browser visit.
*   **Lesson**: "Localhost is trusted" is a dangerous assumption. MCP Any must implement strict origin validation and rate limiting even for loopback connections.

### Claude Code: Silent Hacking via Malicious Hooks
*   **Finding**: Check Point researchers found vulnerabilities in Anthropic's Claude Code where specially crafted configuration files in a cloned repo could execute arbitrary commands via "hooks".
*   **Exploit**: Claude Code requested approval for project files but NOT for hook commands defined in `.claudecode.json`.
*   **Impact**: Remote Code Execution (RCE) on developer machines during project initialization.
*   **Lesson**: All automated hooks or commands triggered by configuration files must require explicit, out-of-band user approval.

### Gemini CLI: Policy Engine & Browser Agent
*   **Update**: v0.31.0 released. Includes Gemini 3.1 Pro support and an experimental browser agent.
*   **Policy Engine**: Now supports project-level policies and MCP server wildcards.
*   **Observation**: The "Policy Engine" is becoming the standard for agent governance. MCP Any's "Policy Firewall" must remain a top priority to stay competitive.

### Agent Swarms (CrewAI, AutoGen, LangGraph)
*   **Trend**: The market is moving from "smart models" to "robust orchestration." State management and "Controllability" are the primary benchmarks for 2026.
*   **Observation**: MCP Any's role as a "Stateful Buffer" (A2A Stateful Residency) is highly aligned with the need for multi-agent mission persistence.

## 2. Autonomous Agent Pain Points
*   **"Local Trust" Fallacy**: Agents assuming local network environments are secure.
*   **Configuration Poisoning**: Using repo-level configs to bypass security prompts.
*   **Context Fragmentation**: Difficulty in maintaining a "Source of Truth" across 50+ specialized subagents.

## 3. Security Vulnerabilities
*   **Cross-Origin WebSocket Hijacking**: Specific to local-running AI agents.
*   **Supply Chain Hook Injection**: Malicious `.json` or `.yaml` configs in open-source repos.
