# Design Doc: Swarm Trust Domains (STDs)

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As AI agents evolve from single-purpose bots to complex, self-organizing swarms, the traditional "Flat" security model (where an agent either has access to a tool or doesn't) is insufficient. Agents within a swarm need to share context, state, and tools fluidly among themselves but must remain strictly isolated from other swarms or external entities. Swarm Trust Domains (STDs) provide a cryptographic and policy-driven boundary to manage this "Clustered Trust."

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable automatic context and tool inheritance within a defined "Trust Domain."
    *   Implement cryptographic attestation for agents joining a Trust Domain.
    *   Allow for "Domain-Scoped" policies (e.g., "All agents in Domain-A can read /tmp/swarm-a/*").
    *   Support hierarchical domains (e.g., `org.finance.audit`).
*   **Non-Goals:**
    *   Replacing individual agent authentication.
    *   Managing the LLM's internal weights or biases.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Architect.
*   **Primary Goal:** Deploy a swarm of 5 specialized agents for a "Financial Audit" task where they share sensitive data but cannot access the general-purpose "Web Search" tools or be accessed by the "Marketing" swarm.
*   **The Happy Path (Tasks):**
    1.  Architect defines a Trust Domain: `audit-2026-001`.
    2.  Architect assigns a "Domain Key" to the 5 audit agents.
    3.  When an agent calls a tool, MCP Any verifies its `Domain-ID` and `Attestation-Token`.
    4.  The Policy Firewall applies the `audit-2026-001` wildcard rules (e.g., `audit_db:*`).
    5.  Agents within the domain share a common "Blackboard" (Shared KV Store) scoped only to their Domain-ID.

## 4. Design & Architecture
*   **System Flow:**
    - **Registration**: Agents present a signed JWT containing their `Domain-ID`.
    - **Validation**: MCP Any validates the signature against the registered Domain Public Key.
    - **Context Isolation**: The `Recursive Context Protocol` is extended to include `Swarm-Domain-ID`, ensuring context doesn't leak across boundaries.
*   **APIs / Interfaces:**
    - `POST /v1/domains`: Create a new trust domain.
    - `GET /v1/domains/:id/tools`: List tools available to a specific domain.
*   **Data Storage/State:** Domain definitions and active swarm session states are stored in the encrypted SQLite backend.

## 5. Alternatives Considered
*   **Manual ACLs for Every Agent**: Too complex to manage as swarms scale to dozens of agents.
*   **Network-Level Isolation (VPC)**: Lacks the granular "Intent-Aware" control needed for AI tool calls.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Even within a domain, "Least Privilege" is enforced via subagent scoping. Domain keys must be rotated regularly.
*   **Observability:** The Management Dashboard will visualize agents grouped by their Trust Domains.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
