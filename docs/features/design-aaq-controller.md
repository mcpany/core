# Design Doc: Adaptive Attestation Quota (AAQ) Controller
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move toward high-frequency, autonomous coordination, the overhead of per-call hardware attestation (e.g., TPM/SEP signatures) is becoming a critical performance bottleneck. Today's research confirms the emergence of "Attestation Exhaustion," where agents are "stalled" by the 100ms+ latency of cryptographic signing during tight reasoning loops.

The AAQ Controller evolves the MCP Any security layer to dynamically scale the frequency of hardware attestation. By shifting from a "Per-Call" to an "Adaptive Trust Lease" model, we can maintain mission-root sovereignty while achieving sub-millisecond coordination for low-risk operations.

## 2. Goals & Non-Goals
* **Goals:**
    * Implementing dynamic scaling of hardware attestation frequency based on real-time risk scoring.
    * Utilizing hardware-bound monotonic counters to bridge trust windows between full signatures.
    * providing "Attestation Bursts" for high-stakes tool sequences.
* **Non-Goals:**
    * replacing hardware attestation entirely (it remains the root of trust).
    * Bypassing user-mandated HITL policies for high-risk actions.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Execute a sequence of 50 local filesystem reads and internal state updates in under 500ms without compromising attestation strength.
* **The Happy Path (Tasks):**
    1. The orchestrator agent initiates a "Low-Risk Burst" session with the AAQ Controller.
    2. The AAQ Controller issues a "Trust Lease" backed by a single hardware signature and a TPM-signed monotonic counter.
    3. The agent executes 50 tool calls, with each call verified via the monotonic counter (latency: <5ms).
    4. Upon task completion or encountering a high-risk tool (e.g., `rm -rf`), the AAQ Controller mandates a fresh hardware signature.
    5. The mission-root attestation is reconciled and finalized.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `AAQ Middleware` -> `Trust Lease Broker` (Check Lease) -> `Hardware Root` (If Expired) -> `Verification`.
* **APIs / Interfaces:**
    * `/v1/aaq/lease`: Request a time-bound or call-bound trust lease.
    * `x-aaq-counter`: Header for passing monotonic attestation tokens.
* **Data Storage/State:**
    * Lease states are stored in non-persistent, kernel-protected memory.
    * Risk scores are calculated in real-time by the `Policy Engine`.

## 5. Alternatives Considered
* **Persistent Trust**: Rejected as it fails to account for subagent compromise during the session; leases must be "ephemeral and revocable."
* **Software-only signatures**: Rejected due to lack of non-repudiation; hardware-bound root is mandatory for Universal Agent Bus compliance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Trust History" of an agent influences its AAQ quota; a single failed attestation results in immediate revocation of all active leases.
* **Observability:** "Attestation Latency Tax" and "Lease Efficiency" metrics are tracked to optimize quota thresholds.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
