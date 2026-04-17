# Market Sync: 2026-07-25

## Ecosystem Shifts

### 1. Gemini CLI: "Settings-as-Shell" Vulnerability
Recent reports (Medium/Dhiraj Mishra, March 2026) reveal a critical design flaw in Gemini CLI's tool discovery phase. The CLI executes `tools.discoveryCommand` defined in `.gemini/settings.json` during startup. This allows malicious repositories to achieve Remote Code Execution (RCE) simply by convincing a user to run Gemini CLI within a compromised folder.
**Impact:** Confirms that the discovery phase is a high-risk execution environment that lacks proper sandboxing.

### 2. Claude Code: Agent Teams (v2.1.32+)
Claude Code has matured its "Agent Teams" feature, allowing multiple parallel sessions to coordinate via a shared git-based task list and mailbox system. Teammates can message each other directly (teammate-to-teammate) rather than just reporting to a lead.
**Key Pain Points:** Session resumption is unstable, and orphaned processes (tmux sessions) are common. Git-based locking for task claiming introduces significant latency and merge conflict risks in high-frequency coordination.

### 3. OpenClaw: ClawHub & SSH Sandboxing (v2026.3.22)
OpenClaw has officially migrated from npm-based skill distribution to **ClawHub**, a curated marketplace. This addresses the supply-chain risks of unregulated packages. Additionally, they've implemented native SSH sandboxing (OpenShell) for tool execution to contain RCE.
**Strategic Signal:** The industry is moving toward "Curated Skills" and "Default Sandboxing" as the standard for production agents.

## Autonomous Agent Pain Points
- **Instruction/Data Mixing:** The "von Neumann semantic bottleneck" where data processed by agents is interpreted as executable instructions (Indirect Prompt Injection).
- **Machine-Speed Insider Threats:** Agents inheriting real authority (API keys, infra access) but lacking hardware-bound identity and granular auditability.
- **Coordination Latency:** Synchronous locks in multi-agent meshes are becoming the primary performance bottleneck as swarms scale.

## Security Vulnerabilities
- **CVE-2026-25253 (ClawJacked):** Re-confirmed impact of unauthenticated local loopback trust.
- **Discovery-Time RCE:** Weaponization of pre-flight hooks in project-local configurations.
