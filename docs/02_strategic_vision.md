# Strategic Vision: MCP Any

## The Universal Agent Bus
MCP Any is positioned as the indispensable core infrastructure layer for all AI agents, subagents, and swarms. By providing a configuration-driven gateway that bridges existing APIs with the Model Context Protocol, we enable a "Plug and Play" ecosystem for agentic capabilities.

## Core Pillars
1. **Decoupling**: Protocol-agnostic adapters for HTTP, gRPC, CLI, and Filesystems.
2. **Security**: Zero Trust architecture with granular policy enforcement.
3. **Observability**: Real-time tracing and health monitoring of all agent-tool interactions.
4. **Recursive Coordination**: Standardized protocols for agent-to-agent context passing.

## Strategic Evolution: [2026-02-22]
### Shift to Recursive Context Protocols
The rise of local swarms (OpenClaw) and developer-centric MCP clients (Claude Code) necessitates a transition from "Simple Gateway" to "Universal Agent Bus".
- **Standardized Context Inheritance**: Implementing a header-based inheritance system to ensure subagents operate within the same security and preference bounds as their parents.
- **Shared State coordination**: Providing a "Blackboard" tool for agents to share state without polluting the main LLM context window.
- **Zero Trust Local Execution**: Moving towards sandboxed execution for command-based skills to mitigate shell access risks identified in today's research.
