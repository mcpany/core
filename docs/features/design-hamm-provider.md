# Design Doc: Hardware-Attested Mission Manifest (HAMM) Provider
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms evolve from single-session tasks to horizontal teammate coordination (e.g., Claude Code Agent Teams) and persistent background infrastructures (OpenClaw), the risk of "Intent Ghosting" and "Teammate State-Splicing" has increased. Current security models often rely on session-level trust, which is insufficient for long-running or deeply nested swarms.

The Hardware-Attested Mission Manifest (HAMM) Provider serves as the "Manifest Mint" for MCP Any. It allows a lead agent to pre-declare the absolute boundaries of a mission—including authorized tools, allowed state shards, and subagent roles—and cryptographically bind this declaration to a hardware-attested (TPM) signature. This ensures that every member of the swarm operates within an immutable, verified perimeter.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide an authoritative service for issuing TPM-signed "Mission Manifests."
    * Enable pre-declaration of tool and capability sets for all subagents/teammates.
    * Integrate with the Mission-Locked Execution (MLE) Gateway for real-time enforcement.
    * Neutralize "Discovery-Phase Shadowing" by making unauthorized capabilities invisible to agents not listed in the manifest.
* **Non-Goals:**
    * Executing the tool calls themselves (this is the responsibility of the MLE Gateway).
    * Providing general-purpose hardware attestation for the host OS (only for mission-specific manifests).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (Lead Agent)
* **Primary Goal:** Securely spawn 3 specialized teammates for a code migration task while ensuring they cannot access production credentials or external networks not required for the task.
* **The Happy Path (Tasks):**
    1. Lead Agent defines a mission scope (Tool IDs, State Shard IDs, Teammate Roles).
    2. Lead Agent requests a Mission Manifest from the HAMM Provider.
    3. HAMM Provider validates the request against the current hardware-attested session.
    4. HAMM Provider generates a manifest and signs it using the local TPM.
    5. The resulting HAMM token is distributed to all spawned teammates.
    6. Teammates use the HAMM token to prove their authority to the MLE Gateway.
    7. Any attempt to use a tool or access a shard not in the manifest is immediately blocked by the gateway.

## 4. Design & Architecture
* **System Flow:**
    `[Lead Agent] -> (Mission Params) -> [HAMM Provider] -> (Sign Request) -> [TPM] -> (HAMM Token) -> [Teammates]`
* **APIs / Interfaces:**
    * `hamm.IssueManifest(missionScope) -> (HAMMToken, error)`: Generates and signs a new manifest.
    * `hamm.ValidateToken(token) -> (bool, error)`: Verifies a token's cryptographic integrity.
* **Data Storage/State:**
    * **Manifest Registry:** A volatile, memory-resident store of active mission manifests for the current hardware session.

## 5. Alternatives Considered
* **Software-Only Manifests:** Rejected because they are vulnerable to local privilege escalation or "Ghost" agents bypassing the gateway.
* **Per-Call Attestation:** Rejected as the primary path due to the "Attestation Tax" (latency) in high-frequency horizontal coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The HAMM Provider must be the only service capable of requesting mission-root signatures from the TPM.
* **Observability:** Manifest issuance and validation failures are logged in the "Sovereignty Audit Log."

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
