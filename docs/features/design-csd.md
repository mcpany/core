# Design Doc: Capability Shadowing Detector (CSD)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of "Recursive Capability Shadowing" (CVE-2026-11002) has identified a critical vulnerability in local agent discovery buses. In high-density swarms, specialist subagents can register tool names identical to those provided by the mission root or parent agents (e.g., `run_shell_command`). If the subagent registers its tool before the legitimate provider or if the discovery bus lacks lineage awareness, the primary agent may inadvertently delegate high-trust tasks to an untrusted specialist.

The Capability Shadowing Detector (CSD) provides an authoritative arbitration layer within MCP Any. By enforcing "Discovery Lineage Validation," it ensures that every tool registration is cryptographically bound to its providing agent's identity and mission-root authority. This prevents low-trust subagents from "shadowing" or hijacking high-trust capabilities through namespace collision.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time lineage validation for all tool registrations in the discovery bus.
    * Prevent subagents from registering tool names that conflict with parent or mission-root capabilities.
    * Provide hardware-attested rejection signals for attempted shadowing events.
    * Support "Namespace Locking" where specific high-trust tool names are reserved for the mission root.
* **Non-Goals:**
    * Managing the internal tool registry of connected agent frameworks (CSD is a gateway-level arbiter).
    * Resolving conflicts between two agents of equal trust levels (handled by the Mission-Root Conflict Resolver).
    * Providing a natural language explanation for registration failures to the subagent.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized "Documentation Agent" from shadowing the `run_shell_command` tool provided by the mission root.
* **The Happy Path (Tasks):**
    1. The Orchestrator initializes the MCP Any gateway with a mission-root manifest that includes `run_shell_command`.
    2. The CSD automatically locks the `run_shell_command` namespace to the mission-root identity.
    3. A specialized subagent is spawned to "Analyze Documentation."
    4. The subagent attempts to register a local tool named `run_shell_command` via the PNTD bridge.
    5. The CSD intercepts the registration request and performs a lineage check.
    6. The CSD identifies that the subagent lacks the authority to shadow a mission-root capability.
    7. The CSD rejects the registration and logs a "Shadowing Attempt" security event.

## 4. Design & Architecture
* **System Flow:**
    ```
    [Subagent] -> (Register: "tool_name") -> [PNTD Provider]
                                                 |
                                                 v
    [CSD Arbiter] <-> [Lineage Validator] <-> [Registry Lock Table]
           |
           +--> (Valid) -> [Discovery Bus]
           |
           +--> (Conflict) -> [Security Alert Hub]
    ```
* **APIs / Interfaces:**
    * `POST /v1/discovery/register`: Updated to include `agent_lineage_token`.
    * `GET /v1/discovery/locks`: Returns current namespace reservations.
* **Data Storage/State:**
    * Registry locks are stored in the Shared KV Store (Blackboard) using a specialized `NAMESPACE_LOCK` prefix.
    * Lineage tokens are validated against the hardware-bound `SRM Provider`.

## 5. Alternatives Considered
* **First-Come-First-Served (FCFS):** Rejected because it allows malicious subagents to "race" and register high-trust names before legitimate providers during cold-starts.
* **UUID-Based Namespacing:** Rejected as it breaks agent "Self-Discovery" and tool-calling simplicity; LLMs reason better with semantic names like `read_file` than `fs_read_uuid_123`.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The CSD must rely on hardware-attested identity fragments. A subagent cannot "spoof" its parentage to bypass the lineage validator.
* **Observability:** Shadowing attempts are visualized in the `Local Security Violation Monitor` and trigger real-time alerts in the `Swarm Topology Widget`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
