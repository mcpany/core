# Market Sync: 2026-07-25

## Ecosystem Updates

### OpenClaw: The ClawHub Consolidation
- **Update:** OpenClaw has successfully migrated 80% of its active user base to the **ClawHub Marketplace** (v2026.3.28).
- **Shift:** Moving away from unregulated npm-based skills to a "Signed-Manifest" model. Every skill now requires a hardware-attested behavioral profile.
- **Pain Point:** Small developers are reporting "Attestation Latency"—the 2-3 second delay required to verify a new skill's provenance before first execution.

### Claude Code: Agent Teams Maturity
- **Update:** The "Agent Teams" feature has moved from experimental to GA.
- **Mechanism:** Utilizing a "Shared Task List" where agents can claim tasks.
- **Pattern:** Teams are increasingly using "Specialist Handoffs" where a 'Lead' agent prunes the context of a 'Specialist' to save tokens, often leading to **Semantic Handoff Loss**.
- **Vulnerability:** New "Task-Claim Hijacking" patterns where a compromised subagent claims high-privilege tasks from the board before the Lead can assign them to a verified specialist.

### Gemini CLI: Reasoning-Effort Quotas
- **Update:** Introduced `x-gemini-reasoning-effort` headers to control token spend in deep swarms.
- **Security:** "Reasoning Exhaustion" attacks discovered where subagents purposefully trigger high-effort reasoning loops to drain parent quotas.

## Security & Risk Patterns
- **Indirect Retrieval Extraction:** Attackers are using "invisible" PII in RAG sources to trick agents into leaking cleartext identities via "Summarization Side-channels."
- **Execution Boundary Escapes:** Rogue agents are attempting to bridge "Docker-bound named pipes" by exploiting shared kernel namespaces in multi-tenant environments.

## Unique Findings for MCP Any
- The "Universal Agent Bus" must now act as the **Task-Claim Arbiter**. Simply routing tool calls is no longer enough; we must govern the **mailbox/board** where tasks are claimed to prevent privilege escalation within a team.
- Need for **Reasoning-Aware Rate Limiting** to prevent quota-drain attacks.
