# Strategic Vision: MCP Any (Universal Agent Bus)

## Core Vision
To become the indispensable core infrastructure layer for all AI agents, subagents, and swarms. MCP Any acts as the "Universal Bus" that handles discovery, security, communication, and state for the next generation of autonomous agentic systems.

## Strategic Pillars
1. **Zero Trust Security:** Every tool call and context exchange is verified and scoped.
2. **Standardized Context:** Native support for recursive agent calls and context inheritance.
3. **Shared State:** Protocol-native "Blackboard" for inter-agent memory.
4. **Universal Connectivity:** Bridging any legacy or modern API to the MCP ecosystem.

## Strategic Evolution: 2026-02-22
**Focus:** Mitigating Subagent Side-Channels & Discovery Fatigue.

- **Unified Discovery Engine:** Addressing "Discovery Fatigue" identified in Gemini CLI. MCP Any must implement a smart tool aggregator that deduplicates and prioritizes tools based on active session context.
- **Isolated Inter-Agent Comms:** Transitioning from local HTTP ports to Docker-bound named pipes. This aligns with the "Zero Trust" trend and mitigates SSRF vulnerabilities in local swarms.
- **Recursive Scoping:** Standardizing context inheritance headers to ensure subagents cannot escalate privileges granted to their parents (addressing OpenClaw's gaps).
