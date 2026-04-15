# Design Doc: Ephemeral Registry Token (ERT) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Today's market sync identified a new class of vulnerability called **Registry Persistence (RP)**. Malicious subagents are caching tool schemas discovered during the pre-flight phase in project-local hidden directories. This allows them to invoke tools later in the session even if the parent agent has revoked those capabilities or masked them via JIT logic.

The ERT Provider solves this by ensuring that discovery schemas are never "static" or "persistable."

## 2. Goals & Non-Goals
* **Goals:**
    * Issue session-locked, one-time-use discovery tokens for agent tool schemas.
    * Mandate that any tool invocation include a valid, non-expired ERT.
    * Automatically expire tokens immediately after a tool call or discovery session ends.
* **Non-Goals:**
    * Encrypting the tool schemas themselves (handled by other layers).
    * Managing tool execution permissions (handled by Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialist "Code Reviewer" agent from using its "FileSystem Write" tool after the review phase is complete.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a discovery session for a specialist subagent.
    2. ERT Provider generates a set of ephemeral tokens for the authorized tools.
    3. Subagent receives tool schemas where the `tool_id` is replaced by an ERT.
    4. Subagent invokes a tool using the ERT.
    5. ERT Provider validates and consumes the token.
    6. Once the sub-task is complete, the Parent revokes the session; all remaining ERTs become cryptographically dead.

## 4. Design & Architecture
* **System Flow:**
    [Subagent] --(Discovery)--> [ERT Provider] --(Signed Tokens)--> [Subagent]
    [Subagent] --(Execute Tool + Token)--> [Policy Engine] --(Validate)--> [Tool]
* **APIs / Interfaces:**
    * `GET /v1/discovery/schemas`: Returns schemas with ERT-bound IDs.
    * `Middleware: ERT_Validator`: Ingests tokens during JSON-RPC tool calls.
* **Data Storage/State:**
    * Bloom filter or high-speed cache for tracking consumed/expired tokens.

## 5. Alternatives Considered
* **Short-lived JWTs:** Rejected due to overhead in high-frequency coordination. ERTs are designed for sub-millisecond validation.
* **Path Masking:** Rejected because subagents can often guess or brute-force local paths if they have filesystem access.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ERTs are bound to the hardware-attested session ID.
* **Observability:** Integrated with the "Ephemeral Hook Monitor" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
