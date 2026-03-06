# Market Sync: 2026-03-06

## Ecosystem Shifts & Frontier Updates

### 1. Claude Code: Agent-as-a-Server (Recursive Agency)
*   **Discovery**: Claude Code now natively supports a server mode (`claude mcp serve`), allowing it to expose its high-level toolset (filesystem, bash, etc.) to other MCP clients.
*   **Impact**: This formalizes the "Recursive Agent" pattern where one agent delegates complex sub-tasks to another specialized agent instance via MCP. It bridges the gap between interactive CLI and programmatic tool use.
*   **Trend**: Shifting from "Model-to-Tool" towards "Agent-to-Agent" (A2A) orchestration.

### 2. The "PleaseFix" Vulnerability (Prompt Injection via Routine Content)
*   **Discovery**: Security researchers (Zenity Labs) identified a new attack vector where malicious instructions are embedded in routine data like calendar invitations or emails.
*   **Mechanism**: Agentic browsers (e.g., Perplexity Comet) automatically parse this content. Malicious prompts can override system instructions to exfiltrate local files or change account settings.
*   **Impact**: Proves that "Read-only" tool access is still dangerous if the agent performs "Autonomous Reasoning" on untrusted input data.

### 3. Gemini CLI: OAuth-First Integration
*   **Discovery**: New Gemini CLI MCP wrappers are prioritizing OAuth 2.1 over static API keys.
*   **Impact**: Reduces the risk of credential leakage in local configurations. Standardizes identity propagation between the user's local environment and the LLM service.

### 4. Market Projections
*   **Status**: MCP market projected to reach $10B by end of 2026.
*   **Drivers**: Enterprise adoption of "Contextual AI" and the need for a "USB-C for AI" standard.

## Autonomous Agent Pain Points
*   **Context Pollution**: Managing massive tool libraries (5,800+ servers) without overwhelming the LLM.
*   **Chain of Custody**: Lack of visibility into which subagent performed which action in a recursive chain.
*   **Input Sanitization**: Current gateways focus on *output* (tool results), but "PleaseFix" highlights a critical need for *input* (context) sanitization.

## Summary for MCP Any
MCP Any is perfectly positioned to solve these gaps by acting as the **Recursive Proxy** that manages A2A handoffs and the **Input Firewall** that sanitizes incoming context before it reaches the LLM.
