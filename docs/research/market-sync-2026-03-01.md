# Market Sync: 2026-03-01

## Ecosystem Shifts & Competitor Analysis

### 1. OpenClaw: Malicious Instruction Vulnerability
*   **Discovery**: A major security flaw (CVSS 8.7) was identified where OpenClaw agents can be "hijacked" by malicious websites or repositories. If an agent reads a malicious file or browses a compromised site, the site can inject instructions that the agent executes as if they came from the user (Remote Code Execution, API Key theft).
*   **Impact**: Demonstrates that "Provenance of Tool" is insufficient; "Provenance of Instruction" is the new frontier.
*   **MCP Any Opportunity**: Implement an "Intent-Lock" that requires a cryptographic handshake between the user's initial prompt and any high-risk tool call.

### 2. Gemini CLI: v0.31.0 & v0.30.0 Updates
*   **Policy Engine Expansion**: Now supports project-level policies and tool annotation matching. This moves beyond global allow-lists to context-sensitive security.
*   **SessionContext**: Introduced in v0.30.0. Allows the SDK to maintain state across tool calls without re-sending the entire context, aligning with our `Recursive Context Protocol`.
*   **Experimental Browser Agent**: Google is pushing native browser integration, increasing the risk of the "malicious order" exploit mentioned above.

### 3. Claude Code: Critical Exploits (CVE-2026-25725, CVE-2026-21852)
*   **Vulnerability**: Privilege escalation and RCE via malicious MCP servers and environment variable injection.
*   **Trend**: The supply chain for agent "skills" (MCP servers) is being actively targeted.
*   **MCP Any Opportunity**: Strengthen the `Supply Chain Integrity Guard` and `Safe-by-Default Hardening`.

### 4. Agent Swarms & Context Scaling
*   **Pain Point**: Context bloat remains the #1 performance bottleneck. `mcp-cli` demonstrated a 99% token reduction via dynamic discovery (lazy-loading schemas).
*   **Standardization**: Shift towards A2A (Agent-to-Agent) protocols to handle multi-agent refinement.

## Summary of Unique Findings for Today
*   **Instruction Provenance**: The industry has shifted from "Is this tool safe?" to "Is this instruction authorized?".
*   **Dynamic Tooling is Mandatory**: Agents can no longer afford to load 50+ tool schemas upfront; lazy-loading is the only way to scale to enterprise levels.
*   **MFA for High-Risk Tools**: As agents become more autonomous, the need for session-based MFA for sensitive actions (e.g., deleting a repo, transferring funds) is becoming a standard request in enterprise circles.
