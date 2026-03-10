# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Market Sync: 2026-03-08

## Ecosystem Updates

### A2A Trust & Identity
- **Agent-DID Emergence**: The community is gravitating towards Decentralized Identifiers (DIDs) for agents. MCP Any must support resolving and verifying these identities to enable secure cross-framework swarms.
- **OpenClaw Multi-Tenant Swarms**: New requirements for isolating subagents from different vendors within the same Blackboard (shared state).

### Claude Code & Tool Verification
- **Verified Tooling**: Claude Code's latest internal updates suggest a shift towards "Signed Tool Results," where the tool itself (or its execution environment) provides a cryptographic hash of the output to prevent man-in-the-middle tampering.

## Security & Vulnerabilities

### Tool Output Sanitization (TOS) Gap
- **Prompt Injection via Tool Results**: A new exploit class where a compromised or malicious tool returns "system-like" instructions in its output to hijack the parent LLM.
- **Blackboard Poisoning**: Attackers are targeting shared state repositories (like MCP Any's Shared KV Store) to inject malicious "Context Bombs" that trigger when an agent reads them.

## Autonomous Agent Pain Points
- **Trust Negotiation Latency**: The overhead of verifying agent identities in A2A protocols is becoming a performance bottleneck.
- **Context Integrity**: No standardized way to ensure that the context inherited by a subagent hasn't been tampered with by an intermediate "proxy" agent.
