# Market Sync: 2026-06-28

## Ecosystem Updates

### OpenClaw 2026.3.1 GA
* **General Availability**: OpenClaw 2026.3.1 has reached GA, solidifying "Isolated Named-Pipe" as the primary transport for local agent teams.
* **Subagent Event Streaming**: New support for subagent-level event triggers, allowing parent agents to react to reasoning monologues in real-time.
* **Context Anchoring Maturity**: The "Cognitive Anchoring" hooks are now stable, enabling MCP Any to host mission-root intents that persist across teammate rotations.

### Gemini CLI (v0.43.0)
* **ZKD Mandate**: Gemini CLI now defaults to Zero-Knowledge Discovery (ZKD) for all remote tool registries. Unauthenticated peers see only cryptographic proofs of skill, not the raw JSON-RPC schemas.
* **Attention Masking (v2.0)**: Introduction of "Attention Masks" allowing agents to cryptographically prioritize context fragments, neutralizing high-entropy noise injection from specialist subagents.

### Claude Code (v2.4.2)
* **Mailbox Splicing Defense**: Claude Code v2.4.2 introduces "Stylometric Anchoring" for teammate mailboxes. It cross-references inter-agent messages against the leader's behavioral profile to prevent unauthorized instruction splicing.
* **Horizontal Scaling Locks**: Even with v2.4.2, the 2-second "Mailbox Lock" latency remains a critical pain point for teams larger than 10 agents, fueling the demand for CRDT-native mailbox sharding.

## Autonomous Agent Pain Points
* **Stylometric Mimicry**: Attackers are now using "Fine-Tuned Persona Mimicry" to bypass hardware-attested identity filters by generating reasoning traces that match the stylometry of authorized supervisors.
* **Attention-Density Attacks**: The use of "Imperative Metadata" in tool descriptions to evict mission-critical instructions from parent attention windows is increasing.

## Security Vulnerabilities
* **CVE-2026-92002 (Attention Eviction)**: A vulnerability where high-frequency, plausible-sounding subagent feedback causes the parent agent to "forget" its root mission constraints.
* **CVE-2026-95001 (Schema Shadowing)**: A pre-flight discovery exploit where malicious agents inject identical capability cards with higher "reputation" scores to intercept handshakes.
