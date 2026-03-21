# Market Sync: 2026-03-22

## Ecosystem Shifts & Competitor Analysis

### Claude Code: Stability and Security Hardening (v2.1.74)
*   **Update**: Anthropic released v2.1.74, fixing a critical failure where the sandbox could be silently disabled if dependencies were missing. It also hardened protection for `.git` and `.claude` directories.
*   **Context Optimization**: New `/context` suggestions identify memory bloat and context-heavy tools, signaling a move towards "Intelligent Context Budgeting."

### Universal Agent I/O Bus: The End of PTYs
*   **Shift**: Community consensus is moving away from 1970s-style PTY/terminal emulation for agents. The "Universal Agent I/O Bus" proposal treats commands and outputs as structured events (framed JSON) rather than raw byte streams.
*   **Impact**: Enables deeper inspection and sanitization of command outputs before they are ingested by the reasoning engine.

### Enterprise Visibility Crisis: The "Gravitee Report"
*   **Findings**: A new report indicates that while 80% of technical teams are deploying agents, only 21% of organizations have visibility into what their agents can access or which tools they are calling.
*   **Opportunity**: MCP Any can fill this "Governance Gap" by providing a standardized audit and visibility sidecar for all agent frameworks.

## Autonomous Agent Pain Points
1.  **Configuration-Based Redirection**: The discovery of CVE-2026-21852 (Base URL hijacking) highlights that agents are vulnerable to "Settings-as-Code" attacks in cloned repositories.
2.  **Context-Heavy Tooling**: Agents frequently "stall" or hallucinate when tools return un-optimized, large data structures that exceed context window efficiencies.

## Security Vulnerabilities (New)
*   **CVE-2026-21852**: Claude Code Information Disclosure. Malicious repository settings can override `ANTHROPIC_BASE_URL`, redirecting API requests (and keys) to attacker-controlled endpoints.
*   **Shadow-Sandbox Escape**: Confirmed patterns where agents write to protected directories if `bypassPermissions` mode is enabled, even if global sandbox settings are "on."

## Unique Findings
*   The transition from "Chat-with-Tools" to "Horizontal Mesh Coordination" (Agent Teams) is accelerating, requiring MCP Any to move from a single-agent proxy to a multi-teammate coordination bus.
