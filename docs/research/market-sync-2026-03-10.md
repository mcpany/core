# Market Sync: 2026-03-10

## Ecosystem Updates

### OpenClaw (formerly Moltbot/Clawdbot)
- **Security Crisis**: Documented CVEs regarding Remote Code Execution (RCE) even on localhost-bound instances.
- **Isolation Shift**: The community is moving towards "never run on daily driver" and mandatory containerization.
- **Malicious Skills**: Increasing reports of "poisoned" skills circulating in the wild that attempt to exfiltrate host environment variables.
- **Multi-Agent Refinement**: OpenClaw is leaning heavily into multi-agent patterns where specialized subagents refine each other's work, increasing the need for secure state handoffs.

### Gemini CLI & Claude Code
- **Local-to-Cloud Bridging**: Continued push for agents running in cloud sandboxes to reach back to local tools securely.
- **Context Management**: Both platforms are struggling with "Context Fatigue" as tool schemas grow, reinforcing the need for MCP Any's Lazy-Discovery.

## Autonomous Agent Pain Points
1. **Trust Deficit**: Users are afraid to grant agents filesystem access due to RCE risks.
2. **Context Pollution**: Subagents inheriting too much irrelevant state, leading to hallucinations.
3. **State Fragmentation**: Loss of progress when switching between specialized agents in a swarm.

## Unique Findings
- **"Sociality Illusions"**: New research suggests agents in multi-agent environments can be "tricked" by other agents into bypassing local security policies through social engineering (Agent-to-Agent).
- **Silent RCE**: Exploits that use legitimate-looking git hooks or project-local configs to trigger agent actions that exfiltrate data.
