# Design Doc: Atomic Handoff Monotonicity (AHM) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the release of Sovereign Node Tunneling (SNT) in OpenClaw, agent swarms can now span multiple physical devices. However, "Tunnel-Racing" vulnerabilities have emerged, where a malicious node can intercept a P2P handshake by spoofing a high-frequency node-switch event.

The AHM Provider is designed to neutralize racing attacks by binding every inter-node handoff to a hardware-attested, monotonic counter. This ensures that session resumes are only valid if they strictly follow the chronological sequence of node transitions.

## 2. Goals & Non-Goals
* **Goals:**
    * Anchor mesh handshakes to monotonic hardware counters (TPM-based).
    * Prevent "Tunnel-Racing" and handshake-replay attacks in P2P tunnels.
    * Maintain sub-millisecond latency for fast-path resumption.
* **Non-Goals:**
    * Managing the underlying P2P network topology (handled by AMT Broker).
    * Providing end-to-end encryption for the tunnel data (handled by T2T Encryption).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Agent Swarm
* **Primary Goal:** Securely resume a session when a specialist agent migrates from a local laptop node to a high-compute desktop node.
* **The Happy Path (Tasks):**
    1. Laptop node initiates a handoff to the Desktop node.
    2. Laptop node requests a monotonic "Handoff Token" from its local AHM Provider.
    3. The token is incremented and signed by the laptop's TPM.
    4. Laptop sends the token and session ticket to the Desktop node via the P2P tunnel.
    5. Desktop node's AHM Provider verifies the token sequence and TPM signature.
    6. Session is resumed atomically; any out-of-order racing attempts from other nodes are rejected.

## 4. Design & Architecture
* **System Flow:**
    [Laptop AHM] --(Signed Token N+1)--> [Desktop AHM] --(Validation)--> [Session Resumed]
* **APIs / Interfaces:**
    * `GetHandoffToken()`: Issues a new monotonic token for a session.
    * `VerifyHandoffSequence(token, sender_identity)`: Validates the incoming token.
* **Data Storage/State:**
    * Monotonic counters are maintained in the TPM's NVRAM for each verified session.

## 5. Alternatives Considered
* **Timestamp-based Validation:** Rejected because local clock drift across nodes allows for racing windows that attackers can exploit.
* **Global Coordination Lock:** Rejected because it introduces a single point of failure and unacceptable latency for high-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If the monotonic counter is exhausted or tampered with, the AHM Provider triggers a full mission-root re-attestation.
* **Observability:** Node transition counts are exported to the System Health Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
