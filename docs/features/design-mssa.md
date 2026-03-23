# Design Doc: Multi-Signature Skill Attestation (MSSA)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
The "ClawHub" compromise revealed a critical weakness in the AI agent supply chain: a single point of failure in tool trust. Currently, once an agent framework or a user "trusts" a skill provider, subsequent updates or dynamic tool grafts are often ingested without further verification. This allows for "Rug-Pull" attacks where a malicious update is pushed to a previously clean skill.

MSSA addresses this by mandating that any dynamic skill grafting or high-risk tool installation be attested by multiple independent parties. Specifically, it requires cryptographically bound approval tokens from both the agent framework and a verified third-party security auditor before the capability is exposed to the swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a multi-signature requirement for dynamic tool grafting.
    * Decouple tool discovery from tool execution trust.
    * Provide a standardized interface for third-party security auditors to sign skill manifests.
    * Neutralize "Rug-Pull" supply chain attacks at the infrastructure layer.
* **Non-Goals:**
    * Auditing the code itself (MSSA is the enforcement mechanism for auditor results).
    * Restricting static, pre-verified internal tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure that no un-audited community skills are executed within the production swarm.
* **The Happy Path (Tasks):**
    1. A subagent discovers a new "Data Analytics" skill on the marketplace.
    2. The subagent proposes grafting the skill to complete a task.
    3. MCP Any intercepts the graft request and identifies it as "High Risk."
    4. MCP Any queries the MSSA Hub for attestations.
    5. The MSSA Hub finds a Framework signature but no Auditor signature.
    6. The graft is blocked, and a request is sent to the configured Auditor service.
    7. Once the Auditor signs the skill manifest, the MSSA Hub allows the graft to proceed.

## 4. Design & Architecture
* **System Flow:**
    [Subagent] --(Graft Request)--> [MSSA Middleware] --(Lookup)--> [Attestation Registry]
    [Attestation Registry] --(Query)--> [Auditor Portal]
    [Auditor Portal] --(Signature)--> [Attestation Registry]
    [MSSA Middleware] --(Authorize)--> [Registry]
* **APIs / Interfaces:**
    * `POST /v1/mssa/attest`: Submit a multi-signature manifest for a skill.
    * `GET /v1/mssa/verify?skill_hash=[sha256]`: Check if a skill meets the quorum requirements.
* **Data Storage/State:**
    Multi-signature manifests are stored in the hardware-bound Namespace-Locked Registry.

## 5. Alternatives Considered
* **Single-Signer (Framework Only):** Rejected because it doesn't protect against framework-level compromises or "Shadow Updates" from marketplace providers.
* **Human-in-the-Loop (HITL) for every graft:** Rejected as it doesn't scale to autonomous swarm speeds.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The MSSA middleware itself is origin-locked and runs within the secure enclave.
* **Observability:** Audit logs track every graft attempt, signature provenance, and rejection reason.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation.
