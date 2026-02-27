# Market Sync: 2026-02-27

## Ecosystem Updates

### Gemini CLI (v0.30.0)
- **SDK & Custom Skills**: Introduced an initial SDK package for dynamic system instructions and `SessionContext` specifically for SDK tool calls.
- **Policy Engine**: Major shift with the `--policy` flag and "strict seatbelt profiles," deprecating simple `--allowed-tools` in favor of a structured policy engine. This aligns with our Policy Firewall initiative.
- **Terminal Integration**: Improved Vim support and `Ctrl-Z` suspension, hinting at deeper local developer workflow integration.

### Claude Code & CoWork
- **MCP Tool Search**: Anthropic is emphasizing "Tool Search" for large tool sets, moving away from static tool pushing. This validates our "Lazy-Discovery" (P0) priority.
- **Sandboxed Execution**: Claude CoWork uses local containerized environments (Docker) to mount folders, highlighting the need for secure "Environment Bridging" between the container and the host.

### Agent Swarms & A2A (OpenClaw/Yarrow)
- **Filesystem-as-a-Bus**: Emerging patterns use the filesystem (JSON, Markdown, SQLite) for inter-agent communication instead of direct API calls. This ensures inspectability, Git-trackability, and crash-proof recovery.
- **Inspectable State**: The "Apothecary" model (Yarrow) treats agent swarms as a collection of specialized nodes communicating via a "Blackboard" on disk.

## Security & Vulnerability Landscape

### LangGrinch (CVE-2025-68664)
- **Criticality**: 9.3 CVSS.
- **Mechanism**: Prompt injection that extracts environment secrets and API keys through serialization pipelines in popular agent frameworks.
- **Impact**: Highlights a massive flaw in how structured data flows downstream; serialization must be hardened and "intent-aware."

### MCP Supply Chain (Clinejection)
- **Risk**: Malicious commands embedded in public repository issues/files can hijack developers' locally running AI agents if they read and process the content without a sandbox or strict egress policy.
- **Mitigation**: Requires cryptographic provenance (Attestation) for all MCP tools.

## Today's Unique Findings
1. **The Policy Engine War**: Google and Anthropic are racing to build "Policy Engines" rather than simple allowlists. MCP Any must provide a vendor-agnostic policy layer that can be injected into any of these tools.
2. **Virtual Filesystem Bus**: There is a growing demand for a "Virtual Filesystem" that MCP Any can provide to swarms, serving as both a communication bus and a secure, audited state store.
3. **Serialization Hardening**: We need to move beyond simple JSON-RPC and implement a "Hardened Schema Middleware" that sanitizes outputs before they reach the LLM or other subagents.
