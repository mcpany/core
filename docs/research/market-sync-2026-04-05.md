# Market Sync: 2026-04-05

## Ecosystem Updates

### OpenClaw (The "Security Crisis" Phase)
*   **Vulnerability Taxonomy**: Recent reports identify 190+ advisories. Key clusters: identity spoofing, exec allowlist bypasses (lexical parsing failures), and cross-layer composition.
*   **ClawJacked (CVE-2026-25253)**: Cross-site WebSocket hijacking (CVSS 8.8) is the primary "make-or-break" security failure. Attackers are bridging the browser-to-local gap to execute arbitrary code.
*   **Malicious Skills**: 335+ malicious skills discovered in ClawHub using innocuous names and "two-stage droppers" that execute entirely within LLM context.

### Claude Code (Horizontal Agent Teams)
*   **Peer-to-Peer Messaging**: Introduction of "Agent Teams" where agents coordinate via a directory-based mailbox system rather than just hierarchical lead-synthesis.
*   **Git-based Locking**: Task claiming is managed via filesystem writes in a shared directory, enabling autonomous coordination on shared codebases.
*   **Headless Remote Control**: Structural change allowing connection to running sessions from outside the terminal, moving Claude Code from a solo tool to infrastructure.

### Gemini CLI (On-Demand Skills)
*   **Agent Skills Standard**: Packaging instructions and assets into discoverable capabilities.
*   **Progressive Disclosure**: Only metadata is loaded initially; full instructions/resources are pulled only when activated to save tokens.
*   **Workspace vs. User Skills**: Standardizing the hierarchy of capability discovery (local project vs. global user).

## Agent Swarm Pain Points
*   **Cognitive Stall**: High-density swarms failing to reach convergence in state synchronization.
*   **Token Storms**: Peer-to-peer messaging causing exponential cost increases when context windows are mirrored across teammates.
*   **Identity Hijacking**: Compromised specialist agents "ghosting" or impersonating team leads within the mailbox bus.

## Unique Findings for Today
*   The shift from **Hierarchical Delegation** to **Horizontal Teammate Meshes** (Claude Code v2.1.32) confirms that MCP Any must support multi-mailbox coordination.
*   The **Progressive Disclosure** pattern in Gemini CLI suggests that our "Lazy-MCP" implementation should evolve into a full "Skill Registry" that supports instruction-bundling, not just JSON-RPC definitions.
