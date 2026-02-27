# Design Doc: Agent System Bus (ASB) Connector
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As frameworks like OpenClaw evolve into Agent Operating Systems, they require a low-level, high-performance "System Bus" for inter-agent communication and tool routing. The ASB Connector will position MCP Any as that bus, providing standardized IPC (Inter-Process Communication) mechanisms like named pipes and shared memory for local swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide high-throughput IPC for local agents (Named Pipes, Unix Domain Sockets).
    * Implement a "Broadcasting" mechanism for agent state updates.
    * Support "Zero-Copy" tool result passing for large data payloads.
* **Non-Goals:**
    * Replacing the LLM-based reasoning (ASB is purely transport and routing).
    * Managing agent lifecycle (that's the job of the Agent OS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent OS Developer (e.g., OpenClaw Contributor)
* **Primary Goal:** Efficiently route tool calls and state between 50+ local subagents without the overhead of HTTP/TCP.
* **The Happy Path (Tasks):**
    1. Agent OS initializes and connects to the MCP Any ASB via a named pipe.
    2. Subagent A produces a 50MB dataset and "broadcasts" its availability on the bus.
    3. Subagent B receives the notification and reads the data via shared memory.
    4. Subagent B calls a tool via the ASB; MCP Any routes it to the correct provider and returns the result with minimal latency.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A] <-> [ASB (Named Pipes/Sockets)] <-> [MCP Any Router] <-> [Tools/Agent B]`
* **APIs / Interfaces:**
    * `mcpany-asb.sock`: Unix Domain Socket for local IPC.
    * `broadcast(topic, payload)`: Bus primitive for state sharing.
* **Data Storage/State:** Shared memory segments for large payloads; transient state in the ASB ring buffer.

## 5. Alternatives Considered
* **gRPC over Loopback:** Rejected due to serialization overhead for very large agent swarms.
* **HTTP/2:** Too much overhead for high-frequency inter-agent state synchronization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Socket-level permissions and capability-based tokens for every ASB message.
* **Observability:** Performance metrics (throughput, latency) for the bus itself.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
