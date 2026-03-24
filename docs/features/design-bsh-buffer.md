# Design Doc: BSH State Buffer
**Status:** Draft
**Created:** 2026-03-24

## 1. Context and Scope
The "Token Storm" crisis in deep swarms (OpenClaw v2.4) proves that JSON is no longer a viable transport for inter-agent state. BSH State Buffer provides a high-speed, memory-mapped buffer for binary state handoffs between agents to minimize context transfer latency and eliminate serialization overhead.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement zero-copy memory-mapped buffers for binary state.
    * Reduce inter-agent handoff latency for large context objects by >70%.
    * Provide a stable API for state differential synchronization.
* **Non-Goals:**
    * Provide long-term persistence for binary state.
    * Replace standard MCP JSON-RPC for simple tool calls.

## 3. Critical User Journey (CUJ)
* **User Persona:** Large-Scale Agent Developer
* **Primary Goal:** Pass 50MB of research context between three agents in less than 50ms.
* **The Happy Path (Tasks):**
    1. Agent A writes its final research state to a memory-mapped BSH buffer.
    2. Agent A signals a handoff to Agent B via MCP Any.
    3. MCP Any validates the buffer address and permissions.
    4. Agent B receives the buffer handle and maps it directly into its address space.
    5. Agent B continues the task without re-parsing JSON.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A] -> [Shared Memory Region] <- [MCP Any Buffer Manager] -> [Agent B]`
* **APIs / Interfaces:**
    * `mmap()` based buffer management.
    * `POST /v1/state/buffer/handoff` for handle exchange.
* **Data Storage/State:**
    * Session-bound memory regions (ephemeral).

## 5. Alternatives Considered
* **Shared Redis:** Rejected due to network serialization overhead.
* **File-based handoff:** Rejected due to Disk I/O latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Strictly isolated memory regions per session. Hardware-enclave (TPM) attestation for buffer handles.
* **Observability:** Monitor memory usage and map/unmap latency.

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.
