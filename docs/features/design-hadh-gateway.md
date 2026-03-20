# Design Doc: Hardware-Attested Discovery Handshake (HADH)
**Status:** Draft
**Created:** 2026-05-30

## 1. Context and Scope
With the maturation of Gemini CLI's A2A auth and the shift toward horizontal "Agent Teams," the discovery phase has become a primary attack vector. Unauthenticated agents can "probe" the capabilities of local and remote peers, leading to "Capability Shadowing" and sensitive metadata exfiltration. HADH mandates a cryptographically bound, hardware-attested handshake *before* any tool or capability is revealed to a peer agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Encrypt the "Agent Card" (capability list) until a handshake is completed.
    * Require TPM/Secure Enclave signatures for identity verification during discovery.
    * Neutralize "Pre-Flight Shadow Mapping" by rogue subagents.
    * Support session-bound "Trust Leases" to minimize handshake latency.
* **Non-Goals:**
    * Managing the execution security of the discovered tools (handled by the Policy Firewall).
    * Providing a global directory service (focus is on peer-to-peer discovery).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local-First Agent Swarm Developer
* **Primary Goal:** Ensure that a newly spawned "Security Auditor" agent can only see the tools of the "Database specialist" after proving its hardware-attested lineage.
* **The Happy Path (Tasks):**
    1. Agent A (Requester) broadcasts a discovery request to the Mesh.
    2. Agent B (Provider) responds with a "Handshake Challenge" signed by its local TPM.
    3. Agent A signs the challenge with its own hardware key and includes its "Mission Root Token."
    4. MCP Any validates the cross-signatures and mission-alignment.
    5. MCP Any decrypts and reveals Agent B's capability card to Agent A for the duration of the mission.

## 4. Design & Architecture
* **System Flow:**
    `[Requester] --(Discovery Req)--> [HADH Gateway] --(Challenge)--> [Requester] --(Signed Proof)--> [HADH Gateway] --(Agent Card)--> [Requester]`
* **APIs / Interfaces:**
    * `hadh.v1.InitiateHandshake(peer_id, mission_token)`
    * `hadh.v1.VerifyIdentity(signature, cert_chain)`
* **Data Storage/State:**
    * Ephemeral "Handshake Sessions" stored in the Blackboard with a 300s TTL.
    * Persistent "Trusted Lineage" cache to accelerate repeat handshakes.

## 5. Alternatives Considered
* **TLS-Only Auth:** Rejected because it doesn't provide "Lineage Proof"--an attacker could still spoof a valid cert if the local environment is partially compromised.
* **Public Discovery + Per-Call Auth:** Rejected because metadata (tool descriptions) itself often contains sensitive intent fragments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HADH is the "Zero-th" gate of discovery. No "Shadowing" is possible as capabilities are cryptographically invisible to un-attested peers.
* **Observability:** Integration with the `Mesh Discovery Handshake Monitor` for real-time visualization of auth events.

## 7. Evolutionary Changelog
* **2026-05-30:** Initial Document Creation.
