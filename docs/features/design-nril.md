# Design Doc: Non-Repudiable Intent Logger (NRIL)
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As AI agent swarms move from experimental linear sessions to massive, heterogeneous service meshes, the concept of "Session Accountability" is breaking down. Currently, while tool calls can be gated, the cryptographic link between a high-level user intent and a low-level tool call often decays across multiple subagent handoffs. This leads to "Identity Squatting," where specialized agents retain access to parent capabilities long after their specific sub-task is complete.

The Non-Repudiable Intent Logger (NRIL) solves this by acting as the authoritative "Mission Notary." It implements a hardware-attested, hash-chained audit log that cryptographically binds every single tool call and state mutation back to a specific, user-authorized mission-root intent fragment.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a hash-chained audit log for all tool-call and coordination events.
    * Require hardware-bound (TPM/SEP) signatures for every log entry.
    * Provide sub-millisecond cryptographic binding between tool calls and "Active Intent" fragments.
    * Neutralize "Identity Squatting" by validating the active mission branch for every call.
* **Non-Goals:**
    * NRIL will NOT act as a general-purpose logging sink for agent debug text.
    * NRIL will NOT perform real-time semantic analysis (this is delegated to ARI/AID Hubs).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor for Enterprise Swarms
* **Primary Goal:** Verify that a "Database Specialist" agent only accessed the production database while executing the specific "Query User History" intent approved by the user.
* **The Happy Path (Tasks):**
    1. User authorizes the mission "Sync User Analytics" with a hardware-attested root intent.
    2. Primary Agent spawns a "Database Specialist" subagent with a task-bound sub-intent.
    3. The Specialist agent initiates a `db_query` tool call.
    4. NRIL intercepts the call, verifies the subagent's hardware-attested lineage back to the root intent.
    5. NRIL generates a hash-chained log entry: `hash(prev_entry + mission_id + intent_id + tool_call_meta)`.
    6. NRIL appends the hardware signature and commits to the notarized log.
    7. Tool call proceeds only after notarization.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>Gateway: Tool Call + Session Token
        Gateway->>NRIL: Notarize(Call Meta, Active Intent)
        NRIL->>TPM: Sign(Hash(Prev + Meta + Intent))
        TPM-->>NRIL: Hardware Signature
        NRIL->>Log: Append Entry
        NRIL-->>Gateway: Notarized (Ack)
        Gateway->>Tool: Execute
    ```
* **APIs / Interfaces:**
    * `NotarizeIntent(intent_fragment, parent_hash) -> intent_hash`
    * `VerifyLineage(tool_call_hash, root_mission_id) -> bool`
* **Data Storage/State:**
    * Notarized logs are stored in a local, append-only SQLite store with periodic merkle-tree anchors committed to a secure remote sink.

## 5. Alternatives Considered
* **Centralized Database Logging**: Rejected due to high latency and lack of non-repudiability (anyone with DB access could alter logs). NRIL requires local hardware signatures.
* **Header-only Lineage**: Rejected because headers can be "squatted" or replayed by compromised specialist processes. NRIL forces a unique, chained hash for every event.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** NRIL is the foundation of Zero Trust identity. If an agent cannot prove its lineage to a currently active intent, it is denied execution.
* **Observability:** Notarized logs provide a high-fidelity forensic trail that is immune to "Gaslighting" by compromised agents.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
