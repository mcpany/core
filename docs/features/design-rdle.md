# Design Doc: Recursive Depth-Limit Enforcer (RDLE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Tracking delegation depth to prevent CVE-2026-71001 shadow handoffs.

## 2. Goals & Non-Goals
* **Goals:** Track depth via headers, enforce hard limits.
* **Non-Goals:** Restricting horizontal width.

## 3. Critical User Journey (CUJ)
* **Persona:** Security Officer
* **Goal:** Prevent >3 levels of delegation.
* **Happy Path:** User sets limit, RDLE appends signed marker, Level 4 blocked.

## 4. Design & Architecture
Hardware-attested headers (X-UAB-Delegation-Depth).

## 5. Alternatives Considered
Stateless tokens (strippable).

## 6. Cross-Cutting Concerns
Security via SMI Relay.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
