# Market Sync: 2026-03-05
**Status:** Final
**Author:** Lead Systems Architect

## 1. Ecosystem Shift: The "Great Exposure" Crisis
Recent reports from Trend Micro and Check Point have identified a massive security gap in the rapidly expanding MCP ecosystem.
*   **Exposed Infrastructure**: Trend Micro discovered 492 MCP servers exposed to the public internet with zero authentication. These servers often provide direct filesystem or shell access to the host machine.
*   **Skill Poisoning**: Antiy CERT confirmed 1,184 malicious skills across ClawHub (the OpenClaw marketplace), many designed to exfiltrate chat history or environment variables once installed by an autonomous agent.

## 2. Competitive Intelligence: OpenClaw & Claude Code
### OpenClaw "ClawJacked" Vulnerability
*   **Context**: A critical vulnerability (patched in v2026.2.25) allowed malicious websites to silently hijack a developer's OpenClaw agent without user interaction.
*   **Impact**: Full agent takeover via unauthenticated local port exposure.
*   **Implication for MCP Any**: We must prioritize "Safe-by-Default" local bindings and origin-locked sessions.

### Claude Code (v2.1.26–2.1.30)
*   **MCP Improvements**: Now supports MCP servers without dynamic client registration (e.g., Slack via OAuth).
*   **Sub-agent Tool Access**: Sub-agents can now access SDK-provided MCP tools, validating our "Recursive Context" strategy.
*   **Performance**: Significant RAM reduction (~68%) when resuming sessions, highlighting the need for lean state management in MCP Any.

## 3. Emerging Trends: Supply Chain Integrity
*   **Pentagon Designation**: Anthropic was designated a "supply chain risk" by the Pentagon, the first US AI lab to receive such classification, primarily due to the potential for "Agentic Injection" through the MCP protocol.
*   **Shift to Attestation**: The market is rapidly moving from "Any Tool" to "Attested Tooling."

## 4. Key Takeaways for MCP Any
1.  **Mandatory Auth**: Public IP exposure must trigger mandatory, non-bypassable authentication.
2.  **Origin-Locking**: Tie agent sessions to a verified host origin to prevent "ClawJacked"-style cross-site attacks.
3.  **Skill Sandboxing**: Implementation of a "Verified Skill Proxy" to scan and sanitize third-party MCP tool definitions before execution.
