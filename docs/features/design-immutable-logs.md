# Design Doc: Immutable Execution Logs

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
As AI agents move from simple assistants to autonomous swarms, the potential for "untraceable" malicious actions increases. If an agent is compromised or follows a malicious configuration (like the Claude Code RCE exploits), it may attempt to hide its tracks by deleting logs or altering history.

MCP Any needs to provide a "Black Box" recorder for agentic execution that is separate from the agent's memory and tamper-proof.

## 2. Goals & Non-Goals
* **Goals:**
    * Create a cryptographic, append-only ledger of all tool calls, inputs, and outputs.
    * Support "Proof of Execution" using SHA-256 chaining.
    * Enable external auditing without allowing agents to modify historical entries.
* **Non-Goals:**
    * Replacing standard application logs (stderr/stdout).
    * Providing long-term archival (this is a runtime integrity feature).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor / DevOps Engineer
* **Primary Goal:** Verify that a suspicious agent action was actually requested and see its full context.
* **The Happy Path (Tasks):**
    1. Agent initiates a tool call (`fs:write`).
    2. MCP Any intercepts the call.
    3. MCP Any hashes the previous log entry, the current timestamp, tool name, and arguments.
    4. Entry is written to a local, restricted-access SQLite database (the Ledger).
    5. Tool is executed.
    6. Tool response is appended to the Ledger with a new link in the hash chain.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> Gateway Middleware (Audit Hook) -> [Hash Chainer] -> SQLite Ledger -> Upstream Tool`
* **APIs / Interfaces:**
    * `GET /api/v1/audit/logs`: Retrieve the ledger.
    * `POST /api/v1/audit/verify`: Verify the integrity of the ledger hash chain.
* **Data Storage/State:**
    * SQLite with `journal_mode=WAL` for high-performance append-only writes.
    * File permissions restricted to the `mcpany` service user.

## 5. Alternatives Considered
* **Standard File Logging:** Rejected because files are easily modified or deleted by a process with shell access (RCE risk).
* **Remote Logging (Syslog/SIEM):** Good for enterprise, but adds latency and complexity for local-first users. Will be considered as a future "Export" feature.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Ledger itself must be protected from the Agent. Even if the Agent gains RCE, it should not have write-access to the SQLite file directly (enforced via OS-level user isolation).
* **Observability:** Audit logs will be exposed via the UI in a dedicated "Audit Trail" viewer.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
