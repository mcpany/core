# Market Sync: 2026-04-17 (Part 2)

## Unique Findings: Metadata Sovereignty & Lifecycle Hijacking

### 1. Metadata-Layer Instruction Injection (Agent Card Poisoning)
- **Finding**: New Proof-of-Concept (PoC) exploits demonstrate "Agent Card Poisoning," where malicious A2A discovery metadata (Agent Cards) embed adversarial instructions.
- **Impact**: Leads to data exfiltration via the host LLM when it parses the "shadowed" capabilities of a peer agent.
- **Requirement**: MCP Any must evolve from simple transport security to **Structural Metadata Sovereignty**, ensuring all discovery-time schemas are semantically sanitized.

### 2. Init-Phase RCE (Hackerbot-Claw)
- **Finding**: The "Hackerbot-Claw" campaign has been identified exploiting GitHub Actions workflows using poisoned `init()` functions in Go-based agent components.
- **Impact**: Remote Code Execution (RCE) occurs during the module loading phase, before any tool-call safety logic is triggered.
- **Requirement**: Urgent need for **Init-Function Isolation (IFI)** where agent plugins and MCP servers are loaded in an ephemeral pre-flight sandbox to profile initialization side-effects.

### 3. Semantic Parameter Injection (OpenAI Codex)
- **Finding**: Discovery of a command injection vulnerability in OpenAI Codex via unsanitized branch name parameters.
- **Impact**: Attackers can steal GitHub OAuth tokens by manipulating branch metadata ingested by the agent.
- **Significance**: Re-affirms the need for **Argument-Level Semantic Validation (ALSV)** even for seemingly benign metadata fields like branch names.

## Defensive Ecosystem Shifts
- **Pipelock**: A new 11-layer open-source agent firewall has emerged, focusing on DLP and MCP tool poisoning detection.
- **NVIDIA NemoClaw**: A reference stack for running OpenClaw in the NVIDIA OpenShell runtime, providing hardware-level isolation for agentic reasoning.

## Strategic Gap for MCP Any
- **Pre-Flight Behavioral Profiling**: While MCP Any has strong execution-time gating, today's findings reveal a critical gap in **Discovery-Phase Loading**. We must implement behavioral profiling for the *loading* of tool definitions themselves, not just their execution.
