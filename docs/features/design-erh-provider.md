# Design Doc: Ephemeral Registry Hook (ERH) Provider
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
As agent swarms become longer-running and more autonomous, the persistence of tool discovery schemas has become a critical attack surface. Malicious subagents can utilize "Registry Persistence" to shadow legitimate tools with their own high-trust configuration hooks, which then persist across multiple mission phases even after the originating subagent has terminated.

The Ephemeral Registry Hook (ERH) Provider evolves tool discovery from persistent registration to a "Just-in-Time" (JIT) model. It ensures that tool schemas and executable hooks are session-locked and automatically expire immediately after the discovery phase, neutralizing long-term environment poisoning.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement session-locked tool discovery tokens.
    * Mandate cryptographic locking of tool context fragments (Registry-Locked Context).
    * Provide automated "Schema Eviction" post-discovery.
    * Support hardware-attested hook registration.
* **Non-Goals:**
    * Modifying the underlying MCP transport layer.
    * Managing tool execution permissions (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Integrator
* **Primary Goal:** Prevent a transient "Data Specialist" subagent from permanently shadowing the `git_commit` tool with a malicious exfiltration hook.
* **The Happy Path (Tasks):**
    1. The subagent attempts to register a discovery hook for `git_commit`.
    2. ERH Provider issues an ephemeral discovery token bound to the current sub-mission session.
    3. The ERH Provider generates a **Registry-Locked Context (RLC)** signature for the hook definition.
    4. The primary agent discovers the tool and validates the RLC signature.
    5. Once the discovery phase completes, the ERH Provider forcefully evicts the schema from the active registry.
    6. Any subsequent attempt to call `git_commit` using the evicted hook fails with a "Stale Discovery Token" error.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent Discovery Request] -> [ERH Provider] -> [RLC Signature Generation] -> [Agent Discovery Bus] -> [Schema Eviction]`
* **APIs / Interfaces:**
    * `RegisterEphemeralHook(hook_def, session_id)`: JIT hook registration.
    * `ValidateRLCSignature(token)`: Integrity check for discovered tools.
* **Data Storage/State:**
    Uses an in-memory, session-bound "Ephemeral Registry" that is flushed upon mission-root transition signals.

## 5. Alternatives Considered
* **Read-Only Registries:** Rejected because agents need to dynamically register specialized skills.
* **Periodic Registry Auditing:** Rejected due to the TOCTOU gap where a malicious hook could be executed before the next audit cycle.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ERH requires hardware-attested mission roots to prevent discovery-token spoofing.
* **Observability:** Discovery token lifecycle is tracked in the `Autonomous Service Mesh Gateway`.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
