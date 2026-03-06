# Market Sync: 2026-03-02

## Ecosystem Shifts & Findings

### 1. Claude Code Security Crisis & Vulnerabilities
Recent disclosures by Check Point researchers have exposed critical flaws in the Claude Code ecosystem that directly impact the design of "Safe-by-Default" infrastructure.
*   **CVE-2025-59356 (Malicious Hooks):** Bad actors can inject malicious commands into a project's `.claudecode/config` via "Hooks." These execute automatically when a developer opens the project, leading to full terminal compromise.
*   **CVE-2025-59536 (MCP Pre-Auth RCE):** MCP servers can be configured to execute commands before the user is even prompted with a security warning.
*   **CVE-2026-21852 (Credential Harvesting):** Adversaries can intercept API-related communications by altering configuration files, leading to silent API key theft.
*   **Takeaway:** MCP Any must implement **Sandboxed Hook Execution** and **Configuration Change Attestation** to prevent similar exploits.

### 2. OpenClaw & Modular Agent Refinement
OpenClaw continues to gain traction as the leading self-hosted alternative, emphasizing modular "Skills" and integrated vector/FTS5 search.
*   **Skill Modularity:** Their approach to modular code packages as skills suggests MCP Any should further simplify the "Tool-to-Skill" mapping.
*   **Local-First Indexing:** OpenClaw's immediate index updates on file changes set a benchmark for our "Lazy-Discovery" feature's performance.

### 3. OWASP MCP Top 10 & Industry Hardening
The emergence of the "OWASP MCP Top 10" highlights the shift from "How do we build agents?" to "How do we secure agents?".
*   **Top Risks:** Tool poisoning, RCE via argument injection, and overprivileged access are the primary concerns for 2026.
*   **The "8,000 Exposed Servers" Problem:** Scans show nearly 40% of production MCP endpoints lack protocol-level authentication.

### 4. Agent-to-Agent (A2A) Maturity
A2A is moving from a concept to a production requirement. The "Agent Swarm" architecture (e.g., OpenClaw's multi-agent refinement) requires a stateful bridge.
*   **Pain Point:** Intermittent connectivity between specialized subagents often leads to state loss.
*   **Requirement:** A "Stateful Residency" or "Mailbox" model where the gateway (MCP Any) holds the message until the recipient agent is ready.

## Summary of Autonomous Agent Pain Points
1.  **Context Pollution:** Still a major issue for large-scale deployments.
2.  **Configuration Injection:** The "Claude Hooks" exploit proves that configuration files are a major attack vector.
3.  **Trust Deficit:** Lack of cryptographic provenance for tools makes enterprises hesitant to adopt open-source MCP servers.
