# Market Sync: 2026-04-06

## Ecosystem Updates

### 1. OpenClaw (v2026.3.22)
- **Marketplace Transition**: Replaced the vulnerable npm-based plugin system with **ClawHub**, a curated marketplace for agent skills. This addresses the "ClawHavoc" supply chain crisis.
- **Enhanced Sandboxing**: Implemented **OpenShell SSH Sandboxes** for all shell-based tools. Tool execution is now physically isolated from the host OS by default.
- **JVM Injection Protection**: Patched critical JVM injection paths used by malicious subagents to bypass early sandbox versions.

### 2. Gemini CLI (v0.34.0-preview)
- **Multi-Registry Architecture**: Subagents can now query multiple tool registries simultaneously. MCP Any should position itself as the primary "Enterprise Registry" in this mesh.
- **Subagent Local Execution**: Added support for isolated local execution of subagents, ensuring that child agents do not inherit the full capability set of the parent unless explicitly delegated.
- **A2A Acknowledgment**: New protocol command for agent-to-agent task synchronization, improving the reliability of deep swarm coordination.

### 3. Claude Code (v1.0.68+)
- **Refined Project Trust**: Enhanced permission checks for `allow/deny` tools and project-level trust. The `.claude.json` history field is now a protected structure.
- **SDK Callbacks**: Introduced `canUseTool` callbacks, allowing programmatic orchestration layers to intercept and approve tool calls in real-time.
- **Hook Governance**: Added `disableAllHooks` to neutralize malicious project-local configurations.

## Unique Findings & Pain Points
- **"Context Echoing"**: A new side-channel exploit where subagents use micro-timing of state updates to exfiltrate parent context fragments.
- **Registry Shadowing**: Malicious subagents are attempting to register "Shadow Tools" with identical names to high-trust tools in multi-registry environments (Gemini CLI vulnerability).
- **The "Delegation Gap"**: Swarms still struggle with 100% autonomous task completion when hardware-bound secrets (like TPM-locked keys) are required, leading to "approval fatigue" in human-in-the-loop flows.

## Strategic Match for MCP Any
- **Universal Registry Interop**: MCP Any can solve "Registry Shadowing" by acting as the **Authoritative Namespace Broker** for Gemini and OpenClaw swarms.
- **Hardware-Attested Intent Lineage (HAIL)**: bridging the "Delegation Gap" by providing hardware-signed proofs that a subagent's request is a direct descendant of a user-authorized mission root.
