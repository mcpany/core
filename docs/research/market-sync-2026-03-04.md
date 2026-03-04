# Market Sync: 2026-03-04

## Ecosystem Updates

### OpenClaw (v2026.2.17+)
- **Multi-Agent Orchestration**: Introduced deterministic sub-agent spawning and nested orchestration. Agents can now spin up specialized sub-agents for refinement tasks.
- **Security Crisis**: Growing reports of "ClawHub" malicious plugins. Discovery of 1,184 malicious skills.
- **Context Management**: Support for 1M token windows, but high latency remains a bottleneck for swarm coordination.

### Claude Code
- **Terminal Agent Maturity**: Claude Code is now the primary terminal interface for many developers.
- **RCE Vulnerability**: Check Point Research disclosed Remote Code Execution (RCE) via poisoned repository configuration files (e.g., malicious `.claudecode` configs).
- **Tool Discovery**: Moving towards on-demand "Skills" that are pulled from a central or local registry.

### Google Gemini CLI
- **Multimodal Tooling**: Strong integration with local file systems and multimodal inputs (images/PDFs) in the CLI.
- **MCP Native**: Standardizing on MCP for all external tool connections.

## Security & Pain Points

### The "Exposed Server" Crisis
- Trend Micro identified over 8,000 MCP servers exposed to the public internet with zero authentication.
- This creates a massive "Shadow AI" surface where agents can be tricked into calling unauthorized remote tools.

### Identity Crisis
- Agents are still largely using shared API keys. There is a lack of "Agent Identity" (Who is calling this tool? Was it the parent or the sub-agent?).

### Supply Chain Integrity
- The "Clinejection" and ClawHub incidents show that the MCP supply chain is the new primary attack vector.

## Unique Findings for MCP Any
- **Requirement for Config Sandboxing**: We must ensure that MCP Any does not trust repository-level configurations by default to prevent Claude Code-style RCE.
- **Agent-to-Agent (A2A) Residency**: As swarms become nested (OpenClaw style), MCP Any needs to act as the "Local Bus" that holds the state for the entire tree of agents.
