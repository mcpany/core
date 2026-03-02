# Market Context Sync: 2026-03-02

## 1. Ecosystem Updates

### OpenClaw (v2026.2.17)
*   **Multi-Agent Nesting**: Introduced deterministic sub-agent spawning and nested orchestration. This allows for complex, hierarchical agent swarms where a parent agent can delegate to specialized sub-agents with high reliability.
*   **Skill Executability**: Emphasis that "Skills" are executable code, not just plugins, increasing the surface area for supply-chain attacks (e.g., malicious skills on ClawHub).
*   **Security Audit Tooling**: New `openclaw security audit --deep` command to identify over-privileged configurations.

### Gemini CLI (v0.31.0)
*   **Policy Engine Evolution**: Added support for project-level policies, MCP server wildcards, and tool annotation matching. This allows for more granular and scalable permission management.
*   **Gemini 3.1 Pro Support**: Integration with the latest model, emphasizing faster reasoning for tool selection.
*   **SessionContext**: Improved handling of state across SDK tool calls.

### Claude Code
*   **Tool Discovery Fixes**: Resolved bugs where MCP tools were not discovered when tool search was enabled in specific launch modes.
*   **Memory Management**: Improved cache compaction and clearing of large tool results to maintain performance in long-running agentic sessions.

## 2. Emerging Patterns & Pain Points

### "Time-to-Trust" Framework (CSA 2026)
*   A proposed standard for progressive permissions. Agents start with "Probation" (read-only, no network) and escalate to "Junior" (limited write/network) and finally "Senior" (full access) based on incident-free duration and manual attestation.

### The "Vibe-Code" Vulnerability
*   Security researchers (Snyk 2026) identified a trend where developers grant "sudo" or "admin" access to agents because it "vibes" (works faster/easier), bypassing critical system safety boundaries in favor of model safety.

### Autonomous Agent Persistence
*   Trending GitHub projects (e.g., `adk-go`, `Memori`) are focusing on solving task interruptions and state persistence, which remains a primary blocker for truly autonomous workflows.

## 3. Strategic Implications for MCP Any
*   **Policy Granularity**: We must match or exceed Gemini CLI's wildcard and annotation-based policy support.
*   **Progressive Security**: The "Time-to-Trust" model should be a first-class citizen in our Policy Firewall.
*   **Deterministic Orchestration**: Our A2A bridge needs to support the deterministic handoffs seen in OpenClaw.
