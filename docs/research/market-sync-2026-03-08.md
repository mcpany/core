# Market Sync: 2026-03-08

## 1. Ecosystem Updates

### Gemini CLI & FastMCP Convergence
- **Integration**: Gemini CLI now natively supports FastMCP (v2.12.3+), allowing one-command installation of local STDIO servers (`fastmcp install gemini-cli`).
- **Impact**: Lowers the barrier for local tool integration but increases the risk of "unmanaged" local MCP servers being added without central oversight.

### Claude Code & Dynamic Tooling
- Claude Code is increasingly using dynamic tool search (MCP Tool Search) to handle the explosion of available tools.
- **Trend**: Shift from static tool configurations to "Just-in-Time" (JIT) tool discovery.

## 2. Security & Vulnerability Landscape

### OWASP MCP Top 10 (2025/2026 Drafts)
- **MCP04: Software Supply Chain Attacks**: Vulnerabilities in third-party MCP servers or their dependencies.
- **MCP09: Shadow MCP Servers**: The rise of unauthorized or unmanaged MCP servers running locally and being consumed by agents.
- **MCP10: Context Injection & Over-Sharing**: Agents leaking sensitive system context into tool calls or receiving malicious context from untrusted tool outputs.

### "Clawdbot" Post-Mortem Insights
- Recent findings suggest that subagent routing is vulnerable to "Intent Flow Subversion," where a malicious tool output can trick the parent agent into delegating tasks to a compromised subagent.

## 3. Emerging Patterns

### A2A (Agent-to-Agent) Maturity
- The A2A protocol is stabilizing around SSE (Server-Sent Events) for asynchronous task status updates, moving away from pure polling.
- **Pattern**: Agents are increasingly seen as "High-Level Tools" that require stateful handoffs.

## 4. Autonomous Agent Pain Points
- **Discovery Exhaustion**: LLMs are struggling with "Tool Overload" when presented with 50+ tools, leading to increased latency and hallucination.
- **Permission Fatigue**: Users are overwhelmed by granular permission prompts, leading to "Allow All" behaviors.
