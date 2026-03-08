# Market Sync: 2026-03-05

## Ecosystem Updates

### OpenClaw: Handoff Verification Standard
OpenClaw has released a major update (v2.4) focusing on "Secure Agent Handoffs." They've identified a vulnerability where a malicious subagent could intercept a handoff and impersonate a trusted peer. Their new "Verification Token" standard requires every handoff to be signed by the orchestrator.

### Claude Code: Ephemeral Tool Sessions
Anthropic's Claude Code has transitioned to "Ephemeral Tool Sessions" for its remote sandbox. This means tool permissions are now tied to a single user message/turn rather than the entire session. This significantly reduces the window for "Session Hijacking" via prompt injection.

### Gemini CLI: Streaming MCP Support
Google's Gemini CLI now natively supports streaming responses from MCP tools. This is a significant shift for long-running data processing tools, allowing the LLM to process partial outputs before the tool execution is fully complete.

## New Pain Points & Vulnerabilities

### Prompt-Injected Tool Discovery (PITD)
A new class of attack, "Prompt-Injected Tool Discovery," has been identified. Attackers use crafted prompts to trigger "Auto-Discovery" of malicious MCP servers on the same local network or in a shared Docker bridge. This bypasses traditional firewall rules by exploiting the "Discovery Middleware" itself.

### The "Latency Tax" in Federated Meshes
As agents move towards the "Federated Mesh" (A2A), early adopters are reporting a "Latency Tax" where the overhead of multi-hop tool discovery and attestation is significantly slowing down agent response times.

## Summary for MCP Any
- **Urgent**: We need to ensure our "On-Demand Discovery" (Lazy-MCP) is immune to PITD by requiring attestation *before* discovery, not just before execution.
- **Opportunity**: Implementing "Ephemeral Scoping" in our Policy Firewall to match Claude Code's security posture.
- **Architecture**: We should consider supporting "Streaming Tool Pipelines" to stay compatible with the latest Gemini updates.
