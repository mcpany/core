# Design Doc: AI System Inventory Auditor (ASIA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the enforcement of the EU AI Act beginning in August 2026, organizations are required to maintain a comprehensive inventory of all AI assets—including models, plugins, and MCP servers—to enable risk classification and regulatory reporting. Currently, enterprises struggle with "Shadow AI" (unmanaged agent deployments), which increases security risks and potential non-compliance penalties.

ASIA is required to provide a hardware-attested, real-time registry of all agentic assets within the MCP Any mesh, facilitating automated compliance auditing and risk management.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a hardware-attested, real-time inventory of all models, plugins, and MCP servers.
    * Enable automated risk classification based on regulatory metadata (e.g., EU AI Act, SEC requirements).
    * Generate cryptographically signed compliance reports for external auditors.
    * Detect and flag "Shadow AI" deployments that lack proper provenance.
* **Non-Goals:**
    * Replacing existing enterprise asset management systems (it should integrate with them).
    * Directly enforcing security policies (this is handled by the Policy Firewall and Risk-Classifier Middleware).
    * Managing dataset lifecycle (focus is on active system components).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Compliance Officer
* **Primary Goal:** Generate an inventory report for a regulatory audit to prove all active AI plugins are risk-classified.
* **The Happy Path (Tasks):**
    1. The Officer accesses the "Compliance Dashboard" in the MCP Any UI.
    2. ASIA continuously monitors the mesh and aggregates hardware-attested metadata from all connected nodes.
    3. The Officer filters the inventory by "Risk Class: High" (as defined by EU AI Act rules).
    4. ASIA displays a list of active specialists, their model lineage, and verified plugin provenance.
    5. The Officer triggers "Generate Signed Report."
    6. ASIA produces a TPM-signed PDF/JSON document containing the full system state.
    7. The report is submitted to regulators as deterministic proof of compliance.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent/Plugin Node] -->|Hardware-Attested Metadata| B[ASIA Aggregator]
        B --> C[Inventory Registry]
        D[Regulatory Rules Engine] -->|Risk Mapping| C
        C --> E[Compliance Dashboard]
        C --> F[Signed Report Generator]
    ```
* **APIs / Interfaces:**
    * `asia.GetInventory(filters) -> InventoryList`: Retrieves current system state.
    * `asia.ClassifyRisk(componentID) -> RiskLevel`: Returns regulatory risk mapping.
    * `asia.GenerateAttestedReport() -> SignedDocument`: Produces audit-ready artifacts.
* **Data Storage/State:**
    * **Compliance Ledger:** A cryptographically chained log of component additions and upgrades.
    * **Regulatory Metadata Store:** Cached classification rules for EU AI Act and SEC standards.

## 5. Alternatives Considered
* **Manual Spreadsheets:** Rejected as they are non-deterministic, prone to error, and cannot keep pace with dynamic agent swarms.
* **Network Scanners:** Rejected because they cannot identify model lineage or "Plugin-as-Update" provenance without deep agentic awareness.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Inventory data is restricted to authorized "Auditor" roles. Reports are hardware-signed to prevent tampering.
* **Observability:** Integrated with the "Service Mesh Topology Monitor" to provide a compliance-aware view of the mesh.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
