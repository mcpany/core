# Design Doc: Mailbox Injection Shield (MIS)
**Status:** Draft
**Created:** 2026-06-21

## 1. Context and Scope
With the rise of horizontal agent meshes (e.g., Claude Code Agent Teams), the shared teammate mailbox has become a primary coordination vector. Recent research identified "Mailbox Splicing" vulnerabilities where a compromised specialist agent can inject unauthorized task-claiming metadata or instructions into the shared inbox, potentially escalating privileges or hijacking the mission root.

MCP Any needs to provide a secure, hardware-attested coordination bus that ensures message integrity and mission-root alignment. MIS addresses this by mandating cryptographic signatures for every inter-agent message and validating them against a pre-declared mission manifest.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Mandate TPM 2.0 signatures for all inter-agent mailbox messages.
    *   Validate message metadata against hardware-attested "Mission Manifests."
    *   Neutralize "Mailbox Splicing" and instruction-injection attacks.
    *   Provide sub-millisecond validation for high-frequency teammate coordination.
*   **Non-Goals:**
    *   Managing the internal reasoning logic of the agents.
    *   Replacing the transport layer (MIS is a security middleware).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Delegate a sensitive file-write task to a specialist teammate without risking instruction hijacking by a rogue auditor agent.
*   **The Happy Path (Tasks):**
    1.  The primary agent defines a "Mission Manifest" containing authorized tools and teammate IDs.
    2.  The user signs the manifest via a hardware-bound key (TPM).
    3.  A subagent attempts to claim a task from the shared mailbox.
    4.  MIS intercepts the claim, verifies the subagent's identity signature, and checks the claim against the hardware-attested manifest.
    5.  The claim is authorized, and the message is delivered with a "Verified Integrity" token.

## 4. Design & Architecture
*   **System Flow:**
    `Agent A -> [MIS Middleware: Signature Check] -> [MIS Middleware: Manifest Validation] -> Teammate Mailbox -> Agent B`
*   **APIs / Interfaces:**
    *   `POST /mailbox/sign`: Endpoint for agents to request a hardware-attested signature for a message fragment.
    *   `POST /mailbox/verify`: Internal verification hook for incoming coordination messages.
*   **Data Storage/State:**
    *   State is managed in an in-memory "Authenticated Message Ledger," which is periodically flushed to the hardware-attested ARI Hub.

## 5. Alternatives Considered
*   **Software-only JWTs**: Rejected because they are susceptible to local environment exfiltration (e.g., memory scraping). Hardware-bound (TPM) signatures are required for Zero Trust.
*   **Static Manifests**: Rejected because missions evolve. MIS supports "Hierarchical Expansion" where a user can re-attest new branches.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All coordination is "Auth-at-the-Pipe." No message is delivered without a valid hardware signature.
*   **Observability:** MIS logs all verification failures to the "Logic-Grafting Alert Center" for real-time monitoring.

## 7. Evolutionary Changelog
*   **2026-06-21:** Initial Document Creation.
