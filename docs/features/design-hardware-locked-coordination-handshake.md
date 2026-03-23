<!-- markdownlint-disable -->
# Design Doc: Hardware-Locked Coordination Handshake

**Status:** Draft
**Created:** 2026-06-14

## 1. Context and Scope

The emergence of Identity-Decay Attacks (IDA) and session-hijacking
vulnerabilities in the A2A protocol necessitate a transition from software-only
handshakes to hardware-locked sessions. Existing mechanisms like ARI and HAAL
provide semantic and attention-level security, but HLCH establishes a hardware-
bound root of trust for the coordination session itself.

## 2. Goals & Non-Goals

* **Goals:**
    * Mandate hardware-attested (TPM/Secure Enclave) session tokens for all
      inter-agent coordination requests.
    * Bind all coordination fragments (metadata, task bidding) to a unique
      hardware signature.
    * Neutralize Identity-Decay Attacks (IDA) by ensuring subagents cannot mimic
      the parent session's hardware lineage.
* **Non-Goals:**
    * Replacing software-based ARI checks (HLCH is an additional layer).
    * Enforcing local port isolation (handled by BSH Gateway).

## 3. Critical User Journey (CUJ)

* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Secure inter-agent coordination against identity mimicry.
* **The Happy Path (Tasks):**
    1. Agent A initiates a coordination request (e.g., task bidding) to Agent B.
    2. The HLCH Gateway intercepts the request and challenges for a hardware-
       attested session token.
    3. Agent A's local MCP Any instance generates a TPM-signed token bound to
       the current mission root.
    4. The HLCH Gateway verifies the token's lineage and hardware signature.
    5. The handshake is finalized, and coordination proceeds with the hardware
       signature embedded in all subsequent fragments.

## 4. Design & Architecture

* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A] --> B[HLCH Gateway]
        B --> C{TPM Signature Valid?}
        C -- Yes --> D[Finalize Handshake & Sign Fragments]
        C -- No --> E[Block & Alert Sovereignty Monitor]
        D --> F[Agent B]
    ```
* **APIs / Interfaces:**
    * `hlch.VerifyHandshake(sessionToken, missionRoot) -> bool`: Verifies the
      hardware signature and lineage of a coordination request.
    * `hlch.SignFragment(fragment, sessionKey) -> signedFragment`: Binds a
      coordination fragment to the hardware-locked session.
* **Data Storage/State:**
    * **Mission-Root Registry:** A hardware-attested registry of verified
      session keys for the current mission.

## 5. Alternatives Considered

* **Software-Only Signing:** Rejected because software keys are vulnerable to
  exfiltration and stylometric mimicry in long-running sessions.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Root of trust resides in the local hardware (TPM).
* **Observability:** Integrated with the "Sovereignty Monitor" for real-time
  tracking of hardware-locked session health.

## 7. Evolutionary Changelog

* **2026-06-14:** Initial Document Creation.
