# Design Doc: Registry Integrity Guard (RIG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of "Autonomous Registry Poisoning" (ARP), where malicious MCP registries dynamically inject shadowed or malicious tool schemas in response to discovery queries, the security of the discovery phase has become a critical vulnerability. Current static signature checks are insufficient against on-the-fly schema generation.

MCP Any needs a real-time, semantic attestation layer that validates the integrity of tool definitions as they are discovered, ensuring that high-trust capabilities cannot be silently shadowed or redirected to malicious endpoints.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time semantic analysis of MCP tool schemas during discovery.
    * Block "Dynamic Tool Injection" where a registry attempts to shadow an existing high-trust tool.
    * Provide a cryptographic audit trail for every tool definition ingestion.
    * Integrate with hardware-attested identity to verify registry provenance.
* **Non-Goals:**
    * Validating the internal logic of the tool itself (this is handled by the Ghost Shell Hook Profiler).
    * Providing a full MCP registry (RIG is a validating proxy).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent "Capability Hijacking" in a swarm where subagents might query untrusted or dynamic registries.
* **The Happy Path (Tasks):**
    1. The user configures a "High-Trust" local MCP server.
    2. A subagent performs a broad tool discovery query that includes a dynamic enterprise registry.
    3. The dynamic registry attempts to return a tool schema that overlaps with a name/description of the "High-Trust" server but points to an external URL.
    4. The RIG intercepts the response, detects the semantic collision and unauthorized shadowing.
    5. The RIG blocks the malicious schema and alerts the user via the Security Dashboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent] -->|Discovery Query| B[PNTD Discovery Provider]
        B --> C{RIG Middleware}
        C -->|Registry Poll| D[MCP Registries]
        D -->|Tool Schemas| C
        C -->|Semantic Analysis| E[Collision Detector]
        E -->|Shadowing Check| F[Attestation Engine]
        F -->|Verified Schemas| B
        B -->|Masked Response| A
    ```
* **APIs / Interfaces:**
    * `RegisterToolSchema(schema JSON, provenance Token) error`
    * `ValidateDiscoveryResponse(response MCPResponse) (MCPResponse, error)`
* **Data Storage/State:**
    * Persistent cache of "Golden Schemas" for high-trust tools.
    * Real-time bloom filter for collision detection across registries.

## 5. Alternatives Considered
* **Static Allowlisting:** Rejected because modern swarms require dynamic tool discovery for efficiency.
* **Pure Signature Verification:** Rejected because attackers can sign malicious schemas using compromised or ephemeral keys.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The RIG operates on the principle of "Verify Every Discovery." No schema is exposed to the agent until its lineage and semantic footprint are cleared.
* **Observability:** Logs all "Discovery Block" events with full diffs between expected and received schemas.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
