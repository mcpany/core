# Design Doc: ACR Hub Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Granular revocation of capabilities in real-time for deep agent swarms.

## 2. Goals & Non-Goals
* **Goals:** Real-time revocation, cross-framework sync.
* **Non-Goals:** Long-term identity management.

## 3. Critical User Journey (CUJ)
* **Persona:** Swarm Orchestrator
* **Goal:** Revoke specific tool access for search subagent.
* **Happy Path:** Supervisor sends revoke, Hub invalidates token, Gateway
blocks.

## 4. Design & Architecture
Middleware for A2A Bridge using TPM-signed manifests.

## 5. Alternatives Considered
Session termination (too heavy).

## 6. Cross-Cutting Concerns
Security (Zero Trust) via TPM.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
