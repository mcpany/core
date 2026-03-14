# Design Doc: Full-Stack Metadata Lineage Guard
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
With the rise of CVE-2026-42001 (Metadata Poisoning), it is no longer sufficient to just sanitize metadata. Mature agent swarms operating in decentralized environments (like DCA) require proof that the tool schemas they are using haven't been tampered with.

The Full-Stack Metadata Lineage Guard extends the Verified Metadata Lineage (VML) protocol to provide end-to-end cryptographic provenance for every structural "hint" given to the LLM—from top-level tool descriptions down to individual property-level examples and parameter titles.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographic chain of custody for all tool metadata.
    * Enable property-level attestation (individual parts of a schema can be signed by different entities).
    * Support "Lineage-Aware Discovery" where agents can filter tools based on their metadata provenance.
* **Non-Goals:**
    * Sanitizing the content (handled by the Tool Metadata Sanitizer).
    * Signing the tool execution results (handled by Signed Context Chain).

## 3. Critical User Journey (CUJ)
* **User Persona:** Decentralized Swarm Orchestrator
* **Primary Goal:** Verify that a tool's documentation and examples haven't been modified to "trick" the agent into incorrect usage.
* **The Happy Path (Tasks):**
    1. Agent receives a list of tools from a DIQ (Decentralized Intent Quorum).
    2. MCP Any extracts the VML provenance tokens embedded in the JSON-RPC response.
    3. The Lineage Guard verifies the signatures against the developer's public key and the organization's policy.
    4. The agent is presented with a "Trust Score" for the tool's documentation.
    5. High-risk tools with broken lineage are automatically quarantined.

## 4. Design & Architecture
* **System Flow:**
    * **Provenance Parser**: Extracts Merkle-tree based signatures from schema extensions (e.g., `x-vml-signature`).
    * **Trust Resolver**: Maps signatures to known identities and enterprise policy tiers.
    * **Attestation Registry**: Local cache of verified public keys and lineage history.
* **APIs / Interfaces:**
    * Extension of the MCP `list_tools` and `get_tool` methods to include provenance metadata.
* **Data Storage/State:**
    * SQLite-backed provenance store for historical lineage tracking.

## 5. Alternatives Considered
* **Whole-File Signing**: Rejected because it prevents legitimate middleware from augmenting schemas (e.g., adding local auth examples).
* **Centralized Registry**: Rejected to maintain compatibility with decentralized agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "No Signature, No Trust" policy for metadata in enterprise profiles.
* **Observability:** Visualizes the "Provenance Chain" in the UI Dashboard.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation.
