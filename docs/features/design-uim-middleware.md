# Design Doc: Universal Identity Mapping (UIM) Middleware
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
In the current agentic ecosystem, swarms are increasingly heterogeneous, consisting of agents from multiple frameworks (Claude Code, Gemini CLI, OpenClaw, AutoGen) collaborating on a single mission. Each framework has its own internal identity and trust model. When a Claude agent delegates a task to an OpenClaw specialist, there is often a "Trust Gap"—either the specialist inherits too much privilege (Privilege Shadowing) or is blocked due to lack of recognizable credentials.

The **Universal Identity Mapping (UIM) Middleware** for MCP Any acts as a sovereign identity bridge. it normalizes framework-specific identity tokens into a universal, mesh-resident identity, ensuring consistent permission enforcement and preventing unauthorized privilege escalation during cross-framework handoffs.

## 2. Goals & Non-Goals
* **Goals:**
    * Reconcile framework-specific identities (e.g., Anthropic's Session Tokens, Gemini's A2A IDs) into a unified MCP Any Mesh Identity.
    * Enforce "Trust Parity" where a subagent's permissions are limited by both the parent's trust level and the framework's capability constraints.
    * Provide a cryptographically signed "Lineage Proof" that persists across framework boundaries.
    * Neutralize "Privilege Shadowing" by explicitly mapping inherited permissions during every delegation event.
* **Non-Goals:**
    * Replacing framework-level authentication (UIM maps them, it doesn't replace their internal auth).
    * Managing user-level IAM (UIM is for Non-Human Identities/NHIs).

## 3. Critical User Journey (CUJ)
* **User Persona**: Local LLM Swarm Orchestrator
* **Primary Goal**: Securely delegate a filesystem-write task from a high-trust Claude agent to a specialist OpenClaw subagent without exposing host-level credentials.
* **The Happy Path (Tasks):**
    1. A Claude Agent initiates a delegation to an OpenClaw Specialist via the MCP Any A2A Hub.
    2. The delegation request includes the Claude framework token.
    3. The **UIM Middleware** intercepts the request and extracts the Claude identity.
    4. UIM maps the Claude identity to a "Mesh Identity" and verifies the active mission-root.
    5. UIM generates a "Shadow Identity" for the OpenClaw specialist, which contains only a subset of the Claude agent's permissions (Least Privilege).
    6. The OpenClaw Specialist receives the task with a Mesh-bound token.
    7. When the Specialist calls a tool (e.g., `write_file`), the Gateway validates the Mesh token against the UIM mapping to ensure it hasn't shadowed unauthorized privileges.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        subgraph Framework A (Claude)
            AgentA[Claude Agent]
        end
        subgraph MCP Any Mesh
            A2A[A2A Hub]
            UIM[UIM Middleware]
            Hub[Identity Hub]
        end
        subgraph Framework B (OpenClaw)
            AgentB[OpenClaw Specialist]
        end

        AgentA -- "Delegates (Claude Token)" --> A2A
        A2A -- "Intercept" --> UIM
        UIM -- "Verify & Map" --> Hub
        Hub -- "Mesh-Resident Token" --> UIM
        UIM -- "Task + Mesh Token" --> AgentB
        AgentB -- "Tool Call (Mesh Token)" --> Gateway
    ```
* **APIs / Interfaces:**
    * `POST /v1/identity/map`: Internal endpoint for the A2A Hub to request framework-to-mesh mapping.
    * `GET /v1/identity/lineage/{mesh_id}`: Returns the hardware-attested lineage of a specific mesh identity.
    * `Header x-mcpany-mesh-identity`: Mandatory header for all cross-framework tool calls.
* **Data Storage/State:**
    * **Mapping Table**: In-memory (Redis) mapping of `framework_token_hash` -> `mesh_identity_id` -> `permission_set`.
    * **Lineage Store**: Persistent audit log of identity transitions (who delegated to whom and when).

## 5. Alternatives Considered
* **Framework-Specific Bridges**: Build individual adapters for every framework pair (Claude-to-OpenClaw, Gemini-to-Claude).
    * *Rejected*: O(N^2) complexity. Unscalable as new frameworks emerge.
* **JWT-only Pass-through**: Trust frameworks to pass their own tokens.
    * *Rejected*: Different frameworks use different claims and formats, making "Universal" policy enforcement impossible.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * **TPM-Bound Anchoring**: Every Mesh Identity must be anchored to a hardware-root (HAMD compliance).
    * **Token Rotation**: Mesh tokens have a maximum TTL of 15 minutes or mission-end, whichever comes first.
* **Observability:**
    * **Identity Heatmap**: Visualize which framework identities are currently active and their respective trust levels.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
