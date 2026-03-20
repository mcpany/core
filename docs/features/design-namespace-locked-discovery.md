# Design Doc: Namespace-Locked Discovery (NLD) Gateway
**Status:** Draft
**Created:** 2026-05-19

## 1. Context and Scope
As agent swarms grow in complexity and heterogeneity, the tool discovery phase has become a primary attack surface. **"Namespace Collision"** is being weaponized via **Discovery Hijacking**, where malicious subagents or unverified registries inject tools with names identical to high-trust capabilities (e.g., `git_push`). This shadowing allows them to intercept parent agent calls. MCP Any needs to implement a "Namespace-Locked Discovery" mechanism to ensure deterministic, collision-free capability mapping.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory "Namespace-Lock" for all tool and agent capability registrations.
    * Ensure that a capability registered in a high-trust namespace cannot be "shadowed" or overwritten by a lower-trust registry.
    * Provide a deterministic mapping between a tool's unique identity (e.g., `github.com/org/repo/tool`) and its discoverable name.
* **Non-Goals:**
    * Restricting the total number of tools (focus is on unique and safe naming).
    * Modifying the internal logic of the discovery command itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Admin
* **Primary Goal:** Ensure that a developer's `git_push` tool call always routes to the verified organization-approved script, never to a malicious subagent's shadow.
* **The Happy Path (Tasks):**
    1. The admin configures the NLD Gateway to prioritize the `core` and `org` namespaces.
    2. The organization's verified `git_push` tool is registered in the `org` namespace.
    3. A malicious repository attempting to subvert the swarm registers its own `git_push` in the `local` namespace.
    4. When an agent searches for `git_push`, the NLD Gateway returns the `org:git_push` capability and blocks/flags the `local` collision.
    5. The agent executes the verified tool with confidence.

## 4. Design & Architecture
* **System Flow:**
    `[Discovery Provider] -> [NLD Gateway] -> [Agent Discovery Bus]`
* **APIs / Interfaces:**
    * `NLD.register_capability(name, namespace, metadata)`: Validates and locks a name to a specific namespace and origin.
    * `NLD.resolve_collision(name)`: Applies precedence policies to return the highest-trust capability.
* **Data Storage/State:**
    * Namespace mappings and registry priorities are stored in a persistent, hardware-attested registry index.

## 5. Alternatives Considered
* **UUID-Only Discovery**: Rejected as it makes the discovery bus non-human-readable and breaks existing agent framework naming conventions.
* **Manual Name Pre-registration**: Rejected as it adds excessive friction to the dynamic agentic lifecycle.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: NLD prevents "Shadow Capability" mapping, a critical step in neutralizing machine-speed coercion.
* **Observability**: Collision events and "Shadowing Attempts" are surfaced in the UI Federated Discovery Monitor.

## 7. Evolutionary Changelog
* **2026-05-19:** Initial Document Creation.
