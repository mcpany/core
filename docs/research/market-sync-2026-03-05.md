# Market Sync: 2026-03-05

## 1. Ecosystem Shifts & Findings

### OpenClaw "ClawJacked" Crisis
*   **Finding:** A critical multi-vector security crisis has hit OpenClaw. The most significant is the "ClawJacked" exploit chain, where a malicious website can take over a local AI agent by abusing localhost exposure (WebSocket + brute force). This collapses the security boundary between a browser tab and the local agent control.
*   **Impact:** Any agent running on a local port without strict origin/auth enforcement is vulnerable to RCE via a malicious website.

### Claude Code Configuration Exploits
*   **Finding:** Three critical vulnerabilities were identified in Claude Code.
    1.  **RCE via Hook Injection:** Manipulation of `.claude/settings.json` to inject malicious shell commands as "hooks."
    2.  **Consent Bypass:** Manipulating `.mcp.json` to override MCP consent mechanisms, allowing immediate command execution.
    3.  **API Key Theft:** Overriding `ANTHROPIC_BASE_URL` in project configuration to redirect API requests to an attacker-controlled server.
*   **Impact:** Repository-level configuration files (dotfiles) are becoming a primary attack vector for supply chain attacks in AI-integrated dev workflows.

### ModelScope MS-Agent Injection (CVE-2026-2256)
*   **Finding:** A vulnerability in ModelScope's MS-Agent enabled OS command execution via the Shell tool due to improper input sanitization.
*   **Impact:** Direct host compromise path via tool arguments.

### Agentic CI/CD Exploitation
*   **Finding:** "Hackerbot-claw" and similar AI-driven bots are actively exploiting GitHub Actions misconfigurations to get RCE and exfiltrate GITHUB_TOKENs.
*   **Impact:** Automation is being used to attack automation, turning CI/CD into an always-on exploit surface.

## 2. Autonomous Agent Pain Points
*   **Trust Boundary Collapse:** The transition from "Browser" to "Local Agent" is poorly protected.
*   **Shadow Configuration:** Repository-local config files are being treated as "trusted" by agents when they should be treated as "untrusted input."
*   **Sanitization Gaps:** Tools are often thin wrappers around shell commands without robust argument sanitization.

## 3. Summary for MCP Any
MCP Any must move beyond simple connectivity to become a **Security Hardened Perimeter**. The focus must shift to protecting the localhost boundary and verifying the integrity of any configuration that influences agent behavior.
