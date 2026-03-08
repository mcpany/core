# Market Sync: 2026-03-08

## Ecosystem Shifts & Findings

### 1. "Agent Skills" Standard Adoption
*   **Context**: The "Agent Skills" format (introduced by Anthropic in late 2025) has become the de facto standard for portable agent capabilities. It uses a structured folder format (`SKILL.md`, `scripts/`, `references/`) and a "Progressive Disclosure" loading pattern.
*   **Significance**: Agents now download and execute these skills dynamically, moving beyond static tool definitions. This aligns with MCP Any's mission but introduces significant new attack vectors.

### 2. OpenClaw Hyper-Growth & The "ClawHavoc" Crisis
*   **Findings**: OpenClaw has reached 145,000 GitHub stars and 100,000+ users within weeks. However, the "ClawHavoc" campaign has exposed the fragility of this ecosystem, with over 300 malicious skills (crypto drainers, typosquats) discovered on the ClawHub marketplace.
*   **Pain Point**: Users are currently forced to choose between "Ease of Use" (dynamic skills) and "Security" (static, audited tools). There is no middle ground for secure, isolated execution of unverified skills.

### 3. Progressive Disclosure as a Security Pattern
*   **Concept**: The Agent Skills standard's "Progressive Disclosure" (Level 1: Name/Desc; Level 2: SKILL.md; Level 3: Code Execution) is being used to save context window space, but it also provides a framework for multi-stage security gates.

## Strategic Implications for MCP Any
*   **Skill-to-MCP Bridge**: MCP Any can act as the "Safe Runtime" for Agent Skills, converting the `SKILL.md` definitions into MCP tools while wrapping the execution in its Zero-Trust Policy Engine.
*   **Supply Chain Guardrails**: The "ClawHavoc" incident reinforces the need for our `Provenance-First Discovery` and `MCP Provenance Attestation` features.
*   **Isolation**: Requirement for Docker-bound or WebAssembly-based execution for skill-related scripts to prevent host-level compromise.
