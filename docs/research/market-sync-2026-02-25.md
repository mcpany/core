# Market Sync: 2026-02-25

## Ecosystem Updates

### Claude Code: MCP Tool Search (Lazy Loading)
- **Insight**: Anthropic released "MCP Tool Search" to combat "context pollution." Instead of loading all tool definitions at session start, Claude now discovers tools on-demand using similarity search (Regex/BM25).
- **Impact**: Reduces token overhead by ~85%, preserving context for actual work.
- **MCP Any Opportunity**: Implement a "Discovery Proxy" that provides a searchable index of thousands of tools without exposing them all to the LLM at once.

### Gemini CLI: FastMCP & Slash Commands
- **Insight**: Gemini CLI now supports FastMCP (v2.12.3+), which uses Pythonic decorators for rapid tool creation. It also integrates prompts as `/command` style slash commands.
- **Impact**: Lowers the barrier for developers to create and share "agent extensions."
- **MCP Any Opportunity**: Support FastMCP-style metadata and native "Prompt-to-Slash" command mapping.

### OpenClaw & "Clinejection" Supply Chain Attack
- **Insight**: A major supply chain vulnerability ("Clinejection") was exploited in February 2026. It combined indirect prompt injection and GitHub Actions poisoning to install unauthorized agents (OpenClaw) on developer machines.
- **Impact**: Huge focus on "Agent Supply Chain Integrity" and "Zero Trust Execution."
- **MCP Any Opportunity**: Strengthening the "Policy Firewall" and introducing "Provenance Verification" for MCP servers.

## Autonomous Agent Pain Points
- **Context Pollution**: Large MCP setups (e.g., GitHub + AWS) can consume >70% of the context window just for tool schemas.
- **Tool Selection Degradation**: LLMs struggle with accuracy when presented with too many tools (the "lost in the middle" problem for tools).
- **Unauthorized Lateral Movement**: Subagents often inherit too many permissions, leading to potential "lateral" exploits if one subagent is compromised via prompt injection.

## Security Vulnerabilities
- **Indirect Prompt Injection**: Still the primary vector for agent hijacking (as seen in Clinejection).
- **Cache Poisoning**: Exploiting CI/CD pipelines to inject rogue MCP configurations.
