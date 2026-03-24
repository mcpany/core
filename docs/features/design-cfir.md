# Design Doc: Cross-Framework Identity Relay (CFIR)
**Status:** Draft
**Created:** 2026-06-16

## 1. Context and Scope
As AI agent swarms evolve from single-framework silos into heterogeneous meshes (e.g., Claude Code orchestrating OpenClaw specialists), a critical bottleneck has emerged: **Identity Fragmentation (ID-Frag)**. Current hardware-attested identities (TPM/Secure Enclave) are often bound to the specific framework that initiated the mission. When a task is delegated across framework boundaries, the "attestation strength" often decays or is lost entirely, forcing the recipient agent to either operate with reduced trust or trigger redundant, latency-heavy handshakes.

MCP Any needs to solve this by acting as the authoritative **Identity Relay**. CFIR will provide the "Trust Glue" that allows a mission-root identity to traverse different frameworks while maintaining a cryptographic chain of custody and full hardware-attested strength.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested "Identity Fragments" that are framework-agnostic.
    * Enable sub-millisecond identity re-attestation during cross-framework handoffs.
    * Maintain a verifiable "Mission Root" lineage across Claude Code, OpenClaw, and AutoGen.
    * Neutralize "ID-Frag" by implementing a unified trust-root authority.
* **Non-Goals:**
    * Replacing existing framework-specific identity providers (e.g., Gemini CLI tokens).
    * Managing persistent user identities beyond the scope of a specific mission.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Ensure a high-trust database-migration mission maintains its "Admin" attestation as it moves from a Claude Code supervisor to an OpenClaw execution specialist.
* **The Happy Path (Tasks):**
    1. Claude Code initiates the mission and requests a `MissionRoot` token from MCP Any CFIR.
    2. MCP Any generates a TPM-signed `IdentityFragment` bound to the mission scope.
    3. Claude Code delegates the "migration" task to an OpenClaw specialist via the UACO bridge.
    4. The OpenClaw specialist receives the `IdentityFragment` and submits it to MCP Any for "Relay Verification."
    5. MCP Any verifies the fragment's hardware signature and lineage, issuing a framework-local `RelayToken` to OpenClaw.
    6. OpenClaw executes the migration tool using the `RelayToken`, which is accepted by the database MCP server as having "Mission-Root" authority.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Framework_A as Claude Code
        participant CFIR as MCP Any (CFIR)
        participant Bridge as UACO/A2A Bridge
        participant Framework_B as OpenClaw Specialist

        Framework_A->>CFIR: Request Mission-Root Identity
        CFIR-->>Framework_A: Issue Hardware-Attested Fragment
        Framework_A->>Bridge: Delegate Task + Fragment
        Bridge->>Framework_B: Deliver Task + Fragment
        Framework_B->>CFIR: Relay Verification (Fragment)
        CFIR->>CFIR: Validate Lineage & TPM Sig
        CFIR-->>Framework_B: Issue Framework-Local RelayToken
        Framework_B->>Tool: Execute with RelayToken
    ```
* **APIs / Interfaces:**
    * `POST /v1/identity/fragment/issue`: Generates a hardware-bound mission fragment.
    * `POST /v1/identity/fragment/relay`: Exchanges a fragment for a framework-specific session token.
    * `x-mcp-identity-fragment`: New standard header for inter-agent coordination.
* **Data Storage/State:**
    * Ephemeral storage of active `MissionRoot` lineages in a hardware-isolated memory region.
    * No persistent logging of PII; only cryptographic hashes of fragments for revocation checks.

## 5. Alternatives Considered
* **Framework-Specific Bridges:** Building N-to-N bridges between all frameworks. *Rejected* due to O(N^2) complexity and maintenance overhead.
* **Shared JWT Secrets:** Using common symmetric keys across frameworks. *Rejected* as it violates Zero-Trust principles and doesn't provide hardware attestation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Fragments are time-bound and cryptographically linked to a specific mission UUID. Compromise of a single fragment does not grant access to other missions.
* **Observability:** Centralized audit logs in MCP Any will show the "Lineage Path" of identities as they move between frameworks, without exposing the internal reasoning of the agents.

## 7. Evolutionary Changelog
* **2026-06-16:** Initial Document Creation.
