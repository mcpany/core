# Design Doc: Entanglement Shard Sanitizer (ESS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In sharded meshes and P2P tunnels (like OpenClaw SNT), agents from different environments synchronize state. This leads to "Contextual Entanglement," where environment-specific data (local paths like `/Users/jules/...`, environment variables, shell history) is accidentally synced into the shared teammate mesh. This "Environmental Residue" is both a security risk and a source of halluncination for teammates in different environments.

The ESS provides a mandatory sanitization layer for all state handoffs, ensuring that shared fragments are semantically cleaned of local residue before synchronization.

## 2. Goals & Non-Goals
* **Goals:**
    * Strip local filesystem paths and environment variables from shared state fragments.
    * Neutralize "Environmental Residue" in binary state handoffs (BSH).
    * Perform high-entropy semantic scanning to detect hidden secrets in coordination messages.
* **Non-Goals:**
    * Sanitizing the internal reasoning monologue (handled by SRM).
    * Modifying the core intent of the state fragment.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Agent Mesh Developer
* **Primary Goal:** Securely share a "Project Context" shard between a local developer agent and a remote specialist agent without leaking the local developer's absolute paths or API keys.
* **The Happy Path (Tasks):**
    1. The local agent writes a state fragment to the sharded mailbox.
    2. The ESS interceptor catches the write event.
    3. ESS performs regex and semantic analysis to identify local residue (e.g., `/home/user`, `SECRET_KEY`).
    4. ESS redacts or normalizes the residue (e.g., replaces `/home/user/repo` with `./`).
    5. The sanitized fragment is synced to the remote peer.

## 4. Design & Architecture
* **System Flow:**
    `Agent A -> [State Fragment] -> ESS (Scrubbing) -> [Sanitized Fragment] -> T2T Bridge -> Agent B`
* **APIs / Interfaces:**
    * `SanitizeFragment(fragment_data, environment_profile)`: Main logic entry point.
* **Data Storage/State:**
    * ESS uses "Environment Profiles" (sets of local paths and variables to block) stored in the session context.

## 5. Alternatives Considered
* **Agent-side Redaction:** Rejected because compromised subagents cannot be trusted to redact themselves.
* **Global Redaction Rules:** Rejected as too blunt; sanitization needs to be context-aware (e.g., don't redact a path that is part of the mission root).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ESS must operate in an isolated sandbox (WASM) to prevent "Sanitizer Hijacking."
* **Observability:** Log redaction events (without logging the raw secrets) to the "Sovereignty Audit Dashboard."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
