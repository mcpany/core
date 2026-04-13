# Design Doc: Dynamic Sandbox Expansion (DSE) Controller
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
As agents move from simple chat sessions to complex, multi-phase missions (e.g., Gemini CLI "Chapters"), their resource and isolation requirements fluctuate. A subagent might start with "Read-Only" access to a codebase but later require "Write" access or network connectivity to deploy a fix. Static sandboxes are either too restrictive (blocking progress) or too permissive (violating Least Privilege).

The DSE Controller enables "Elastic Sovereignty" by allowing agent isolation boundaries to expand or contract in response to real-time mission requirements. It leverages OS-native sandboxing primitives (macOS Seatbelt, Windows Sandbox, Linux Namespaces) to modify the agent's environment without requiring a full session restart.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide an API for agents to request "Boundary Expansion" (e.g., adding a mount point or opening a port).
    * Implement "Just-in-Time" (JIT) permission granting via the HITL Middleware.
    * Use OS-native drivers to modify sandbox profiles in-place.
    * Ensure that all expansions are cryptographically bound to a specific mission "Chapter."
* **Non-Goals:**
    * Automatically granting permissions (always requires policy or user attestation).
    * Managing persistent infrastructure (focus is on ephemeral agent sandboxes).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Allow a "Bug Fix" subagent to request temporary write access to a specific directory.
* **The Happy Path (Tasks):**
    1. The agent identifies a fix and determines it needs to write to `/src/lib`.
    2. The agent calls `mcp.request_boundary_expansion({ path: "/src/lib", mode: "rw" })`.
    3. DSE Controller validates the request against the mission-root manifest.
    4. If it exceeds the manifest, the user is prompted via the `MFA Attestation Dialog`.
    5. Upon approval, DSE calls the OS driver to update the sandbox profile for the active process.
    6. The agent performs the write and continues the mission.
    7. When the "Chapter" ends, DSE automatically rolls back the expansion.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `DSE Proxy` -> `Policy Engine / HITL` -> `OS Sandbox Driver` -> `Process Environment`
* **APIs / Interfaces:**
    * `mcp.request_boundary_expansion`: JSON-RPC method for agents.
    * `v1/sandbox/expand`: Internal API for the controller to talk to OS-specific drivers.
* **Data Storage/State:**
    * "Sandbox Manifests" stored in the Blackboard, tracking current vs. baseline boundaries.
    * Transactional logs of all environment modifications.

## 5. Alternatives Considered
* **Docker-Only Sandboxing:** Rejected because it introduces significant overhead for local CLI tools and doesn't support fine-grained OS-native features like macOS Seatbelt.
* **Pre-emptive Over-Privileging:** Rejected as it violates Zero Trust principles and increases the impact of "The Lethal Trifecta."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** DSE ensures that agents only have the minimum necessary access at any given moment. Expansion is always time-bound and mission-scoped.
* **Observability:** Current sandbox boundaries are visualized in the `Config Sandbox Monitor` with real-time expansion alerts.

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
