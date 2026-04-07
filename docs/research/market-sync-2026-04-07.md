# Market Sync: 2026-04-07

## Ecosystem Updates

### OpenClaw
- **Version 2026.3.22 Release**: Significant shift from unregulated npm packages to the curated **ClawHub marketplace**. This aligns with the industry-wide move toward supply chain integrity.
- **SSH Sandboxing**: Introduced native OpenShell SSH sandboxes for tool execution, moving away from host OS execution to mitigate RCE risks.
- **Reasoning Engine**: Shifted to GPT-5.40 as the default reasoning engine, indicating a demand for higher-order multi-step reasoning.

### Gemini CLI
- **A2A Authentication**: Introduced mandatory HTTP authentication for A2A remote agents.
- **Authenticated Agent Card Discovery**: Agent capabilities (Agent Cards) are now hidden behind an authentication layer, preventing "pre-flight shadow mapping" of a node's capabilities.

### Claude Code / MCP
- **Agentic Programming**: Emergence of "Agent Teams" where multiple Claude instances collaborate via MCP.
- **Sandbox Sovereignty**: Focus on resolving CVE-2026-25725 (sandbox escape) via deterministic environment integrity and "Non-Existence Proofs".

## Market Pain Points & Vulnerabilities
- **Supply Chain Poisoning**: High vulnerability rates (up to 87% in some reports) in agent-generated PRs and third-party skills.
- **"Invisible" Instructions**: Instruction injection via natural-language context files (e.g., `GEMINI.md`) or external metadata (GitHub issues).
- **Execution-as-Configuration**: Agents automatically ingesting hooks from project-local settings files, creating persistent RCE vectors.

## Summary of Findings
The "Local Trust" model is effectively dead. Leading frameworks are moving toward **Hardware-Attested Identity**, **Authenticated Discovery**, and **Isolated Execution Enclaves**. MCP Any must accelerate its transition from a "Connectivity Bridge" to a "Sovereignty Broker" that validates not just the connection, but the entire reasoning lineage and environment state.
