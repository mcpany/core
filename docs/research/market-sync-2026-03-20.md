# Market Sync: 2026-03-20

## Ecosystem Shifts

### Claude Code: Path & Command Vulnerabilities
- **Path Validation Bypass**: Recent reports indicate that Claude Code's reliance on `startsWith()` for filesystem restriction is insufficient, allowing attackers to bypass directory boundaries by using carefully crafted paths that share an allowed prefix.
- **Command Injection**: Command validation using regex blacklists has proven fragile, leading to multiple command injection vulnerabilities discovered by third-party security firms.
- **Sandboxing Trends**: While moving toward Docker-based isolation, the lack of hardened security boundaries remains a systemic anti-pattern.

### Gemini CLI: Skill Activation Lifecycle
- **Skill Discovery**: Gemini CLI has formalized a discovery tier where skill names and descriptions are injected into the system prompt.
- **Activation Flow**: Implements a tool-based activation (`activate_skill`) with explicit user consent in the UI, detailing access scopes.
- **Context Injection**: Post-activation, Gemini injects the full `SKILL.md` and directory structure into the conversation history, highlighting the need for "Context-Aware Scoping" to prevent context window flooding.

### OpenClaw: Agent Teams & Governance
- **Horizontal Swarms**: Increased adoption of "Agent Teams" where coordination is decentralized but requires a shared "Blackboard" or state mailbox.
- **Autonomous Agent Pain Points**: "Cognitive Stall" in deep refinement loops and "Identity Spoofing" in heterogeneous meshes are the primary operational risks identified this week.

## Security & Vulnerability Findings
- **Agentic Social Engineering**: Malicious skills are increasingly using "Delayed Payload" tactics, appearing benign during the discovery phase and only executing malicious logic after gaining high-trust status.
- **Loopback Trust Gap**: The "ClawJacked" exploit pattern (CVE-2026-25253) continues to be relevant, as agents implicitly trust local WebSocket/HTTP traffic, allowing browser-to-local bridge attacks.

## Summary of Findings
The industry is shifting from "Connectivity-First" to "Trust-First" orchestration. The primary bottlenecks are no longer tool access, but the **integrity of the reasoning path** and the **sovereignty of the project-local environment**. Hardware-bound identity (TPM/SEP) and "Zero-Trust Discovery" are becoming the standard requirements for enterprise-grade agent infrastructure.
