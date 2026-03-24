# Design Doc: Binary State Handoff (BSH) Gateway
**Status:** Draft
**Created:** 2026-03-23

## 1. Context and Scope
As agent swarms grow in depth and complexity, the overhead of transferring massive JSON-encoded context objects between agents leads to "Token Storms" and excessive latency. BSH Gateway implements a low-latency, binary-encoded transport mechanism (inspired by OpenClaw v2.0) to streamline state handoffs between MCP Any nodes and UACO-compliant agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-performance binary transport (e.g., Protobuf or FlatBuffers) for agent state.
    * Reduce context handoff latency by at least 50% compared to JSON.
    * Support "State Differential" updates (only sending what changed).
* **Non-Goals:**
    * Replace JSON-RPC for standard MCP tool calls (BSH is for internal state transfer).
    * Provide persistent storage for binary state (this is handled by the Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Coordinate a swarm of 10+ subagents on a complex task without incurring 5+ second latencies for every agent-to-agent transition.
* **The Happy Path (Tasks):**
    1. Parent agent completes a sub-task and prepares state for handoff.
    2. Parent agent encodes the state into a BSH buffer.
    3. MCP Any Gateway receives the binary buffer via a dedicated BSH endpoint.
    4. Gateway validates the buffer integrity and routes it to the target subagent.
    5. Subagent decodes the buffer and resumes the task with minimal overhead.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A] --(Binary Buffer)--> [BSH Gateway] --(Validated Buffer)--> [Agent B]`
* **APIs / Interfaces:**
    * `POST /v1/state/handoff` (Binary body).
    * WebSocket stream for real-time state synchronization.
* **Data Storage/State:**
    * Ephemeral memory-mapped buffers for high-speed routing.

## 5. Alternatives Considered
* **Compressed JSON (GZIP):** Rejected because decompression overhead at the agent level remains significant, and it doesn't solve the "Token Storm" issue in the LLM window.
* **Direct Agent-to-Agent P2P:** Rejected to maintain MCP Any's role as a central security and policy enforcement hub.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All BSH buffers must be encrypted at rest and in transit. Origin validation is mandatory.
* **Observability:** Track buffer size and encoding/decoding latency metrics.

## 7. Evolutionary Changelog
* **2026-03-23:** Initial Document Creation.
* **2026-03-24:** BSH-Native Buffer Update.
    * **Context:** OpenClaw v2.4 identifies JSON-based state transfer as the primary cause of "Token Storms."
    * **Architecture Adjustment:** Introduced "Memory-Mapped BSH Buffers" and "State Differential Sync" to eliminate serialization overhead and reduce binary delta sizes.
    * **Performance Impact:** Projected 30% reduction in state transfer latency and significant reduction in LLM token consumption.
    * **2026-03-25:** WASM-Bound Zero-Copy State Update.
    * **Context:** OpenClaw v2.5 moves toward "Active State Sanitization" to prevent binary context poisoning.
    * **Architecture Adjustment:** Integrated a "WASM-BSH Sanitizer" into the memory-mapped transport flow. Binary state is now validated against a signed schema within a WASM sandbox before being mapped into the target agent's address space.
    * **Security Impact:** Neutralizes "Binary Context Injection" attacks while maintaining sub-millisecond Zero-Copy performance.
    * **2026-03-26: WASM-BSH Active Sanitization Update**
        * **Context**: OpenClaw v2.5 and UACO v1.8 emphasize the risk of "Binary Context Poisoning" in deep swarms.
        * **Architecture Adjustment**: Integrated a pluggable WASM sandbox into the BSH handoff flow. Binary buffers are now validated against a signed Protobuf schema within the sandbox before being mapped to the target agent.
        * **Security Impact**: Neutralizes "State Injection" attacks while maintaining the performance benefits of Zero-Copy transport.
    ### Update: 2026-03-24 (v2) - BSH Efficiency & Token Storm Mitigation
    **Context:** Today's findings confirm that JSON-based state transfer is causing "Token Storms" in swarms of 10+ agents.
    **Architecture Adjustment:** * Transitioning to Protobuf-based Binary State Handoffs (BSH).
    * Introducing a high-speed "BSH State Buffer" using memory-mapped regions.
    **Performance Impact:** Eliminates JSON serialization overhead and reduces inter-agent latency by 50%.
