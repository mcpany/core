<!--
Copyright 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Design Doc: Project-Level Policy Engine Adapter
**Status:** Draft
**Created:** 2026-06-03

## 1. Context and Scope
As AI agents move from general-purpose assistants to project-specific collaborators (e.g., Claude Code, OpenClaw), the static global security policy becomes a bottleneck. Gemini CLI v0.30.0 has demonstrated the utility of project-resident policies.

MCP Any needs a "Project-Level Policy Engine Adapter" to ingest and enforce security rules defined within the project repository itself. This ensures that tool permissions and data access are context-aware and move with the code, while remaining under the authoritative control of the user's hardware-attested gateway.

## 2. Goals & Non-Goals
* **Goals:**
    * Ingest Gemini-style \`.mcp-policy.json\` or \`mcp-policy.rego\` files from project roots.
    * Enforce repository-specific tool allow/deny lists.
    * Require hardware-attestation for project-local policy overrides.
* **Non-Goals:**
    * Automatically executing arbitrary code found in policy files.
    * Replacing the global security policy (global always takes precedence).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer using OpenClaw Swarms.
* **Primary Goal:** Restrict a specialized 'DB Specialist' agent to only read-only tools when working in a specific production-critical repository.
* **The Happy Path (Tasks):**
    1. User adds a \`mcp-policy.rego\` to the repository root.
    2. MCP Any discovers the policy during the project-local handshake.
    3. User provides a hardware-bound (TPM) signature to attest the new local policy.
    4. MCP Any enforces the restricted toolset for all agents operating within that repository.

## 4. Design & Architecture
* **System Flow:**
    * Discovery: Project-local scan for \`.mcp-policy.*\` files.
    * Validation: Semantic check and hardware attestation.
    * Enforcement: Policy Engine middleware merges global and local rules.
* **APIs / Interfaces:**
    * \`PolicyRegistry\`: Manages the lifecycle of project-bound policy shards.
* **Data Storage/State:**
    * Project-bound policy state is sharded and indexed by repository root hash.

## 5. Alternatives Considered
* **Manual Global Overrides:** Rejected due to operational overhead and "Policy Drift."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Local policies cannot grant more permission than the global root.
* **Observability:** Audit logs will explicitly flag tool calls permitted or denied by "Project-Local Policy."

## 7. Evolutionary Changelog
* **2026-06-03:** Initial Document Creation.
