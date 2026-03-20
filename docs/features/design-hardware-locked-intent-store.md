# Design Doc: Hardware-Locked Intent Store (HLIS)
**Status:** Draft
**Created:** 2026-05-16

## 1. Context and Scope
As AI agent swarms (e.g., Claude Code Agent Teams, OpenClaw swarms) become more autonomous and handle high-stakes tasks, the risk of "Intent Hijacking" increases. A compromised subagent or a malicious tool can attempt to steer the swarm away from its original "Mission-Root" goals via context injection or semantic manipulation.

MCP Any needs a mechanism to protect the core mission intent in a way that is resistant to software-level tampering. HLIS solves this by cryptographically anchoring mission intents in secure hardware (TPM/Secure Enclave), ensuring that the fundamental goals of an agent session remain immutable and verifiable throughout the swarm's lifecycle.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a hardware-backed (TPM/SEP) storage for mission-root intent strings.
    *   Generate cryptographic "Intent Proofs" that can be verified by downstream tools and agents.
    *   Ensure that "Mission-Root" intents cannot be overwritten without explicit, hardware-bound user re-attestation.
    *   Provide an API for subagents to retrieve (but not mutate) the hardware-locked intent.
*   **Non-Goals:**
    *   Storing high-frequency, transient conversation state (handled by the Blackboard).
    *   Performing full agent reasoning within the secure enclave.
    *   Replacing software-level policy engines (HLIS acts as the immutable anchor for those policies).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Ensure that an autonomous code-refactoring swarm cannot be coerced into exfiltrating source code, even if a subagent's memory is poisoned.
*   **The Happy Path (Tasks):**
    1.  The user initiates a mission with a signed "Root Intent" (e.g., "Refactor module X without making network calls").
    2.  MCP Any commits this intent to the HLIS, binding it to a hardware-backed session key.
    3.  A subagent is spawned and receives a "Lease" cryptographically linked to the HLIS intent.
    4.  The subagent attempts to call a tool that performs an outbound network request.
    5.  The Policy Firewall verifies the tool call against the HLIS-anchored intent and blocks the request, as it diverges from the immutable goal.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        User[User/Lead Agent] -- Signed Intent --> HLIS[HLIS Middleware]
        HLIS -- Store --> Enclave[Secure Enclave / TPM]
        Subagent[Subagent] -- Request Proof --> HLIS
        HLIS -- Sign Proof --> Subagent
        Subagent -- Proof + Tool Call --> Policy[Policy Firewall]
        Policy -- Verify Proof against Intent --> HLIS
        HLIS -- Valid/Invalid --> Policy
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/hlis/anchor`: Anchor a new mission-root intent. Requires hardware signature.
    *   `GET /v1/hlis/proof`: Retrieve a hardware-signed proof of the current mission intent.
    *   `internal: hlis.VerifyIntent(proof, intent)`: Kernel-level function to verify a proof.
*   **Data Storage/State:**
    *   Intents are stored in hardware NVRAM or sealed within the Enclave's memory.
    *   Session keys are ephemeral and hardware-bound.

## 5. Alternatives Considered
*   **Software-Only Immutable Store**: Rejected because a root-level exploit on the host could still modify the intent in memory.
*   **Blockchain-Anchored Intents**: Rejected due to latency and cost overhead for local execution environments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** HLIS forms the "Trust Root" for the session. All other security layers (Policy Firewall, RAMS) derive their authority from the hardware-locked intent.
*   **Observability:** All HLIS access events (anchoring, proof generation, verification failures) are logged to the Hardware-Attested Audit Log.

## 7. Evolutionary Changelog
*   **2026-05-16:** Initial Document Creation.
