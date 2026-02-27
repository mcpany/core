# Market Sync: 2026-02-27

## Ecosystem Updates

### OpenClaw
- **Security Crisis**: Recent patch (2026.1.29) for CVE-2026-25253 highlights vulnerabilities in Control UIs that trust URL parameters, leading to one-click RCE via cross-site WebSocket hijacking.
- **Hardening Trend**: Version 2026.2.23 signals a shift toward treating AI tool invocation and local execution as a "moving control boundary" that requires constant validation rather than a static perimeter.

### Gemini CLI (v0.30.0)
- **Policy-as-Code**: Introduced a dedicated Policy Engine with a `--policy` flag and "strict seatbelt profiles." This moves away from simple allow-lists toward complex, user-defined behavioral constraints.
- **SessionContext**: New SDK support for `SessionContext` in tool calls, allowing tools to be aware of the broader agent session state.

### Claude Code / Anthropic
- **Scale through Embeddings**: Officially moving toward embedding-based tool discovery to handle "thousands of tools" without context pollution.
- **Human-in-the-Loop (HITL)**: Anthropic is framing their human-approval architecture as the mandatory standard for all "consequential" AI agent execution.
- **Vulnerability Discovery**: Claude Opus 4.6 demonstrated the ability to find 500+ zero-days, emphasizing the need for agents to operate within a "Red Teamed" or "Attested" tool environment.

### Agent Swarms & Security Research
- **Cascading Failures**: New research identifies "Cascading Agent Failure" as a primary threat, where a single compromised subagent poisons the entire swarm's state, making root-cause analysis difficult without deep inter-agent observability.
- **Indirect Prompt Injection**: NIST and others are standardizing "Agent Hijacking" as a top threat, where malicious data consumed by an agent triggers unintended tool actions.

## Unique Findings for MCP Any
1. **The "Control UI" is an Attack Vector**: We must ensure our management UI is isolated and does not trust input parameters for sensitive actions (e.g., adding a new MCP server).
2. **Policy Parity**: MCP Any needs to support or bridge Gemini's "Seatbelt" style policies to remain the universal adapter.
3. **Attestation is Mandatory**: With the rise of "Clinejection" and poisoned agent libraries, tool attestation isn't just a feature; it's the foundation of trust for the Universal Agent Bus.
