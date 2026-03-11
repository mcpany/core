# Market Sync: 2026-03-11

## Ecosystem Updates

### OpenClaw (CVE-2026.2.25)
* **Finding**: A critical high-severity vulnerability was patched in OpenClaw. It allowed malicious websites to hijack a developer's local AI agent by failing to distinguish between local trusted connections and untrusted browser-originated requests.
* **Impact**: Potential for RCE, data exfiltration, and unauthorized tool execution via the local agent.
* **Implication for MCP Any**: Re-affirms the "Safe-by-Default" priority. We must implement cryptographic origin validation (attestation) for all incoming requests, even those appearing to be from `localhost`, to prevent cross-origin hijacking.

### Claude Code "Remote Control" & Repo-Driven Attacks
* **Finding**: Claude Code's "Remote Control" bridges the cloud UI to local execution. While data stays local, it amplifies the impact of "repo-driven" attacks where malicious `.claude/settings.json` or MCP server configurations in a repository can be exploited if the user is distracted or "rubber-stamps" approvals.
* **Impact**: Trusting a malicious repo can lead to the agent executing harmful hooks or connecting to rogue MCP servers.
* **Implication for MCP Any**: The "Project Configuration Security Guard" is now a P0 requirement. We need to go beyond simple validation and implement a "Reputation/Provenance" check for project-local tools.

### AI Agent Security Landscape 2025-2026
* **Finding**: The "8,000 Exposed Servers" crisis and "LangGrinch" (credential exfiltration in LangChain) have shifted the market towards "Identity-First" agent security.
* **New Pain Point**: "Skill/Plugin Supply Chain" risks. Malicious "skills" in OpenClaw's ClawHub marketplace were found stealing browser data and credentials.
* **Autonomous Agent Pain Points**: "Context Pollution" in multi-agent swarms and "State Injection" where one subagent can maliciously alter the shared "Blackboard" to influence other agents.

## Strategic Takeaways
1. **Move to "Attested Discovery"**: Tool discovery must move from "is this tool available?" to "is this tool trusted and by whom?".
2. **Cross-Environment Attestation**: As agents move between cloud (Claude Code), local (OpenClaw), and edge, the security token must carry "Contextual Provenance" (where did this intent originate?).
3. **Hardened Local Gateway**: The local MCP gateway must be invisible to the browser unless a cryptographic handshake is performed.
