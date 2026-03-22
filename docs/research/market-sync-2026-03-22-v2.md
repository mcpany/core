# Market Sync: 2026-03-22 (v2)

## Ecosystem Shifts
*   **NVIDIA OpenShell GA**: NVIDIA launched the NVIDIA Agent Toolkit including OpenShell, an open-source runtime for self-evolving agents. This signals a major entry of hardware-optimized infrastructure into the agentic layer, emphasizing "claws" with built-in safety.
*   **Claude Code "Agent Teams" Expansion**: Broad adoption of horizontal mesh coordination. Key patterns emerging include persistent workers with session memory to combat context-compression amnesia and the requirement for "Agentic Lock Managers" to handle concurrent file edits.

## Security Vulnerabilities & Vulnerability Trends
*   **CVE-2026-25253 (OpenClaw WebSocket Exfiltration)**: High-severity flaw where malicious websites could exfiltrate authentication tokens via the `gatewayUrl` parameter in the Control UI, leading to one-click RCE. Confirming that "Localhost Trust" is a primary attack vector.
*   **CVE-2026-0628 (Gemini Live Hijacking)**: Chrome extension vulnerability allowing hijacking of the Gemini Live panel. Enabled unauthorized camera/mic access, screenshots, and local file exfiltration. Highlights the risks of "Agentic Browsers" overstepping isolation boundaries.

## Autonomous Agent Pain Points
*   **Context-Compression Amnesia**: In long-running Claude Code teams, context compression leads to agents forgetting previous module interfaces, resulting in redundant file reads and token waste.
*   **Coordination Deadlocks**: Parallel agents editing the same files or waiting for interdependent tool outputs are creating "Recursive Deadlocks."

## Unique Findings
*   The move toward "Ghost Shell" profiling for un-attested hooks is gaining traction as a middle ground between "Implicit Trust" and "Hard Blocking."
