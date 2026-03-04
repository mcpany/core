# Market Sync: 2026-03-04

## Ecosystem Updates

### OpenClaw: Universal State Handover (USH)
OpenClaw has introduced a preliminary specification for **Universal State Handover (USH)**. This allows subagents to package their current local state (variables, partial results, and execution stack) and hand it off to another subagent or parent agent via a standardized MCP-compatible header. This reduces the need for parent agents to "re-summarize" state before every delegation.

### Claude Code: Tool-Output Sanitization Layer
Anthropic's Claude Code has added a native **Sanitization Layer** for tool outputs. It uses a small, high-speed model to summarize or truncate large tool returns (like 5000-line log files) before they hit the main LLM's context window. This significantly reduces "Context Poisoning" and token costs.

### Gemini CLI: Native Slash Mapping
Gemini CLI now supports direct mapping of MCP tool names to native slash commands (e.g., `/mcp.github.create_issue`). This simplifies the local execution UX, making MCP tools feel like first-class CLI features.

## Autonomous Agent Pain Points

1.  **Context Bloat (The "Log Dump" Problem)**: Agents are still struggling with tools that return too much data, leading to rapid context window exhaustion and increased costs.
2.  **Session Fragmentation**: When a task moves from Agent A (Research) to Agent B (Coding), the fine-grained execution context is often lost, leading to Agent B repeating work or hallucinating missed steps.
3.  **Discovery Latency**: In large agentic swarms, finding the *right* tool among hundreds of available MCP servers is becoming a performance bottleneck.

## Unique Findings for Today
- There is a growing trend towards "Lazy-Loading" of agent state, where the full state is only fetched if the receiving agent explicitly requests it (State-on-Demand).
- Security focus is shifting from "Is this tool safe?" to "Is the *data* returned by this tool safe to ingest?".
