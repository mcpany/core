# Market Sync: 2026-03-05

## 1. Ecosystem Updates

### OpenClaw: The "ClawJacked" Vulnerability
*   **Source**: OASIS Security Research (Mar 3, 2026).
*   **Finding**: A vulnerability chain allowed any website visited by a developer to silently take full control of their local OpenClaw agent.
*   **Impact**: High. Enabled unauthorized command execution, message sending, and data exfiltration.
*   **Resolution**: Fixed in version 2026.2.25.
*   **Implication for MCP Any**: Reinforces the "Safe-by-Default" directive. Local-only bindings are not enough if the web dashboard is vulnerable to CSRF/XSS-style hijacking of the agent's capabilities.

### Gemini CLI: v0.32.0 Release
*   **Key Features**:
    *   **Generalist Agent**: Improved task delegation and routing logic.
    *   **Model Steering**: Native support for steering models within the workspace.
    *   **Adaptive Planning**: Users can modify plans in external editors; multi-select for complex tasks.
*   **Implication for MCP Any**: MCP Any must support "Delegation Metadata" to help generalist agents understand the trust level of specific sub-tools.

### Claude Opus 4.6 & Developer Platform
*   **Adaptive Thinking**: Claude now decides autonomously when to engage in deep reasoning.
*   **Automatic Caching**: Messages API now supports automatic cache point progression.
*   **Implication for MCP Any**: Resource telemetry should include "Thinking Time" or "Reasoning Tokens" to help users understand the cost/latency trade-offs of adaptive thinking.

## 2. Trending Pain Points & Security
*   **"Shadow Skills"**: Continued reports of malicious or unverified skills being injected into agent swarms.
*   **Context Fragmentation**: As agents move between cloud sandboxes (Claude Code) and local environments (OpenClaw), maintaining a consistent "Source of Truth" for tool state is a major friction point.
*   **Zero-Trust Delegation**: Increasing demand for "One-Time-Use" tool tokens that expire after a specific sub-task is completed.

## 3. Summary for Strategic Alignment
The shift is moving from "How do I connect this tool?" to "How do I securely delegate this task to another agent?" MCP Any must pivot to be the **Verification & Delegation Layer** for agentic swarms.
