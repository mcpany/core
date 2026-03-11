# Market Sync: 2026-03-11

## 1. Ecosystem Shift: Command Injection Vulnerabilities in MCP Adapters
*   **Finding:** CVE-2026-0755 was disclosed, revealing a critical command injection vulnerability in `gemini-mcp-tool`. The exploit occurs in the `execAsync` function due to insufficient input validation when bridging Gemini CLI to system calls.
*   **Impact on MCP Any:** This reinforces the urgency of our "Policy Firewall" and "Input Validation" pillars. We must assume that upstream MCP servers might be poorly written and implement a "Zero-Trust Input Sanitization" layer.

## 2. Protocol Maturity: A2A (Agent-to-Agent) Emergence
*   **Finding:** Industry reports from OneReach and Ruh AI indicate that while MCP is the "USB-C for AI-to-Tool," A2A is becoming the standard for "Agent-to-Agent" coordination.
*   **Impact on MCP Any:** Our A2A Interop Bridge is no longer "experimental" but a "core requirement" for Enterprise Agent Swarms. We need to accelerate the A2A Stateful Residency feature to handle asynchronous handoffs.

## 3. Security Trend: "Safe-by-Default" & Attested Discovery
*   **Finding:** Following the "8,000 Exposed Servers" incident, the ecosystem is shifting towards local-only bindings by default and "Attested Discovery."
*   **Impact on MCP Any:** We must prioritize the "Safe-by-Default Hardening" and "Provenance-First Discovery" features to ensure MCP Any remains the most secure choice for developers.

## 4. Competitive Intelligence: Claude Code & Gemini CLI
*   **Finding:** Both Claude Code and Gemini CLI are increasingly relying on project-local configurations for tool "hooks." This is creating a massive surface area for RCE via malicious repository-level settings.
*   **Impact on MCP Any:** Our "Project Configuration Security Guard" (P0) is a unique differentiator that can protect users of these tools.
