# Market Sync: 2026-02-27

## Ecosystem Shifts

### 1. OpenClaw Foundation Transition
Following Peter Steinberger's move to OpenAI, the OpenClaw project has officially transitioned to the "Agentic Open Source Foundation." The primary focus for Q1 2026 is the standardization of **Skill Manifests**—a portable JSON/YAML format for defining agent capabilities that can be shared across OpenClaw, CrewAI, and AutoGen. This directly impacts MCP Any's discovery layer.

### 2. Claude Code: The "Shadow Agent" Exploit
Security researchers have identified a new class of "Shadow Agent" exploits in Claude Code's subagent spawning mechanism. Subagents can sometimes "sidestep" the parent's `mcp.json` restrictions by initiating direct MCP connections to local services if the ports are discoverable. This has led to a call for **Isolated Transport** (moving away from local HTTP to Unix Domain Sockets/Named Pipes).

### 3. Gemini CLI "Deep Thinking" DAGs
Gemini CLI's latest update introduces a "Thinking Mode" that generates massive Directed Acyclic Graphs (DAGs) of tool calls. This has highlighted a major bottleneck in current MCP gateways: **Session State Persistence**. Long-running agent reasoning chains often lose state during high-latency tool calls, necessitating a more robust "Blackboard" or Shared KV store.

### 4. Agent-to-Agent (A2A) Maturity
The A2A protocol is seeing rapid adoption as the "glue" between specialized agents. There is an increasing demand for MCP Any to act as an **A2A Gateway**, allowing an agent in one framework (e.g., OpenClaw) to call an agent in another (e.g., AutoGen) as if it were a simple MCP tool.

## Key Pain Points
* **Port Exposure**: Local HTTP servers for MCP are being targeted by rogue subagents/scripts.
* **Context Loss**: Complex multi-step reasoning chains in Gemini are failing due to lack of shared state between tool calls.
* **Discovery Friction**: Manually configuring `mcp.json` for every new agent is becoming a major UX hurdle for power users.
