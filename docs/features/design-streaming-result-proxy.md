# Design Doc: Streaming Tool Result Proxy
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
Long-running tools (e.g., code execution, large database queries) currently block the agent until completion. This leads to UX "hangs" and prevents agents from aborting failing processes early. Furthermore, streaming tool outputs are vulnerable to "Streaming Injection" attacks. This design introduces a secure proxy for streaming tool results.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized SSE (Server-Sent Events) or WebSocket interface for streaming tool logs.
    * Implement a "Streaming Sanitizer" that filters out potentially malicious control sequences in real-time.
    * Allow agents to send "SIGINT" or "Abort" signals to the upstream tool via MCP Any.
* **Non-Goals:**
    * Buffering the entire output (the goal is to stream).
    * Modifying the tool's internal logging logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using Claude Code with MCP Any
* **Primary Goal:** Run a `npm test` tool and see the logs in real-time while ensuring the logs don't contain "hidden" instructions that trick the LLM.
* **The Happy Path (Tasks):**
    1. Agent calls a tool marked with `supportsStreaming: true`.
    2. MCP Any initiates the tool call and returns a `stream_id`.
    3. The client (UI/CLI) connects to `/v1/streams/{stream_id}`.
    4. Upstream tool writes to stdout.
    5. MCP Any's Streaming Proxy intercepts stdout, runs it through the Sanitizer.
    6. Sanitized chunks are pushed to the client.
    7. Client renders the logs; LLM observes the progress.

## 4. Design & Architecture
* **System Flow:**
    `Upstream Tool (Stdout) -> MCP Any Proxy -> Streaming Sanitizer -> Chunk Transformer -> SSE/WS Client`
* **APIs / Interfaces:**
    * `GET /v1/streams/{stream_id}`: Stream endpoint.
    * `DELETE /v1/streams/{stream_id}`: Abort the tool execution.
* **Data Storage/State:**
    * Transient stream buffers (ring buffers) to handle backpressure.

## 5. Alternatives Considered
* **Direct Tool-to-Client Streaming**: Rejected because it bypasses the security policy engine and sanitization layers.
* **Polling for Logs**: Rejected due to high latency and overhead.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The Sanitizer uses a whitelist of allowed characters and escapes all terminal control sequences (ANSI) except for a safe subset for coloring.
* **Observability**: Stream metadata (duration, bytes transferred) is recorded in the tool execution timeline.

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
