# Market Sync: 2026-03-12

## Ecosystem Shifts & Findings

### 1. OpenClaw: The Malicious Skill Crisis (ClawHavoc)
Recent reports indicate a major security crisis in the OpenClaw ecosystem. Attackers distributed over 300 malicious skills via ClawHub, the public marketplace. These skills, often with innocuous names, were designed to exfiltrate data and execute unauthorized system commands. This highlights the urgent need for a **Verified Skill Registry** and stronger provenance checks within MCP Any.

### 2. Claude Code: Critical Configuration Vulnerabilities
Two major CVEs have been identified in Claude Code's handling of project-local configurations:
- **CVE-2025-59536**: A bypass of the MCP consent mechanism where `.mcp.json` could override safeguards, allowing commands to execute without user approval.
- **CVE-2026-21852**: API key theft via `ANTHROPIC_BASE_URL` hijacking. Malicious repository-level settings redirect authenticated traffic to attacker-controlled servers.
These findings validate our pivot towards **Active Configuration Interception** and **Exfiltration-Resistant Transport**.

### 3. Gemini CLI: Connectivity and Proxy Friction
Analysis of Gemini CLI revealed significant user pain points regarding mandatory internet connectivity and broken proxy configurations. Specifically, `settings.json` proxy settings often fail during authentication, despite command-line flags working. There is a clear market gap for an **Offline-First Local Proxy** that can bridge air-gapped environments to the cloud via a hardened, well-configured gateway.

## Autonomous Agent Pain Points
- **Configuration-as-Execution**: Users are realizing that cloning a repo and opening an agent is now equivalent to running an untrusted script.
- **Context Pollution vs. Security**: Balancing the need for subagents to have context with the risk of state injection/exfiltration.
- **Trust Brushing**: The tendency for users to "trust all" when faced with numerous security prompts, necessitating "Safe-by-Default" automation.
