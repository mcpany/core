# Design Doc: Enterprise Visibility Sidecar
**Status:** Draft
**Created:** 2026-03-22

## 1. Context and Scope
According to the 2026 Gravitee State of AI Agent Security report, while 80.9% of technical teams have deployed agents, only 21% of organizations have actual visibility into what their agents can access, which tools they call, or what data they touch. This "Governance Gap" prevents enterprise adoption and creates massive security risks.

MCP Any must solve this by evolving from a simple tool adapter into a mandatory "Visibility Sidecar." This system will sit between any agent framework (OpenClaw, Claude Code, etc.) and the underlying infrastructure, providing a centralized, immutable audit trail of all agentic activity.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide real-time, structured logging of every tool call, parameter, and result.
    *   Implement "Capability Discovery Logging" to track what tools agents are aware of.
    *   Ensure audit logs are hardware-attested and immutable.
    *   Provide a standardized export format for SIEM (Security Information and Event Management) integration.
*   **Non-Goals:**
    *   The sidecar will not perform real-time blocking (this is handled by the Policy Firewall).
    *   It will not store the full content of agent monologues (to preserve privacy and reduce storage overhead), only the external interactions.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Operator (CISO/SecOps)
*   **Primary Goal:** Audit all interactions between a new "Research Agent" and the internal corporate database.
*   **The Happy Path (Tasks):**
    1.  Operator deploys MCP Any as a sidecar to the Research Agent.
    2.  The Research Agent discovers tools via MCP Any's registry; MCP Any logs the "Discovery Manifest."
    3.  The agent calls `query_database` with specific parameters.
    4.  MCP Any intercepts the call, logs the structured event (Agent ID, Tool, Params, Timestamp).
    5.  MCP Any signs the log entry with a TPM-bound key.
    6.  Operator views the live audit trail in the MCP Any UI or via an exported JSON stream in their SIEM.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent / Team] -->|Tool Request| Sidecar[Enterprise Visibility Sidecar]
        Sidecar -->|1. Event Capture| Buffer[Structured Event Buffer]
        Sidecar -->|2. Protocol Adapt| Upstream[Upstream Infrastructure]
        Buffer -->|3. Attestation| TPM[TPM Signing Engine]
        TPM -->|4. Persist| AuditLog[Immutable Audit DB]
        AuditLog -->|Stream| SIEM[Enterprise SIEM / UI]
    ```
*   **APIs / Interfaces:**
    *   `GET /api/v1/audit/stream`: Real-time Server-Sent Events (SSE) stream of attested tool calls.
    *   `GET /api/v1/audit/export`: Batch export of logs in JSONL format.
*   **Data Storage/State:**
    *   Logs are stored in an embedded SQLite database (`audit.db`) with row-level integrity checks.

## 5. Alternatives Considered
*   **Application-Level Logging**: Rejected because it requires framework-specific code changes and can be bypassed if the agent or its framework is compromised.
*   **Network-Level Inspection (e.g., eBPF)**: Rejected because agents often communicate over encrypted local channels (TLS/Named Pipes) or standard MCP (JSON-RPC over Stdio) which are hard to reconstruct at the kernel level.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The audit log itself must be protected. Logs are signed using hardware keys to prevent "Log Scrubbing" by compromised agents.
*   **Observability:** The Sidecar will export Prometheus metrics on "Log Latency" and "Capture Success Rate."

## 7. Evolutionary Changelog
*   **2026-03-22:** Initial Document Creation.
