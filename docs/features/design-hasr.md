# Design Doc: Hardware-Attested Shard Resumption (HASR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms become more distributed, the cost of "Attestation Fatigue" — where every tool call or context handoff requires a full TPM-bound signature — is becoming a primary performance bottleneck. Agents moving between devices or resuming long-running missions experience 1s+ "Cognitive Stalls" due to redundant re-verification.

MCP Any needs a way to "lease" trust across devices. HASR allows for sub-50ms mission resumption by utilizing hardware-bound "Context Tickets" that attest to the integrity of a sharded context window without requiring a full re-signature of the reasoning lineage.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable sub-50ms mission resumption across physical hardware nodes.
    * Generate TPM-bound "Context Tickets" that are cryptographically linked to both the session and the hardware.
    * Maintain "Mission-Root Sovereignty" during device handoffs.
    * Support Gemini CLI v0.59.0 HASR patterns.
* **Non-Goals:**
    * Bypassing the initial mission-root attestation (initial boot still requires full verification).
    * Providing long-term persistent storage (HASR is for active session resumption).

## 3. Critical User Journey (CUJ)
* **User Persona:** Mobile Agent Developer
* **Primary Goal:** Start a mission on a laptop and seamlessly hand it off to a mobile device toolset without losing state or re-verifying the entire chain.
* **The Happy Path (Tasks):**
    1. Agent performs high-trust reasoning on Node A.
    2. User triggers a "Device Handoff" to Node B.
    3. Node A generates a HASR "Context Ticket" (TPM-signed).
    4. Node B receives the ticket, performs a fast-path hardware handshake.
    5. The mission resumes on Node B in <50ms with all context fragments intact.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant NodeA as Laptop (Node A)
        participant MCP as MCP Any Gateway
        participant NodeB as Mobile (Node B)

        NodeA->>MCP: Request HASR Ticket (Session ID)
        MCP->>NodeA: Challenge (TPM-bound)
        NodeA-->>MCP: Signed Evidence
        MCP->>NodeA: Context Ticket (Lease)
        NodeA->>NodeB: Transfer Ticket + Shard State
        NodeB->>MCP: Resume Mission (Context Ticket)
        MCP->>NodeB: Fast-Path Approval (<50ms)
    ```
* **APIs / Interfaces:**
    * `issueHASRTicket(sessionID string) (Ticket, error)`
    * `resumeWithTicket(ticket Ticket, state Shard) error`
* **Data Storage/State:**
    * Short-lived (5-min) ticket registry in secure memory.
    * Hardware ID (HID) binding table.

## 5. Alternatives Considered
* **Persistent Session Sockets:** Rejected due to fragility in unstable network environments (mobile).
* **Standard JWTs:** Rejected because they can be exfiltrated and replayed on un-attested hardware.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tickets are hardware-locked; even if stolen, they cannot be used on a device that doesn't match the TPM signature.
* **Observability:** Tracks "Handoff Latency" and "Resumption Success Rate" in the performance dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
