# Market Sync: 2026-04-13

## Ecosystem Shifts

### Gemini CLI v0.37.0 (2026-04-08)
- **Dynamic Sandbox Expansion**: Native support for Linux and Windows, improving isolated workflows.
- **Chapters Narrative Flow**: Introduction of "Chapters" for tool-based topic grouping, enhancing session structure.
- **Advanced Browser Capabilities**: Persistent sessions and dynamic tool discovery in browser agents.
- **Multi-Registry Architecture**: v0.36.0 introduced macOS Seatbelt and Windows sandboxing for subagent security.

### OpenClaw & Agent Security
- **Local WebSocket Vulnerability (VU#221883)**: A critical flaw in OpenClaw's local WebSocket gateway allowed malicious websites to hijack AI agents by exploiting implicit localhost trust.
- **Chrome Gemini Live Hijack (CVE-2026-0628)**: High-severity flaw allowing malicious extensions to hijack the AI assistant panel.
- **SlowMist "Agent-Facing" Defense**: Shift toward security guides designed to be read and deployed *by* the AI agent itself (Agentic Zero-Trust).

### Threat Landscape
- **Claude Code Supply Chain Attack**: Threat actors leveraged a packaging error in the `npm` release to distribute "Vidar" and "PureLog" stealers via fake GitHub repositories.
- **The "Lethal Trifecta" Realized**: Access to private data + Processing untrusted content + External communication is now a verified full-system compromise vector across multiple frameworks.
- **Autonomous AI Hackers (shannon)**: New agents like `shannon` achieving 96%+ success on security benchmarks, indicating that agents must now defend against other agents.

## Autonomous Agent Pain Points
- **Implicit Local Trust**: Localhost listeners remain the "soft underbelly" of agentic infrastructure.
- **Supply Chain Integrity**: Trusting third-party registries or packaging releases without hardware-attested provenance.
- **Context/Narrative Fragmentation**: Difficulty in maintaining "Narrative Flow" across complex tool-calling sessions (partially addressed by Gemini's "Chapters").
- **Agentic DoS & Resource Exhaustion**: The cost of autonomous "self-healing" or "refinement" loops becoming an economic attack vector.
