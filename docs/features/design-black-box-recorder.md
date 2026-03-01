# Design Doc: Immutable Black Box Recorder (BBR)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As AI agents move from advisory roles to executing autonomous real-world actions (e.g., financial trades, server maintenance), the risk of "Unaccountable Autonomy" increases. If an agent executes a destructive command, there must be an unalterable, cryptographically signed record of the "Why" and "How." Current logs are easily tampered with or lack the necessary context of the LLM's internal reasoning.

## 2. Goals & Non-Goals
* **Goals:**
    * Create an immutable, append-only ledger of all tool calls and agentic decisions.
    * Capture "Reasoning Context" (the thought process leading to an action).
    * Provide cryptographic proof of the state of the system at the time of execution.
    * Enable "Time-Travel Auditing" for incident response.
* **Non-Goals:**
    * Storing full conversation history (handled by separate memory systems).
    * Real-time monitoring (handled by Observability stacks).

## 3. Critical User Journey (CUJ)
* **User Persona:** Compliance Officer / SRE.
* **Primary Goal:** Investigate why an agent deleted a production database table.
* **The Happy Path (Tasks):**
    1. Incident occurs.
    2. Auditor opens the BBR Viewer.
    3. Auditor locates the specific `delete_table` tool call.
    4. BBR provides the "Signed Intent Package" containing:
        - The specific prompt/instruction from the parent agent.
        - The LLM's internal reasoning (Chain of Thought).
        - The verified tool schema and arguments.
        - The cryptographic hash of the session state.

## 4. Design & Architecture
* **System Flow:**
    - **Hooking**: Every `tools/call` is intercepted by the BBR Middleware.
    - **Signing**: The intent, arguments, and metadata are signed with the MCP Any server's private key.
    - **Storage**: Data is written to a local SQLite-backed "WORM" (Write Once, Read Many) log or a dedicated audit-safe backend (e.g., Immudb).
* **APIs / Interfaces:**
    - `POST /audit/record`: Internal endpoint for BBR logging.
    - `GET /audit/verify/:id`: Endpoint to verify the integrity of a record.
* **Data Storage/State:** Encrypted, append-only local storage with periodic remote attestation.

## 5. Alternatives Considered
- **Standard ELK Logging**: *Rejected* because logs can be deleted/modified and don't include the cryptographic proof needed for non-repudiation.
- **Blockchain**: *Rejected* due to latency and cost, although the BBR could optionally anchor hashes to a public ledger.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The BBR itself is a high-value target. Access to the BBR requires multi-party authorization.
* **Observability:** BBR provides the data source for the "Agent Black Box Player" in the UI.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
