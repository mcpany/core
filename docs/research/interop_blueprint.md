# Interop Blueprint: Universal Agent Bus (MCP Any)

**Status:** Finalized
**Created:** 2026-06-16

## 1. Executive Summary

As multi-agent swarms scale and diversify, the "Model Context Protocol (MCP) Any" standard operates as the Universal Agent Bus. This document details the supported standards and cross-framework adapter integration required to bridge the communication gaps between disparate AI agent paradigms—specifically OpenClaw, CrewAI, and AutoGen.

Our goal is robust **Protocol Sync**, ensuring reliable Agent-to-Agent (A2A), Agent Context Protocol (ACP), and MCP handoffs.

## 2. Supported Standards

1.  **MCP (Model Context Protocol):**
    *   The baseline protocol for agents interacting with external data sources, CLIs, and APIs.
    *   *Role in MCP Any:* The transport mechanism for the discovery and invocation of local tools and pseudo-agent tools.
2.  **A2A (Agent-to-Agent Protocol):**
    *   The standardized Linux Foundation protocol for task negotiation, delegation, and state sharing between independent agents.
    *   *Role in MCP Any:* Handles the messaging hub, authenticated task cards, intent delegation, and shared mailbox coordination.
3.  **ACP (Agent Context Protocol):**
    *   The specialized protocol for maintaining state, reasoning traces, and contextual boundaries across long-lived swarm sessions.
    *   *Role in MCP Any:* Synchronizes state with external Context Engines (e.g., OpenClaw v2026.3.7 lifecycle hooks) and enforces context boundaries.

## 3. Supported Framework Integration (The "Adapter Hub")

MCP Any provides a universal **Bridge Pattern** implementation in Go (`src/interop/`) to interface seamlessly with:

### A. OpenClaw

*   **Focus:** State Versioning, Context Sovereignty, and Reasoning Entropy Exhaustion (REE) Defense.
*   **Adapter Features:**
    *   *State Versioning:* Supports OpenClaw's `reasoning_epoch` to ensure UI and task states match the current adaptive reasoning effort, preventing visual desync.
    *   *Context Engine Bridge:* Translates native OpenClaw plugin hooks into MCP Any security policies and state updates.
    *   *Semantic Pruning:* Enforces low-entropy communication (L7SIH) to prevent REE attacks during swarm coordination.

### B. CrewAI

*   **Focus:** Role-Based Task Delegation and Synchronous Handoffs.
*   **Adapter Features:**
    *   *Task Proposal Mapping:* Translates CrewAI's task definition schema into authenticated A2A task cards.
    *   *Role Discovery:* Ensures "Authenticated Agent Card Discovery" is mandated, mapping CrewAI roles to secure capability tokens.

### C. AutoGen

*   **Focus:** Multi-Agent Conversation, Subagent Execution, and Asynchronous Mailboxes.
*   **Adapter Features:**
    *   *Mailbox Integrity Middleware:* Validates subagent coordination messages against the hardware-attested mission root.
    *   *Stateful Checkpoints:* Hooks into AutoGen's chat history to provide "Sandbox Persistence Proofs" before authorizing high-sensitivity operations.

## 4. Conclusion

By implementing a generalized Bridge interface, MCP Any normalizes external framework quirks into standardized A2A and MCP primitives. This ensures a 100% interoperable, zero-trust swarm environment without framework-specific hacks.