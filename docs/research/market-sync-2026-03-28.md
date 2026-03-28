# Market Sync: 2026-03-28

## Ecosystem Updates

### OpenClaw
- **Version 2026.3.22**: Transitioned from AI assistant to "AI Agent Platform".
- **ClawHub Marketplace**: Officially replaced npm as the primary skill source, hosting 4,000+ community skills.
- **Long-Running Missions**: Default timeout increased to 48 hours for complex batch jobs.
- **Remote Execution**: Introduced pluggable sandbox backends (OpenShell, SSH) for secure remote task execution.

### Gemini CLI
- **Version 0.35.0**:
    - `SandboxManager` introduced for isolated process-spawning tools (bubblewrap/seccomp).
    - JIT context discovery for filesystem tools.
- **Version 0.33.0/0.34.0 (Recent)**:
    - HTTP authentication for A2A remote agents.
    - Authenticated A2A agent card discovery.
    - Plan Mode enabled by default.
    - Native gVisor (runsc) and LXC support.

### Claude Code
- **Agent Teams**: Experimental feature `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`.
- **Collaborative Orchestration**: Move from isolated subagents to teammates sharing a "Shared Task List" and direct messaging.

## Autonomous Agent Pain Points & Vulnerabilities
- **Memory Injection (Sleeper Agents)**: Vulnerability where indirect prompt injection poisons long-term memory, creating persistent false security beliefs.
- **Uncontrolled Retrieval**: Agents inadvertently exposing PII or IP due to lack of semantic validation during RAG/retrieval.
- **Coordination Lock-in**: High latency and bottlenecks in multi-agent coordination when using synchronous locks.

## Unique Findings
- The shift from "isolated subagents" to "integrated teammate meshes" is now the dominant architectural pattern (Claude Code, OpenClaw).
- Security is moving from "per-call" to "environment-attested" and "long-term memory integrity".
