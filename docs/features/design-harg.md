# Design Doc: Hardware-Attested Resumption Guard (HARG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Attested Mesh Tunneling (AMT) and Fast-Path Identity Resumption (FPIR), the "Universal Agent Bus" has gained significant performance in multi-node swarms. However, the "TunnelCrack" exploit has revealed that unauthenticated local processes can intercept and spoof P2P handshake signals, allowing for unauthorized tool invocation.

The Hardware-Attested Resumption Guard (HARG) is a core security service that mandates the use of TPM-bound monotonic counters for all resumption handshakes. It ensures that every "Fast-Path" resumption signal is cryptographically unique and non-reusable, invalidating all previous session-bound trust fragments upon any rotation or node-tunnel event.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement TPM-bound monotonic counters for inter-agent and inter-node resumptions.
    * Mandate hardware re-attestation for any resumption signal exceeding the 50ms "Trust Lease" window.
    * Automatically invalidate stale session tokens upon detection of a monotonic counter jump.
    * Neutralize "TunnelCrack" spoofing by binding handshakes to hardware lineage.
* **Non-Goals:**
    * Replacing TLS/mTLS for transport layer security.
    * Managing individual tool capability scopes (handled by EPM/RBF).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Agent Orchestrator
* **Primary Goal:** Securely resume a distributed research session across a desktop and a mobile device.
* **The Happy Path (Tasks):**
    1. Agent on Node A initiates a P2P tunnel to Node B.
    2. HARG issues a hardware-locked resumption token linked to the current TPM monotonic state.
    3. Session is suspended and then resumed 2 seconds later.
    4. Upon resumption, HARG verifies the monotonic counter increment.
    5. Previous handshake fragments are marked as "Spent" in the mesh registry.
    6. Secure tunnel is resumed with sub-millisecond overhead using the verified hardware lineage.

## 4. Design & Architecture
* **System Flow:**
    `[Resumption Request] -> [HARG Monotonic Validator] -> [Hardware Attestation Provider] -> [Mesh Token Mint] -> [Tunnel Resumption]`
* **APIs / Interfaces:**
    * `harg.IssueResumptionToken(node_id)`: Returns a TPM-signed monotonic token.
    * `harg.ValidateLineage(token)`: Verifies hardware-locked sequence integrity.
* **Data Storage/State:** HARG maintains a "Resumption Sequence Log" in the secure enclave metadata.

## 5. Alternatives Considered
* **Session-Bound Tokens (JWT)**: Rejected because they are susceptible to interception and replay within the session window (the root cause of TunnelCrack).
* **Mandatory Full Handshakes**: Rejected due to high latency (100ms+) impacting horizontal swarm coordination speed.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Counter hijacking is prevented by hardware-level protections on the monotonic counter.
* **Observability:** Resumption failures and "Spent Token" re-use attempts are logged in the Mesh Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
