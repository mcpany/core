# Design Doc: Structured I/O Bus Adapter
**Status:** Draft
**Created:** 2026-03-22

## 1. Context and Scope
Current agent frameworks primarily rely on terminal emulation (PTY) to interact with CLI tools. This 1970s-era technology is designed for human eyes, not machine reasoning. Raw byte streams are difficult for LLMs to parse, and they make semantic security inspection nearly impossible, as malicious commands can be obfuscated within escape codes or interleaved outputs.

The Universal Agent I/O Bus initiative seeks to replace PTYs with a "Structured Event Bus." MCP Any must implement a Command Adapter that follows this standard, treating command execution as a sequence of typed, framed JSON events.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Transition the `Command` Upstream Adapter from raw byte streams to structured JSON framing.
    *   Implement "Output Schema Mapping" to allow CLI tools to return typed data directly to the agent.
    *   Enable real-time semantic sanitization of structured outputs.
    *   Support multiplexed streams (stdout, stderr, telemetry, metadata) over a single bus.
*   **Non-Goals:**
    *   Maintaining 100% compatibility with interactive TUI applications (e.g., `vim`, `htop`). The bus is optimized for agentic tool use, not human interaction.
    *   Replacing the underlying OS shell (it still executes standard binaries).

## 3. Critical User Journey (CUJ)
*   **User Persona:** AI Agent Developer
*   **Primary Goal:** Safely execute a complex CLI refactoring tool and ingest its structured results without parsing raw text.
*   **The Happy Path (Tasks):**
    1.  Developer configures a `command` service in MCP Any with `structured_io: true`.
    2.  Agent issues a `tools/call` for the CLI tool.
    3.  MCP Any spawns the process and wraps its I/O in a framed JSON bus.
    4.  The CLI tool emits a "Progress Event" (JSON).
    5.  The tool emits a "Result Event" containing a structured list of modified files.
    6.  MCP Any validates the JSON against a schema and passes it to the agent.
    7.  The agent receives a typed payload rather than a block of text.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        Agent[AI Agent] -->|JSON-RPC| Adapter[Structured I/O Adapter]
        Adapter -->|Framed JSON| Wrapper[Bus Wrapper Script]
        Wrapper -->|Spawn| Binary[Native CLI Binary]
        Binary -->|Stdout/Stderr| Wrapper
        Wrapper -->|Frame & Type| Adapter
        Adapter -->|Typed Result| Agent
    ```
*   **APIs / Interfaces:**
    *   Standard `tools/call` remains the interface, but the `CallToolResult` now contains a `structured` field with the parsed JSON data.
    *   The framing protocol follows the [Length-Prefixed JSON] standard: `[4-byte Length][JSON Payload]`.
*   **Data Storage/State:**
    *   Stateless at the adapter level; state is managed by the underlying process.

## 5. Alternatives Considered
*   **Regex-based Parsing**: Rejected because it is brittle and cannot handle the non-deterministic nature of interleaved stdout/stderr.
*   **LLM-based Post-Processing**: Rejected due to high latency and the risk of hallucinating the output structure.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Structured I/O allows the Policy Engine to apply granular rules to the *output* of a command (e.g., "Block output if it contains a structured PII field").
*   **Observability:** The bus includes a dedicated `telemetry` event type for sub-millisecond execution tracking.

## 7. Evolutionary Changelog
*   **2026-03-22:** Initial Document Creation.
