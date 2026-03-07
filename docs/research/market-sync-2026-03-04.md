# Market Sync: 2026-03-04

## Ecosystem Updates

### OpenClaw & Agent Swarms
- **OpenClaw 2026.2.17 Update**: Significant jump in context window support to 1 million tokens, allowing for massive repository-scale sessions.
- **Model Support**: Native integration with Claude Sonnet 4.6, delivering high performance for computer-use tasks and instruction following.
- **Multi-Agent Control**: Introduction of deterministic sub-agent spawning directly via chat commands, improving predictability in complex workflows.

### Claude Code & Anthropic Platform
- **MCP Tool Search**: Now in public beta, enabling agents to dynamically discover and load tools from massive catalogs on-demand, solving the "context pollution" problem for servers with 50+ tools.
- **Context Compaction**: Implementation of client-side compaction in Python and TypeScript SDKs, automatically managing conversation history through summarization when using `tool_runner`.

### Gemini CLI (v0.31.0)
- **Model Support**: Added support for Gemini 3.1 Pro Preview.
- **Experimental Capabilities**: Introduction of an experimental browser agent for web interaction.
- **Policy Engine**: Updates to support project-level policies and tool annotation matching.

## Security & Vulnerabilities

### Context Management Security
- The move toward dynamic tool discovery (MCP Tool Search) reduces the "Shadow Tool" attack surface by not loading all definitions upfront, but introduces new risks around "Search Injection" where malicious tool descriptions could hijack discovery.

## Autonomous Agent Pain Points
- **Discovery at Scale**: As tool catalogs exceed 100+ tools, the friction of manual configuration is being replaced by the complexity of search-based discovery.
- **Context Management**: Despite larger windows (1M tokens), efficient management (compaction/summarization) remains critical to maintain reasoning speed and reduce costs.
