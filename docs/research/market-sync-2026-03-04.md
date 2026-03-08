# Market Sync: 2026-03-04

## Ecosystem Updates

### OpenClaw: The Localhost Security Crisis
*   **Vulnerability (CVE-2026-25253):** A critical flaw in OpenClaw (v2026.1.29 and earlier) allowed remote websites to hijack AI agents via cross-site WebSocket hijacking (CSWSH).
*   **Key Insight:** Because browsers do not block WebSocket connections to `localhost` via standard CORS/SOP, a malicious website could interact with an agent's local RPC/WebSocket port.
*   **Impact:** Over 21,000 exposed instances identified. This confirms that binding to `localhost` is NOT a sufficient security boundary for agentic tools.

### Gemini CLI & SDK (v0.30.0 - v0.31.0)
*   **SessionContext:** Introduced a native SDK mechanism for maintaining state across tool calls. This aligns with our "Recursive Context Protocol" but focuses on the SDK-to-Model layer.
*   **Policy Engine Maturation:** Now supports project-level policies and MCP server wildcards. The move from `--allowed-tools` to a formal Policy Engine mirrors our "Policy Firewall" strategy.
*   **Custom Skills:** Initial SDK package enables dynamic system instructions and custom skills, increasing the need for standardized agent-to-agent (A2A) handoffs.

### GitHub & Reddit Trending: Persistence & Coordination
*   **Persistent Memory:** Top trending projects (adk-go, Memori) are focusing on solving task interruptions.
*   **Cross-Framework Gaps:** Coordination between different agent frameworks (e.g., CrewAI vs. AutoGen) remains a primary unsolved pain point for 2026.

## Unique Findings for MCP Any
1.  **"Localhost-Only" is Obsolete:** We must implement Origin validation (Host header/Origin header check) or move to Docker-bound named pipes/Unix sockets to prevent browser-based hijacking.
2.  **SDK-Native Session Parity:** MCP Any should provide a bridge that maps our `Recursive Context` headers to Gemini's `SessionContext` and similar SDK-native state objects.
3.  **Policy Wildcards:** Our Policy Firewall must support wildcarding for MCP server names to match the ease-of-use seen in the latest Gemini CLI updates.
