# Design Doc: Consensus-Based Skill Attestation (CBSA) Hub
**Status:** Draft
**Created:** 2026-04-14

## 1. Context and Scope
The "ClawHub" supply chain crisis, where one in five skills is malicious, has proven that individual agent security and centralized marketplaces are insufficient. Malicious actors use typosquatting and "delayed payload" tactics to bypass static analysis. The CBSA Hub moves the security boundary from the marketplace to the mesh, requiring multi-agent consensus and behavioral profiling before a tool is trusted by the swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a multi-agent quorum for tool discovery and "grafting."
    * Mandate behavioral profiling in a "Ghost Shell" for all newly discovered marketplace tools.
    * Maintain a mesh-resident reputation ledger for all connected skills.
    * Enable sub-millisecond revocation of tool capabilities upon consensus drift or auditor alert.
* **Non-Goals:**
    * Replacing existing marketplaces like ClawHub (MCP Any acts as a validating proxy for them).
    * Blocking all unverified tools (users can still override, but with explicit attestation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Orchestrator
* **Primary Goal:** Safely discover and deploy a specialized "Data Analysis" skill from a public registry.
* **The Happy Path (Tasks):**
    1. The agent discovers a new skill on a public registry (e.g., ClawHub).
    2. MCP Any intercepts the discovery and places the skill in "Quarantine."
    3. The CBSA Hub triggers a "Ghost Shell" profiling session where the skill's behavior is monitored.
    4. Two independent "Auditor Agents" review the profiling logs and the tool schema.
    5. Upon a 2/2 positive consensus, the skill is promoted to the "Trusted" bus.
    6. The primary agent can now invoke the tool within the verified mission scope.

## 4. Design & Architecture
* **System Flow:**
    * Discovery Interceptor -> Behavioral Profiler (Ghost Shell) -> Consensus Engine (Multi-Agent Quorum) -> Trusted Bus.
* **APIs / Interfaces:**
    * `ProposeTool(metadata, origin)`: Initiates the attestation lifecycle.
    * `SubmitAudit(toolID, signature, score)`: Allows auditors to submit their findings.
    * `GetToolReputation(toolID)`: Returns the current consensus score for a tool.
* **Data Storage/State:** Mesh-resident reputation ledger (sharded via the Blackboard).

## 5. Alternatives Considered
* **Manual Human Review**: Rejected as it cannot scale with machine-speed autonomous swarms.
* **Pure Static Analysis**: Rejected as it fails to detect "Delayed Payload" and prompt-injection vulnerabilities in complex skills.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Auditors must be hardware-attested and run in isolated environments.
* **Observability**: All profiling logs and consensus votes are recorded in the mesh-resident audit trail.

## 7. Evolutionary Changelog
* **2026-04-14:** Initial Document Creation.
