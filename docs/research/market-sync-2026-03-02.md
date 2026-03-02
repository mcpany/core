# Market Sync: 2026-03-02

## Ecosystem Shifts

### Gemini CLI 3.1 & Policy Engine Evolution
- **Gemini 3.1 Pro Preview**: Google has released the preview of Gemini 3.1, which includes an experimental **Browser Agent** for native web interaction.
- **Granular Governance**: Gemini CLI v0.31.0 introduced project-level policies and **Tool Annotation Matching**. This allows for more precise control over which tools an agent can call based on the metadata (annotations) of the tool itself.
- **SessionContext**: The new SDK supports `SessionContext` for tool calls, reinforcing the need for MCP Any to provide robust session-bound state.

### Claude Code & The Delegation Shift
- **Terminal & Cloud Delegation**: Claude Code is moving towards a model of total delegation where agents operate in both local terminal environments and isolated cloud sandboxes in parallel.
- **MCP as the "USB-C for Tools"**: The industry is coalescing around MCP as the standard transport layer, but concerns are mounting regarding the "Formalization of Access." A misconfigured MCP tool is now seen as a high-velocity exfiltration path.

### Security: The "Agent Hijacking" Crisis
- **Indirect Prompt Injection (IPI)**: This has moved from a theoretical threat to a primary attack vector. Malicious instructions hidden in data consumed by agents (e.g., a README or a web page) are being used to "hijack" the agent's tool access.
- **Execution Boundaries**: The security perimeter has shifted from the network level to the **Execution Boundary**. In 2026, defending an agent requires monitoring not just *what* it can access, but *how* it uses that access in response to untrusted input.

## Autonomous Agent Pain Points
- **Context Integrity**: Swarms of agents (like those in OpenClaw) struggle with "Intent Drift," where subagents lose the original security context or high-level goals of the parent, leading to unauthorized or anomalous tool calls.
- **Discovery Pollution**: As agents are given access to thousands of tools (via Lazy-Discovery), the "Semantic Collision" of tool names is causing models to call the wrong tool for a given task.

## Summary for MCP Any
MCP Any must prioritize **Autonomous Governance**. We cannot rely on users to manually define every policy. We need "Anomalous Tool Call Detection" that uses LLM-based reasoning to verify if a tool call aligns with the current session's intent.
