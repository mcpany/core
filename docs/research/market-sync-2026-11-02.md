# Market Sync: 2026-11-02

## Ecosystem Updates

### OpenClaw & ClawHub Expansion
OpenClaw v2026.3.22 has transitioned its plugin ecosystem to **ClawHub**, a curated marketplace. Key security features include **OpenShell SSH Sandboxing** for tool execution and active blocking of JVM injection paths. GPT-5.40 is now the default reasoning engine, showing significant improvements in multi-step planning.

### Claude Code: Agent Team Vulnerabilities
Recent disclosures (Issue #24505) highlight critical gaps in horizontal swarms. Current safety hooks lack **Teammate Identity** in their input payload, making it impossible to apply differentiated policies (e.g., Researcher vs. Implementer). More critically, **Inbox-based Social Engineering** has emerged as a vector where a compromised subagent instructs a sibling to execute dangerous commands, bypassing lineage-based checks that only look at the immediate caller.

### Gemini CLI: Terminal Workspace Evolution
Gemini CLI is evolving from a command-generator into a full **AI-Powered Terminal Workspace**. It is moving toward persistent, natural-language-driven terminal sessions where the agent maintains long-term state of the filesystem and environment.

## Emerging Threats & Pain Points

### Memory Injection & "Sleeper Agents"
Lakera AI research has demonstrated **Memory Injection** attacks where poisoned data sources corrupt an agent's long-term memory. This creates "Sleeper Agents" that defend false security beliefs. MCP Any must evolve to provide **Memory Integrity Verification**.

### Uncontrolled Retrieval of PII
As agents gain more autonomy in retrieval, "Uncontrolled Retrieval" of sensitive PII in response to benign queries from low-clearance users is becoming a major enterprise bottleneck.

### Coordination Hijacking via Inbox
The transition from linear sessions to teammate meshes has made the **Teammate Inbox** a high-value target for instruction injection, as coordination messages often carry higher implicit trust than external tools.
