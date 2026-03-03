# Design Doc: Skill Attestation Registry (SAR)
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
The "ClawHavoc" incident revealed that users often install "Skills" (MCP tools) based on name alone, making them vulnerable to "Skill-Squatting" and malicious code injection. MCP Any needs a mechanism to verify that a tool being discovered or executed matches a known-safe behavioral contract and has not been tampered with. The Skill Attestation Registry (SAR) provides a centralized (or federated) source of truth for tool integrity.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographic verification mechanism for MCP tool definitions.
    * Enable "Community Reputation" scores to be attached to tool signatures.
    * Support "Behavioral Contracts" (e.g., "This tool only ever calls `https://api.github.com/*`").
    * Integrate with the Discovery Middleware to filter out unverified tools.
* **Non-Goals:**
    * Hosting the actual tool code (SAR only hosts metadata and signatures).
    * Providing a runtime for the tools (handled by the Sidecar Execution Engine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise AI Architect
* **Primary Goal:** Ensure that no unverified or "Shadow" tools are available to their agent swarm.
* **The Happy Path (Tasks):**
    1. The user enables "Strict Attestation Mode" in MCP Any.
    2. An agent requests a "GitHub helper" tool.
    3. MCP Any discovers a local tool named `github-helper-pro`.
    4. MCP Any queries the SAR Client with the tool's definition hash.
    5. SAR returns a "Verified" status along with a behavioral contract signed by a trusted authority (e.g., "Official MCP Any Maintainers").
    6. MCP Any matches the tool's runtime behavior against the contract and allows execution.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>Gateway: Request Tool (Discovery)
        Gateway->>DiscoveryMiddleware: Scan Local/Remote Sources
        DiscoveryMiddleware-->>Gateway: Tool List (Raw)
        Gateway->>SARClient: Verify(ToolDefinition[])
        SARClient->>SARRemote: GetAttestation(Hashes)
        SARRemote-->>SARClient: AttestationBundle[]
        SARClient-->>Gateway: Verified Tool List + Contracts
        Gateway-->>Agent: Filtered Tool List
    ```
* **APIs / Interfaces:**
    * `GetAttestation(tool_hash string) (Attestation, error)`
    * `VerifySignature(bundle AttestationBundle) bool`
* **Data Storage/State:**
    * Local cache (SQLite) for recently verified tool signatures and contracts.

## 5. Alternatives Considered
* **Manual Allow-listing**: Rejected due to high operational overhead for users as the tool ecosystem scales.
* **In-Gateway Code Analysis**: Rejected because static analysis of arbitrary tool code (often opaque binaries or scripts) is unreliable and computationally expensive.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The SAR client itself must verify the signature of the registry to prevent man-in-the-middle attacks providing false "Verified" statuses.
* **Observability:** Audit logs will record every "Attestation Check Failure" including the tool hash and source.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
