# Market Sync: 2026-02-27

## Ecosystem Updates

### Gemini CLI v0.30.0: SessionContext & Policy Engine
- **Insight**: Google released Gemini CLI v0.30.0 featuring `SessionContext` for SDK tool calls and a robust Policy Engine. The move from `--allowed-tools` to a structured policy system mirrors MCP Any's "Policy Firewall" vision.
- **Impact**: MCP Any should support `SessionContext` propagation to allow Gemini-native agents to maintain state when calling tools through the MCP bridge.
- **MCP Any Opportunity**: Implement a "SessionContext Bridge" that maps Gemini's native session state to MCP Any's Shared KV Store (Blackboard).

### A2A Identity: Agent Attestation Tokens (AAT)
- **Insight**: The A2A protocol (Agent-to-Agent) is standardizing on "Agent Attestation Tokens" (AAT) to solve the identity problem in multi-agent swarms. These tokens allow an agent to prove its identity and authority to its peers.
- **Impact**: Without AAT support, MCP Any risks being a weak link in the security chain of federated agents.
- **MCP Any Opportunity**: Integrate AAT verification into the A2A Interop Bridge. MCP Any can act as an AAT issuer for local agents and a validator for incoming A2A messages.

### Claude Code: The "Verification" Bottleneck
- **Insight**: Claude Code's "Agentic Edit Loop" emphasizes the "Verify" phase. As agents take more complex actions, the latency and reliability of verification tools (tests, linters) become the primary constraint.
- **Impact**: Tool execution is no longer just "Success/Failure"; it's about "Verifiability."
- **MCP Any Opportunity**: Introduce "Verifiable Tool Execution" where tools can return "Proof of Work" or "Verification Metadata" that agents can use to skip manual verification steps.

## Autonomous Agent Pain Points
- **Policy Fragmentation**: Managing different security policies for Gemini, Claude, and local agents is becoming untenable.
- **Identity Spoofing in Swarms**: Difficulty in verifying that a task delegation request actually came from a trusted parent agent.
- **State Drift during Handoffs**: Losing session context (like the `SessionContext` in Gemini) when an agent hands off a task to an MCP tool.

## Security Vulnerabilities
- **Policy Bypass via SDK**: If the Policy Engine isn't enforced at the SDK/Middleware level, agents can bypass CLI-level restrictions.
- **Token Leakage in A2A Handoffs**: Insecure propagation of AATs can lead to session hijacking in federated agent meshes.
