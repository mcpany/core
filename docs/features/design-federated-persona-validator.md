# Design Doc: Federated Persona Validator (SPaaS)
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
Claude Code's "System-Prompt-as-a-Service" (SPaaS) allows agents to load complex personas from remote URLs. This introduces the risk of "Persona Injection," where a malicious persona can redefine core safety guardrails or inject hidden instructions that affect all subsequent sub-intents. The Federated Persona Validator acts as a secure gate for SPaaS, ensuring that remote personas are attested and conform to local security policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate the integrity and provenance of remote persona files (system prompts).
    * Scan personas for imperative instructions that conflict with the "Universal Governance Layer."
    * Enforce "Persona Purity" - preventing a subagent from adopting a more permissive persona than its parent.
* **Non-Goals:**
    * Rewriting the persona content (sanitization is preferred over editing).
    * Managing the storage of the persona files themselves.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Admin
* **Primary Goal:** Ensure that no agent in the organization can be "re-programmed" by a malicious remote system prompt.
* **The Happy Path (Tasks):**
    1. Agent attempts to load a persona from a remote SPaaS provider.
    2. MCP Any intercepts the load request.
    3. The Persona Validator fetches the content and checks for a valid VML (Verified Metadata Lineage) signature.
    4. The validator performs a "Structural Audit" to ensure no "Override" instructions are present.
    5. If valid, the persona is applied to the agent session; otherwise, it is blocked.

## 4. Design & Architecture
* **System Flow:**
    `SPaaS Request` -> `Attestation Check` -> `Structural Audit` -> `Policy Matching` -> `Persona Injection`
* **APIs / Interfaces:**
    * `PersonaValidationProvider`: Interface for checking persona integrity.
* **Data Storage/State:**
    * Cache of "Safe Personas" and "Blocked Signatures."

## 5. Alternatives Considered
* **Local-Only Personas**: Rejected as it limits the flexibility of multi-agent swarms that need specialized (but safe) personas.
* **Prompt Injection Defense**: Traditional PI defense often misses structural persona changes that redefine the "Rules of the Game."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory attestation for all remote personas.
* **Observability:** Logs all persona load attempts and audit failures.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
