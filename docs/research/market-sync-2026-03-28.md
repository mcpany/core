# Market Research Sync: 2026-03-28

## Ecosystem Updates

### Gemini CLI: Pre-Flight RCE Vulnerability
Security researchers have identified a critical trust-boundary mistake in Gemini CLI (v0.17.1). Workspace-controlled settings in `.gemini/settings.json`, specifically `tools.discoveryCommand`, were found to execute automatically during startup tool discovery even when the folder wasn't explicitly trusted.
- **Vulnerability**: "Unknown trust" state defaulted to trusted.
- **Impact**: Arbitrary command execution on the host machine when a user navigates into a crafted repository and runs the CLI.
- **Lesson for MCP Any**: Discovery-time execution must be treated as a high-risk event and isolated in a zero-trust sandbox.

### Claude Code: Agent Teams & Peer-to-Peer Coordination
Claude Code (v2.1.32) has introduced "Agent Teams" in research preview, shifting from a strictly hierarchical sub-agent model to a peer-to-peer coordination system.
- **Coordination Mechanism**: Agents communicate via a mailbox system located at `.claude/teams/<team_id>/inbox/`.
- **Locking**: Uses Git-based locking where agents claim tasks by writing to a shared directory.
- **Communication**: The `SendMessage` tool allows direct and broadcast messages, which are injected into the receiving agent's conversation history as user messages.
- **Opportunity for MCP Any**: Standardizing a secure, framework-agnostic mailbox broker for inter-agent communication.

### OpenClaw: Skill Permission Risks & ClawHub Malware
OpenClaw's rapid growth (100k+ stars) has exposed significant architectural risks in its "Skills" system.
- **Permission Model**: Every skill modular code package inherits the agent's system-wide permissions (full disk, terminal, network).
- **Supply Chain Attacks**: Coordinated campaigns like "ClawHavoc" have been identified on the ClawHub registry. Approximately 20% of the registry (900+ packages) was found to be malicious, distributing "Atomic Stealer" malware.
- **Exfiltration**: Malicious skills were found stealing browser passwords, keychain data, crypto keys, and SSH keys.
- **Strategic Pivot**: MCP Any must enforce granular, capability-based scoping for all tools and provide a verified registry with behavioral profiling.

## Autonomous Agent Pain Points
- **Discovery RCE**: startup-time execution of untrusted configurations.
- **Coordination Deadlocks**: synchronous locks in horizontal swarms.
- **Supply Chain Integrity**: unverified third-party code with system-wide access.
- **Runaway Costs**: autonomous loops exceeding token budgets without caps.
