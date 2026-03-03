# Design Doc: A2A Interop Bridge (Pseudo-MCP)

**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
As AI agent ecosystems diversify, models are no longer just interacting with static tools via MCP; they are increasingly interacting with other agents. While MCP is excellent for Model-to-Tool communication, the Agent-to-Agent (A2A) protocol is emerging as the standard for inter-agent task delegation and state sharing. MCP Any needs to bridge this gap by allowing MCP-native agents to "call" A2A-capable agents as if they were standard MCP tools.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Expose A2A-compliant agents as standard MCP tools.
    *   Support standardized handoff and callback mechanisms between different agent frameworks (e.g., CrewAI, AutoGen, OpenClaw).
    *   Maintain session-aware context during multi-agent handoffs.
    *   Provide a "Pseudo-MCP" wrapper for A2A endpoints.
*   **Non-Goals:**
    *   Replacing the A2A protocol itself.
    *   Implementing agent logic (MCP Any remains the transport/gateway layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Swarm Architect.
*   **Primary Goal:** Enable a Claude-based Research Agent (MCP-native) to delegate a coding task to a GPT-based Coding Agent (A2A-native) and receive the results back.
*   **The Happy Path (Tasks):**
    1.  Architect configures an A2A endpoint (e.g., a CrewAI specialist) in MCP Any.
    2.  MCP Any registers this endpoint as a tool: `call_agent_coding_specialist`.
    3.  The Research Agent calls `call_agent_coding_specialist(task="Write a Python script for...")`.
    4.  MCP Any translates the MCP tool call into an A2A message and routes it to the specialist.
    5.  MCP Any manages the "waiting" state and provides the specialist's response back to the Research Agent as the tool output.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery**: MCP Any polls or receives registrations from A2A agents.
    - **Translation**: The `A2ABridgeMiddleware` maps MCP `tools/call` arguments to A2A `message/post` payloads.
    - **Session Management**: MCP Any uses the `Recursive Context Protocol` headers to track the lineage of the call across agent boundaries.
*   **APIs / Interfaces:**
    - **MCP Side**: Standard `tools/list` and `tools/call`.
    - **A2A Side**: Implementation of A2A transport (likely SSE or WebSockets).
*   **Data Storage/State:** A2A session tokens are stored in the `Shared KV Store` (Blackboard) to allow for asynchronous callbacks.

## 5. Alternatives Considered
*   **Direct A2A Integration in Agents**: Forcing every agent framework to implement A2A. *Rejected* because it increases complexity for agent developers and lacks centralized security/observability.
*   **Custom Tool Callbacks**: Building a proprietary callback system. *Rejected* in favor of the emerging A2A industry standard.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** A2A agents are treated as "External Tools." Every handoff must be authorized by the Policy Firewall. Identity is verified via A2A Attestation tokens.
*   **Observability:** Trace the full "Agent Chain" in the UI, showing which agent called which and the latency of each hop.

## 7. Evolutionary Changelog
*   **2026-02-26:** Initial Document Creation.
*   **2026-03-03:** Updated to align with OpenClaw v2026.3.2 release.
    *   **"Messaging-First" Alignment**: Default configurations for A2A bridge endpoints now favor a messaging-first approach, prioritizing task handoffs over raw tool execution.
    *   **Explicit Auth-First Routing**: All registered A2A routes now require explicit authentication declarations (`auth_required: true` by default) to prevent unauthorized "Ghost Handoffs."
    *   **SecretRef Integration**: A2A tokens and credentials are now managed via the `SecretRef` mechanism, ensuring they are not exposed in plaintext during the handoff process.
