# Design Doc: Environment-Bound Agency (EBA) Middleware
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
AI agents often require access to host environment variables (API keys, config paths) to perform tasks. However, implicit trust in `process.env` allows for "Token Leakage" exploits where a subagent can be tricked into dumping the entire environment. Existing isolation models are either too coarse (blocking all env vars) or too permissive.

EBA Middleware implements a "Hardware-Attested, Strictly Scoped" model. Agents operate in a virtualized environment where only the minimal set of authorized variables are visible, and each access is validated against a TPM-signed mission manifest.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate per-variable mission-root authorization.
    * Provide hardware-attested proof that an agent's environment was correctly scrubbed.
    * Neutralize side-channel jitter analysis of the scrubbing process via noise injection.
* **Non-Goals:**
    * Managing secret rotation (EBA is an access control and isolation layer).
    * Virtualizing the entire OS (EBA focus is specifically on process-level environment isolation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Orchestrator.
* **Primary Goal:** Run a high-trust "Database Specialist" agent that needs `DB_URL` but must be prevented from seeing `AWS_SECRET_KEY`.
* **The Happy Path (Tasks):**
    1. The User defines a mission manifest for the Database Specialist, explicitly allowing `DB_URL`.
    2. MCP Any spawns the subagent process.
    3. EBA Middleware intercepts the process spawn, scrubs all environment variables except `DB_URL`.
    4. EBA injects "Jitter Noise" into the scrubbing timing to prevent side-channel probing.
    5. The subagent executes with a hardware-attested environment receipt.

## 4. Design & Architecture
* **System Flow:**
    `[Spawn Signal] -> [EBA Scrubber] -> [Jitter Injector] -> [Hardware Enclave (Attestation)] -> [Subagent Process]`
* **APIs / Interfaces:**
    * Internal: `EnforceEnvPolicy(process_id, mission_token)`
* **Data Storage/State:** Environment policies are stored in the hardware-locked Policy Store.

## 5. Alternatives Considered
* **Docker/Containerization:** Rejected as primary solution due to the 500ms+ startup latency tax which stalls high-frequency subagent handoffs. EBA operates at the kernel/process level for sub-millisecond isolation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** EBA treats the host environment as "Untrusted-by-Default" for all subagents.
* **Observability:** Blocked environment access attempts trigger an immediate `SECURITY_ALERT` on the UI Dashboard.

## 7. Evolutionary Changelog
* **2026-07-08:** Initial Document Creation.
