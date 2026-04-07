# Design Doc: Federated Skill Reputation (FSR) Provider
**Status:** Draft | In Review | Approved
**Created:** 2026-04-07

## 1. Context and Scope
The "ClawHavoc" malicious registry crisis has destroyed trust in community-driven tool registries. Static analysis and sandbox isolation are no longer sufficient to maintain security in a decentralized agent mesh. The FSR Provider implements a federated reputation system, where tool safety scores are determined by the collective, hardware-attested consensus of independent security nodes across the Universal Agent Bus (UAB).

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a decentralized security middleware for sharing and validating tool safety scores.
    * Mandate multi-signature attestation from independent "Monitor" and "Auditor" agents before promoting skills to a "Trusted" tier.
    * Support all connected frameworks (OpenClaw, AutoGen, Claude Code) with a unified, cross-mesh reputation signal.
* **Non-Goals:**
    * Replacing existing marketplace-specific reputation models (it acts as a cross-framework layer).
    * Providing automated "remediation" of malicious tools (it only identifies and blocks them).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Leverage community-driven skills while maintaining a high security posture via federated reputation.
* **The Happy Path (Tasks):**
    1. Orchestrator discovers a tool (via the Discovery Sandbox).
    2. FSR Provider queries the decentralized reputation mesh for the tool's hash.
    3. Monitor nodes provide hardware-attested security scores based on behavioral profiling.
    4. If the consensus score exceeds the "Trust Quorum" threshold, the tool is exposed to the swarm.
    5. FSR Provider continuously monitors the tool's real-time reputation and revokes access if the score drops below the threshold.

## 4. Design & Architecture
* **System Flow:**
    `FSR -> [Reputation Mesh / P2P] -> Consensus Voting -> [Audit Log / Attestation] -> Tool Visibility`
* **APIs / Interfaces:**
    * `FSR.GetReputation(tool_hash)`: Queries the mesh for a tool's current safety score and attestation tokens.
    * `FSR.SubmitAttestation(tool_hash, score)`: Broadcasts a local attestation to the reputation mesh.
* **Data Storage/State:** Reputation scores and attestation signatures are stored in a distributed ledger or a Gossip-based P2P state bus.

## 5. Alternatives Considered
* **Centralized Skill Registry:** Rejected because it creates a single point of failure and doesn't align with the decentralized mission of MCP Any.
* **Manual Tool Whitelisting:** Rejected as it is not scalable for large-scale autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All reputation signals must be cryptographically signed by a verified mission-root identity and a hardware-bound attestation token.
* **Observability:** Transparency in all reputation voting, with detailed audit trails for each tool's consensus history.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
