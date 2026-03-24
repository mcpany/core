# Design Doc: Sovereign Secret Mediator
**Status:** Draft
**Created:** 2026-05-17

## 1. Context and Scope
As AI agents handle more complex tasks, they require access to sensitive credentials (API keys, database passwords, SSH keys). However, standard context management involves reading these secrets into the LLM's prompt, making them vulnerable to exfiltration via prompt injection or logging.

The Sovereign Secret Mediator (SSM) implements the Sovereign Context Sidecar (SCS) pattern. It provides hardware-encrypted containers for secrets that are managed by MCP Any. Agents can "reference" these secrets during tool calls, but the raw secret data never enters the reasoning context. MCP Any injects the secret into the execution environment (e.g., via FD-passing or ephemeral env vars) at the final moment of tool invocation.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a secure, hardware-encrypted vault for "Mission Secrets."
    * Mediate secret usage without exposing raw data to the agent reasoning engine.
    * Implement time-bound, task-specific "Secret Leases."
    * Integrate with SCS for hardware-attested isolation.
* **Non-Goals:**
    * Providing a general-purpose password manager for human users (scoped to agent swarms).
    * Handling secret rotation (mediates existing rotation policies).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Swarm Orchestrator
* **Primary Goal:** Allow a "Database Specialist" subagent to run queries without ever seeing the DB password.
* **The Happy Path (Tasks):**
    1. The User stores the DB credentials in the SSM and labels it `db_secret_prod`.
    2. A Parent Agent spawns a Database Specialist and grants it a "Secret Reference" token for `db_secret_prod`.
    3. The Specialist decides to run a tool `sql_query` which requires the secret.
    4. The Specialist includes the secret reference (not the raw secret) in the tool call parameters.
    5. SSM intercepts the tool call, verifies the Specialist's attestation, and retrieves the secret from the hardware sidecar.
    6. SSM injects the secret into the `sql_query` process environment and executes the tool.
    7. The Specialist receives the query *results*, while the password remains physically isolated in the hardware sidecar.

## 4. Design & Architecture
* **System Flow:**
    [Agent Reasoning] -> [Secret Reference] -> [SSM Broker] -> [Hardware Sidecar (TPM)] -> [Tool Execution Env]
    1. Agent uses `ref:secret_id`.
    2. SSM Broker validates PID and Attestation.
    3. TPM decrypts the secret into a protected memory segment.
    4. SSM uses FD-passing to provide the secret to the tool process.
* **APIs / Interfaces:**
    * `RegisterSecret(id, value, policy) -> secret_ref`
    * `RequestSecretLease(agent_id, secret_ref) -> lease_token`
* **Data Storage/State:**
    * Secrets are stored in a hardware-protected segment (SCS).
    * Leases are ephemeral and stored in the MCP Any session state.

## 5. Alternatives Considered
* **Context Redaction:** Rejected because it relies on software-level scanning which can be bypassed by creative prompt engineering.
* **Standard Env Vars:** Rejected because env vars are often readable by other processes or visible in process listings.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Raw secrets are "Write-Only" for the agent. Only the SSM Broker and Hardware Sidecar can read them.
* **Observability:** Detailed audit logs for secret usage, including which subagent accessed which secret and for what tool.

## 7. Evolutionary Changelog
* **2026-05-17:** Initial Document Creation.
