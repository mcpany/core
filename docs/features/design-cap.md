# Design Doc: Continuous Attestation Pulse (CAP)

**Status:** Draft
**Created:** 2026-06-21

## 1. Context and Scope
The discovery of "Handshake Hijacking" exposes a weakness in one-time attestation. CAP addresses this by introducing a high-frequency, hardware-bound "security heartbeat" that verifies identity throughout the execution lifecycle.

## 2. Goals & Non-Goals
* **Goals:**
  * Implement a background "attestation pulse" using TPM.
  * Bind individual tool calls to a specific, time-limited pulse token.
  * Terminate reasoning session if the pulse is interrupted.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Swarm Developer
* **Primary Goal:** Prevent a hijacked subagent from ghost-riding a verified session.
* **The Happy Path (Tasks):**
  1. Agent starts reasoning session and performs HAIL-handshake.
  2. CAP middleware initiates the pulse.
  3. Agent attempts a tool call.
  4. MCP Any verifies the tool call carries the *current* pulse token.
  5. Tool call authorized.

## 4. Design & Architecture
* **System Flow:** Session Start -> HAIL Init -> CAP Pulse -> [Tool Call -> Verify Pulse -> Execution].
* **APIs / Interfaces:** `POST /v1/cap/pulse/status`, `POST /v1/cap/verify`.

## 5. Alternatives Considered
* **Short-Lived JWTs:** Rejected as they are still susceptible to hijacking on host.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Pulse failure triggers immediate revocation.
* **Observability:** Pulse health visible in the Identity Pulse Monitor.

## 7. Evolutionary Changelog
* **2026-06-21:** Initial Document Creation.
