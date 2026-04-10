# Market Sync: 2026-04-10

## Ecosystem Updates

### Claude Code & OpenClaw
* **CVE-2026-25725 (Sandbox Escape)**: A critical vulnerability was disclosed where agents could escape their sandbox by manipulating project-local configuration files *before* the environment was fully locked. This has led to a major industry shift toward "Deterministic Boot" sequences.
* **OpenClaw `ContextEngine` Stabilization**: The pluggable `ContextEngine` API has reached v1.0, allowing for more granular control over context pruning and "Intent-Bound" memory.
* **Agent Teams Launch**: Claude Code's "Agent Teams" feature is driving demand for horizontal coordination and sharded mailbox architectures to prevent coordination deadlocks.

### Gemini CLI
* **Ghost-Execution Vulnerability**: A new exploit pattern emerged where discovery commands (e.g., `discoveryCommand`) were executed with excessive privileges during the tool discovery phase, leading to host-level compromise.
* **ARE (Advanced Reasoning Effort) Headers**: Gemini now supports signaling reasoning intensity, requiring infrastructure to perform "Reasoning-Aware Throttling."

## Market Pain Points & Trends
1. **Instruction Eviction**: Agents are losing critical mission guardrails due to aggressive context-window garbage collection in 1M+ token models.
2. **Context Smuggling**: Attackers are using multi-modal metadata (SVG, CSS) to inject "invisible" instructions that bypass text-based filters.
3. **Loopback Trust Gap**: Implicit trust in `localhost` is being exploited via browser-based cross-site scripting (XSS) to hijack local agent gateways.

## Unique Findings
* **Negative Discovery Attestation**: A new security pattern where agents must cryptographically prove the *absence* of unauthorized configuration hooks before execution.
* **Deterministic Boot Manifests**: The move from "Partial Sandbox" to "Full-State Manifests" where the entire environment state is signed before agent boot.
