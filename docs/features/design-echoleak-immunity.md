# Design Doc: EchoLeak Immunity Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The "EchoLeak" vulnerability represents a critical escalation in agentic security, where zero-click prompt injections coerce agents into mirroring sensitive system data or exfiltrating it via out-of-band side channels. Current security models focus on input sanitization (IDS), but EchoLeak exploits the agent's autonomous decision-making after the input has been ingested.

MCP Any needs to solve this by implementing an "Output Scrubber" that validates tool responses before they are re-ingested by the reasoning engine. This ensures that even if an agent is successfully "tricked" by an injection, the resulting exfiltration attempt is blocked at the semantic boundary.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time semantic analysis of all tool outputs.
    * Detect and block "Mirroring" patterns (where tool output contains raw mission-root secrets).
    * Neutralize hidden exfiltration tags (e.g., Markdown-based image pixels) in tool responses.
    * Provide hardware-attested logs of blocked exfiltration attempts.
* **Non-Goals:**
    * Rewriting agent reasoning logic (the middleware only blocks/modifies the tool response data).
    * Providing full PII scrubbing (handled by the PII-Sovereign Context Scrubber).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent an autonomous subagent from exfiltrating environment variables via a malicious GitHub Issue title.
* **The Happy Path (Tasks):**
    1. A specialist agent reads a GitHub Issue containing a zero-click prompt injection.
    2. The agent attempts to call a tool (e.g., `read_env_vars`) and succeeds.
    3. The agent reasoning engine prepares to ingest the environment variables.
    4. The EchoLeak Immunity Middleware intercepts the tool output.
    5. The middleware detects that the output contains sensitive keys that match the "Mission-Root Secret Manifest."
    6. The middleware redacts the output and raises a security alert.
    7. The agent receives a "Security Interdiction" response instead of the raw data.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Tool Call` -> `MCP Server` -> `Tool Result` -> **[EchoLeak Immunity Middleware]** -> `Sanitized Result` -> `Agent Reasoning`
* **APIs / Interfaces:**
    * New middleware hook: `ProcessToolOutput(request *mcpserver.ToolRequest, response *mcpserver.ToolResponse) error`
    * Integration with `IDS` for structural scanning.
* **Data Storage/State:**
    * Uses an in-memory "Mission-Root Secret Cache" (TPM-backed) to perform fast matching against tool outputs.

## 5. Alternatives Considered
* **Input-Only IDS**: Rejected because EchoLeak exploits logic *after* ingestion, making input-only checks blind to the actual exfiltration payload.
* **Full Air-Gapping**: Rejected due to the performance and utility requirements of autonomous swarms needing real-time tool access.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The middleware operates as a kernel-level gate, meaning it cannot be bypassed by subagent reasoning.
* **Observability:** Blocked attempts are logged to the `Security Hub` with the complete "Lineage of Intent" to aid forensic analysis.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
