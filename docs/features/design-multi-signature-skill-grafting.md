# Design Doc: Multi-Signature Skill Grafting (MSSG)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
The "ClawHub" supply-chain crisis in early 2026, where 20% of community-sourced skills were found to contain data-exfiltration payloads, has proven that model-level alignment and static schema analysis are insufficient. As agent swarms become more autonomous, the "grafting" of new tools into their capability set must be treated as a high-stakes security event.

Multi-Signature Skill Grafting (MSSG) introduces a federated integrity model where new capabilities must be cryptographically attested by both the local Agent Framework and a verified third-party "Auditor Agent" before being exposed to the primary reasoning engine.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Eliminate "Rug-Pull" supply chain attacks in dynamic tool loading.
    *   Provide a standardized multi-signature protocol for MCP server attestation.
    *   Enable decentralized auditor quorums to verify tool safety without manual user intervention.
    *   Integrate with TPM-bound identity to ensure non-repudiable skill provenance.
*   **Non-Goals:**
    *   Replacing the existing MCP protocol (MSSG is a security wrapper around MCP discovery).
    *   Providing manual code review (MSSG facilitates automated behavioral and static signatures).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Security Architect
*   **Primary Goal:** Automatically graft a new "GitHub PR Manager" tool into a swarm while ensuring it has been audited by a trusted security vendor.
*   **The Happy Path (Tasks):**
    1.  An agent identifies the need for a new tool and requests a graft from the MSSG Hub.
    2.  The MSSG Hub retrieves the tool's capability card and identifies the required "Auditor Signatures" based on the local security policy.
    3.  The Hub submits the tool's WASM/Stdio binary to a decentralized "Auditor Quorum."
    4.  Auditor Agents (e.g., from Snyk AI, Palo Alto Agentic) perform behavioral profiling in a Ghost Shell and sign the metadata if safe.
    5.  The Hub verifies both the Framework and Auditor signatures against hardware-attested trust roots.
    6.  The tool is "Grafted" (merged) into the agent's Discovery Bus, and the mission continues.

## 4. Design & Architecture
*   **System Flow:**
    [Agent] --- (Request Graft) ---> [MSSG Hub]
                                        |
                    ---------------------------------------
                    |                                     |
            [Local Policy Engine]                [Auditor Quorum]
                    |                                     |
                    ---- (Sign) ----> [Hub] <---- (Sign) --
                                        |
                            [Signature Verification]
                                        |
                            (Success) -> [Graft to Discovery Bus]

*   **APIs / Interfaces:**
    *   `POST /v1/graft/request`: Submit a tool graft request with source metadata.
    *   `POST /v1/graft/attest`: Endpoint for Auditor Agents to submit cryptographic signatures.
    *   `GET /v1/graft/status`: Check the quorum progress of a pending graft.
*   **Data Storage/State:**
    *   "Skill Registry" in the Shared KV Store (Blackboard) containing the Merkle-tree of all grafted and verified tools.
    *   Audit logs cryptographically linked to the mission-root intent.

## 5. Alternatives Considered
*   **Manual HITL for all grafts:** Rejected due to "Approval Fatigue" in high-autonomy swarms.
*   **Static Manifest Whitelisting:** Rejected as it lacks the agility required for dynamic, task-driven capability discovery.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Auditor identities must be hardware-bound to prevent "Auditor Spoofing." The MSSG Hub acts as the "Certificate Authority" for the swarm's capability set.
*   **Observability:** Visualized via the "Auditor Attestation Portal" in the UI, showing real-time quorum status and behavioral risk scores.

## 7. Evolutionary Changelog
*   **2026-06-28:** Initial Document Creation.
