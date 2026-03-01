# Market Sync: 2026-03-01

## 1. Ecosystem Shift: The OpenClaw Security Crisis
*   **Context**: OpenClaw (formerly Clawdbot) is facing a massive security meltdown. Over 335 malicious skills were discovered in its public marketplace (ClawHub), leading to RCE and credential theft.
*   **Vulnerabilities**:
    *   **CVE-2026-25253**: Remote Code Execution via unvetted skill execution.
    *   **CVE-2026-27001**: Indirect prompt injection via crafted environment context (e.g., malicious directory names).
*   **Implication for MCP Any**: We must prioritize **Supply Chain Attestation** and **Safe-by-Default** isolation. The "Marketplace" model for tools is dangerous without strict cryptographic provenance.

## 2. Competitive Intelligence: Anthropic Claude Code "Tool Search"
*   **New Feature**: Claude Code now implements "MCP Tool Search" (Lazy Loading).
*   **Mechanism**: If tool definitions exceed 10% of the context window, it switches from preloading to dynamic search-based loading via a `ToolSearchTool`.
*   **Implication for MCP Any**: Our "Lazy-MCP" middleware should align with this `ToolSearchTool` standard to ensure compatibility with Claude Code and other high-end clients.

## 3. Google Gemini CLI: Policy Engine & Session Context
*   **New Feature**: Gemini CLI v0.31.0 introduced project-level policies and tool annotation matching. They also added `SessionContext` for SDK tool calls.
*   **Implication for MCP Any**: Our Policy Firewall should support "Annotation-based" rules to match the Gemini ecosystem's granularity.

## 4. Key Agent Pain Points (GitHub/Reddit Trending)
*   **Context Bloat**: Users with 7+ MCP servers are hitting token limits (67k+ tokens just for schemas).
*   **Agent Hijacking**: Growing fear of agents being manipulated by untrusted data they "read" (e.g., summarizing a malicious webpage or file).
*   **Local-to-Cloud Gap**: Continued friction in bridging local files/tools to cloud-based agent runtimes securely.
