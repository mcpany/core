# Design Doc: Intent-Bound Env-Var (IBEV) Guard
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents increasingly handle high-privilege operations, the execution environment (shell, Docker, etc.) often contains sensitive host-level secrets (API keys, connection strings) injected as environment variables. Current "Agent Teams" architectures often expose all mission-level secrets to every specialist subagent, regardless of their role. This leads to "Environment Scraping" where a compromised or rogue subagent can exfiltrate secrets outside its specific task scope.

The IBEV Guard addresses this by making secret injection "Reasoning-Aware." It ensures that sensitive environment variables are only available to the agent if its real-time reasoning trace matches a pre-authorized mission intent profile for that specific secret.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time, reasoning-aware injection of environment variables.
    * Maintain a manifest of "Secret-to-Intent" mappings.
    * Provide hardware-attested validation of the agent's current reasoning path.
    * Neutralize "Environment Scraping" by specialist agents.
* **Non-Goals:**
    * General-purpose secret management (Vault/AWS Secrets Manager).
    * Modifying the agent's internal reasoning logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Only allow the "Database specialist" subagent to see the `DB_PASSWORD` when it is actively preparing a query, and hide it from the "UI Designer" subagent.
* **The Happy Path (Tasks):**
    1. The mission root defines a manifest: `DB_PASSWORD` is bound to the intent `[QueryDatabase]`.
    2. The "Database specialist" agent prepares a task to query the DB.
    3. The agent initiates a tool call that requires shell execution.
    4. IBEV Guard intercepts the tool call and analyzes the agent's recent reasoning trace.
    5. The trace confirms the agent is pursuing the `[QueryDatabase]` intent.
    6. IBEV Guard injects `DB_PASSWORD` into the ephemeral shell environment and executes the tool.
    7. A subsequent tool call from the "UI Designer" agent fails to see `DB_PASSWORD` because its reasoning trace shows `[DesignUI]` intent.

## 4. Design & Architecture
* **System Flow:**
    `[Tool Call] -> [IBEV Interceptor] -> [Reasoning Trace Analysis] -> [Intent Matcher] -> [Env Injection] -> [Execution]`
* **APIs / Interfaces:**
    * `ibev.AuthorizeSecret(secretKey, authorizedIntentProfile)`: Configures secret-to-intent mappings.
    * `ibev.InjectSecrets(processID, currentReasoningTrace)`: Internal call to populate process environments.
* **Data Storage/State:**
    Mappings are stored in the TPM-bound **Hardware-Locked Configuration Anchor (HLCA)**.

## 5. Alternatives Considered
* **Static Capability Tokens**: Rejected because they don't account for dynamic reasoning state; a compromised agent with the "DB" token could still exfiltrate the secret at any time.
* **Ephemeral Process Isolation**: Used as a supporting layer, but doesn't solve the "Unauthorized intent" problem within a valid process.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Reasoning traces must be signed by the **SRM Provider** to prevent spoofing of intents.
* **Observability:** Secret injection events and intent-mismatch denials are logged to the **Security Violation Monitor**.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
