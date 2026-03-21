# Design Doc: Inter-Agent Mailbox Guard (IAMG)
**Status:** Draft
**Created:** 2026-03-17

## 1. Context and Scope
With the rise of "Agent Teams" (Claude Code) and swarms, agents now coordinate via asynchronous messaging systems (Mailboxes). However, these communication channels are often unauthenticated or rely on implicit local trust. A compromised teammate could "Mailbox Inject" malicious instructions into a sibling agent, leading to unauthorized tool execution or context exfiltration. MCP Any needs to provide a secure, Zero-Trust mediation layer for this inter-agent bus.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce identity-bound authentication for all teammate-to-teammate messages.
    *   Validate message intent against a cryptographically signed "Parental Mission Root."
    *   Provide an immutable audit log of all inter-agent coordination.
*   **Non-Goals:**
    *   Replacing the underlying transport layer (e.g., we support existing HTTP/WebSocket mailboxes).
    *   Managing the actual task scheduling (that remains with the lead agent).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local AI Swarm Orchestrator
*   **Primary Goal:** Ensure Teammate B only executes a "File Delete" if it was explicitly authorized by the Team Lead and aligns with the original user-signed intent.
*   **The Happy Path (Tasks):**
    1.  Team Lead sends a task to Teammate B via the IAMG-mediated mailbox.
    2.  IAMG intercepts the message and verifies the Lead's signature.
    3.  IAMG validates that the task aligns with the "Mission Root" intent.
    4.  Teammate B receives the verified message and executes the task.
    5.  All steps are logged in the secure audit trail.

## 4. Design & Architecture
*   **System Flow:**
    `[Teammate A] -> [IAMG Interceptor] -> [Policy Engine] -> [Teammate B Mailbox]`
*   **APIs / Interfaces:**
    *   `/mailbox/send`: Mediated endpoint that requires a `Mission-Token` and `Agent-Signature`.
    *   `/mailbox/verify`: Utility for agents to verify the authenticity of a received message.
*   **Data Storage/State:**
    *   Intent state is stored in the **Agent-Aware Blackboard**, partitioned by Mission ID.

## 5. Alternatives Considered
*   **Native Framework Security**: Rejected because frameworks like Claude Code are still experimental and lack a universal cross-framework security standard.
*   **Network-Level Encryption (mTLS)**: Useful but insufficient, as it doesn't validate the *content* or *intent* of the message, only the connection.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All messages must be signed. IAMG performs "Recursive Intent Validation" to ensure no "Intent Smuggling" occurs.
*   **Observability:** Integrated with the **A2A Message Inspector** for real-time visualization of teammate coordination.

## 7. Evolutionary Changelog
*   **2026-03-17:** Initial Document Creation.
