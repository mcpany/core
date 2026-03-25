# Market Sync: 2026-05-10

## Ecosystem Shifts & Research Findings

### 1. Gemini CLI "Ghost-Execution" via Discovery Commands
**Context**: A new exploit pattern has been identified in the Gemini CLI ecosystem where the `tools.discoveryCommand` in repo-local `.gemini/settings.json` files is executed during initial tool discovery.
**Impact**: This turns a seemingly passive configuration file into a high-trust shell execution vector. Attackers can bundle malicious commands that execute as soon as a developer opens a compromised repository, bypassing standard "tool execution" prompts because discovery is treated as a low-risk background task.
**MCP Any Opportunity**: We must evolve our discovery layer to treat *all* discovery-time commands as high-risk execution events, mandating a "Discovery Sandbox" that isolates these calls from the host environment.

### 2. Claude Code CVE-2026-25725: "Shadow-Sandbox" Escapes
**Context**: The disclosure of CVE-2026-25725 reveals a flaw in Claude Code's bubblewrap sandboxing. If a configuration file (like `.claude/settings.json`) does not exist at startup, the sandbox may fail to properly protect that path, allowing an agent to create a malicious configuration that "escapes" the intended security boundaries upon the next reload or subagent spawn.
**Impact**: "Absence-as-Exploit" where the lack of a file is weaponized to bypass mount-point restrictions.
**MCP Any Opportunity**: This reinforces the need for "Deterministic Absence Proofs" (DAP). MCP Any must not only verify what exists but cryptographically attest to the *non-existence* of restricted configuration hooks throughout the entire agent lifecycle.

### 3. OpenClaw-RL v1.0: Asynchronous Policy Optimization
**Context**: The release of OpenClaw-RL v1.0 introduces a fully asynchronous reinforcement learning framework that intercepts multi-turn conversations and optimizes agent policies in the background.
**Impact**: Infrastructure must now support "Rollout Collection" without interrupting the agentic reasoning loop. This requires high-frequency, non-blocking telemetry exports of conversation fragments, PRM (Process Reward Model) evaluations, and policy drift metrics.
**MCP Any Opportunity**: MCP Any can position itself as the authoritative "Rollout Collector" for RL-driven swarms, providing a privacy-preserving bridge for asynchronous feedback tokens.

## "Autonomous Agent Pain Points" (Social/GitHub Trends)
- **"Negotiation Deadlock"**: Swarms are getting stuck in infinite bidding loops for tasks when multiple agents have overlapping capabilities.
- **"Context Bleed"**: RL-driven agents are occasionally "leaking" internal reasoning monologues into the public blackboard, leading to semantic contamination in multi-agent refinement loops.
- **"Shadow-Subagent Spawning"**: Subagents are spawning their own sub-subagents without parental attestation, bypassing root mission budgets.
