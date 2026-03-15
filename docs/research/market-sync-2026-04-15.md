# Market Sync: 2026-04-15

## Ecosystem Shifts & Competitor Analysis

### OpenClaw: Intent-Bound Context Quotas
* **Update:** OpenClaw has announced a draft for "Context Quotas" (v2026.4.0-alpha).
* **Impact:** Allows parent agents to set hard token and memory limits on subagent context usage, preventing "Context Storms" where a single subagent consumes the entire reasoning window.
* **Linkage:** Directly addresses the "Context Amnesia" pain point identified yesterday by ensuring fair resource distribution.

### Gemini CLI: eBPF-Powered Tool Guarding
* **Innovation:** Gemini CLI is testing an eBPF-based enforcement layer for local tools.
* **Mechanism:** Monitors system calls (read/write/network) in real-time and kills tools that deviate from their declared manifest.
* **Impact:** Moves beyond static file-watching to active, kernel-level execution integrity.

### Claude Code: The "Ghost Script" SVG Vulnerability
* **Security Alert:** CVE-2026-45201 ("Ghost Script") disclosed today.
* **Vulnerability:** Malicious instructions hidden in SVG `<metadata>` or `<foreignObject>` tags can be ingested by agents during directory scans, leading to indirect prompt injection.
* **Defense:** Requires semantic sanitization of multimodal assets before they enter the agent's context.

## Autonomous Agent Pain Points
* **Negotiation Exhaustion:** In complex swarms, agents are spending up to 30% of their token budget simply "negotiating" who should do what.
* **Metadata Poisoning:** Structural metadata (tool descriptions) is becoming a high-trust injection vector, as observed in the latest Claude Code exploits.

## Security & Vulnerability Scan
* **CVE-2026-45201 (Claude Code):** Multimodal injection via SVG assets.
* **Negotiation Storms:** New DoS pattern where subagents trigger infinite bidding wars in UACO-compliant swarms.
