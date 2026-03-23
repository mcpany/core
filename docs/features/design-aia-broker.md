# Copyright 2026 Author(s) of MCP Any

# SPDX-License-Identifier: Apache-2.0

# Design Doc: Active Intent Alignment (AIA) Broker

**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope

The AIA Broker ensures that specialist agent reasoning remains mission-anchored
by issuing hardware-attested heartbeats.

## 2. Goals & Non-Goals

* **Goals:**
    * Periodically verify specialist alignment with mission-root intent.
    * Neutralize cumulative drift in valid reasoning chains.
* **Non-Goals:**
    * Replacing real-time entropy monitoring (handled by CEC).

## 3. Critical User Journey (CUJ)

* **User Persona:** Lead Systems Architect.
* **Primary Goal:** Ensure multi-hop delegations remain within intent bounds.
* **The Happy Path (Tasks):**
    1. Parent agent delegates task to specialist via AIA Broker.
    2. AIA Broker attaches a hardware-attested alignment heartbeat requirement.
    3. Specialist provides periodic alignment proofs during reasoning.
    4. Broker validates proofs against mission-root signature.

## 4. Design & Architecture

* **System Flow:** `Specialist` -> `AIA Proof` -> `Broker` -> `Attestation`
* **APIs / Interfaces:** `POST /v1/intent/align`, `GET /v1/intent/status`
* **Data Storage/State:** Proofs persisted in session state metadata.

## 5. Alternatives Considered

* **Manual Review:** Rejected for machine-speed swarms.

## 6. Cross-Cutting Concerns

* **Security:** Cryptographically binds proof to specific mission fragments.
* **Observability:** Metrics on alignment success rates.

## 7. Evolutionary Changelog

* **2026-06-18:** Initial Document Creation.
