# Market Sync: 2026-03-25

## Ecosystem Shifts & Findings

### 1. Claude Code RCE via Poisoned Repository Configs
Check Point Research has disclosed a critical Remote Code Execution (RCE) vulnerability in Claude Code. The exploit involves poisoning project-local configuration files (e.g., `.claude/settings.json`) in shared repositories. When an agent automatically ingests these settings, it can be coerced into executing malicious hooks or shell commands, leading to full host compromise.

### 2. Mass Malicious Skills in ClawHub (OpenClaw)
Antiy CERT confirmed the discovery of over 1,100 malicious skills across ClawHub, the primary marketplace for the OpenClaw AI agent framework. These skills often use "Delayed Payload" tactics, appearing legitimate during initial installation but later exfiltrating sensitive data or establishing reverse shells during high-context tasks.

### 3. Exposed MCP Infrastructure
Trend Micro identified nearly 500 MCP servers exposed to the public internet with zero authentication. This highlights a critical gap in "Safe-by-Default" infrastructure, where developers are deploying MCP tools for remote access without implementing proper gateway security or attestation.

### 4. Supply Chain Designation
The Pentagon has designated certain AI agent infrastructure components as "supply chain risks," specifically highlighting the lack of provenance and attestation in third-party tool protocols like MCP.

## Summary of Findings
- **Security**: "Identity-Only" trust is failing; we need "Execution-Aware" and "Provenance-Bound" security.
- **Pain Points**: Malicious project-local settings are the new "Prompt Injection" for agentic tools.
- **Strategic Gap**: A lack of standardized, behavioral profiling for agent skills before they are granted host access.
