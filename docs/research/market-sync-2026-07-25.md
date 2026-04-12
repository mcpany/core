# Market Sync: 2026-07-25

## Ecosystem Updates

### OpenClaw (v2026.3.22)
- **Plugin Overhaul**: Replaced unregulated npm dependencies with the curated **ClawHub marketplace**.
- **Reasoning Engine**: GPT-5.40 introduced as the default reasoning engine.
- **Security Sandboxing**: Implemented **OpenShell SSH Sandboxes** to prevent host-level RCE, moving away from direct host execution.
- **Discovery Flow**: Transitioned to a more robust, model-agnostic agentic infrastructure.

### Claude Code & Agent Teams
- **Security Vulnerability**: A critical vulnerability in the permission system was disclosed by Adversa AI. A crafted `CLAUDE.md` file can generate a pipeline of subcommands that bypasses deny rules, security validators, and command injection detection.
- **Malware Campaigns**: "Leaked" versions of Claude Code on GitHub are serving Vidar infostealers and GhostSocks proxies, exploiting the curiosity surrounding the tool.
- **Context Hijacking**: Attackers are using invisible project-local instructions in Markdown files to trick agents into executing unauthorized exfiltration commands.

### Gemini CLI
- **Skill Discovery**: Matured the discovery tiers, where skills are injected into the system prompt upon activation.
- **Tool Discovery Patterns**: Using `/skills` list and activation flows that require user consent before SKILL.md body injection.

## Discovery Phase Pain Points
- **Implicit Trust**: Local loopback and project-local files are being weaponized to bypass sandbox boundaries.
- **Discovery-Time Execution**: Discovery commands are high-risk events that can lead to host-level "Ghost-Execution".
- **Semantic Drift**: As swarms become more complex, maintaining the atomic integrity of shared state across teammates is a primary bottleneck.

## Security Vulnerabilities
- **CVE-2026-25253**: Token exfiltration in local listeners due to flawed loopback trust.
- **CVE-2026-25725**: Sandbox escape in Claude Code via environment-binding racing.
- **CLAUDE.md Bypass**: Silent exfiltration of SSH keys and AWS credentials via crafted pipeline commands.
