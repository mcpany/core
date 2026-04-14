# Market Sync: 2026-04-14

## Ecosystem Updates

### OpenClaw
- **Version v2026.3.22 Overhaul**: Peter Steinberger's OpenClaw has transitioned from simple messaging to a robust model-agnostic infrastructure.
- **ClawHub Marketplace**: Replaced unregulated npm dependencies with a curated SDK. Over 5,000 skills published.
- **Security Hardening**: Implemented SSH sandboxing for shell tools and blocked JVM injection paths.
- **Star Power**: Surpassed React's GitHub star growth, reaching 250k+ stars in 60 days.

### Gemini CLI
- **Startup RCE Vulnerability (v0.17.1)**: Workspace-controlled `tools.discoveryCommand` in `.gemini/settings.json` executed automatically during startup even if the folder was untrusted.
- **Trust Boundary Failure**: RCA revealed that "unknown trust" defaulted to `true`, allowing arbitrary command execution when a user `cd`s into a malicious repo.

### Claude Code (Agent Teams)
- **Horizontal Coordination**: Transitioned from hierarchical subagents to peer-to-peer "Agent Teams".
- **Shared Task List**: Teammates use `TaskUpdate` to claim and status tasks.
- **Messaging Architecture**: `SendMessage` tool enables direct and broadcast communication via `.claude/teams/<team_id>/inbox/` local filesystem shards.
- **Shutdown Flow**: Formalized `shutdown_request` and `shutdown_response` handshake to prevent orphaned processes.

## Autonomous Agent Pain Points
- **Discovery-Time RCE**: The Gemini CLI incident proves that tool discovery itself is a high-risk event.
- **Coordination Deadlocks**: High-density teams struggle with "Mailbox Lock" congestion when using synchronous state.
- **Supply Chain Trust**: Even with ClawHub, the "Full disk, full network" permission inheritance of skills remains a primary concern.

## Security Vulnerabilities
- **CVE-2026-0628 (Ghost-Execution)**: Unauthorized code execution during the discovery phase via project-local configuration hooks.
- **Context-Window Flooding (CWF)**: New exploit pattern where malicious subagents inject high-entropy noise to evict mission-root instructions from the attention window.
