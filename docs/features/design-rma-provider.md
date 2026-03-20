# Design Doc: Recursive Mission Attestation (RMA) Provider
**Status:** Draft
**Created:** 2026-06-07

## 1. Context and Scope
As AI agent swarms evolve from linear handoffs to complex, recursive sub-mission hierarchies, the risk of "Intent Hijacking" increases. A subagent, several hops away from the user's original intent, may be coerced or hallucinate a "forked" mission that contradicts parent constraints. Transport-layer security (mTLS) and session-bound tokens only verify *who* is talking, not *what* they are authorized to do in the context of the global mission.

The RMA Provider solves this by issuing hardware-attested, recursive "Mission Receipts." Each receipt cryptographically binds a sub-mission to its parent intent, creating a verifiable chain of sovereignty that can be audited at any tool call or delegation hop.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) mission tokens for every subagent spawn.
    * Provide a recursive validation mechanism that checks sub-intent alignment against the Mission Root.
    * Maintain a deterministic audit trail of intent delegation via the Mission-Receipt Logging Service.
    * Support cross-framework attestation (OpenClaw SRM and Gemini HAIL).
* **Non-Goals:**
    * Performing semantic reasoning about the "correctness" of the mission (this is handled by the AIR Broker).
    * Enforcing per-tool capability scoping (handled by the Policy Firewall and EPM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect for an Enterprise Agent Mesh.
* **Primary Goal:** Ensure that a "Junior Research Agent" spawned by a "Lead Developer Agent" cannot be tricked into exfiltrating the codebase, even if the Lead agent is partially compromised.
* **The Happy Path (Tasks):**
    1. The Lead Developer Agent requests a sub-mission to "Search for vulnerabilities in the auth module."
    2. The RMA Provider issues a "Mission Receipt" (MR-1) signed by the hardware root, binding the sub-mission to the parent "Audit Project" intent.
    3. The Junior Research Agent receives MR-1 and presents it to the `file_read` tool.
    4. The `file_read` tool validates MR-1 with the RMA Provider.
    5. The RMA Provider verifies that "Search vulnerabilities" is a valid branch of "Audit Project" and allows the call.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Parent as Parent Agent
        participant RMA as RMA Provider
        participant TPM as Hardware TPM
        participant Sub as Subagent
        participant Tool as MCP Tool

        Parent->>RMA: Request Sub-mission (Intent + Parent-MR)
        RMA->>TPM: Sign Mission Receipt (MR-sub)
        TPM-->>RMA: Signed Token
        RMA-->>Parent: MR-sub
        Parent->>Sub: Spawn (MR-sub)
        Sub->>Tool: Call Tool (Args + MR-sub)
        Tool->>RMA: Validate(MR-sub)
        RMA->>RMA: Verify Chain of Sovereignty
        RMA-->>Tool: Authorized
    ```
* **APIs / Interfaces:**
    * `POST /rma/issue`: Accepts Parent Mission Receipt and new sub-intent; returns a signed MR.
    * `POST /rma/validate`: Accepts a Mission Receipt; returns validity and root intent summary.
* **Data Storage/State:**
    * Mission receipts are stateless but logged to the Mission-Receipt Logging Service for auditability.
    * The RMA Provider maintains a transient "Intent Tree" in memory for high-frequency validation.

## 5. Alternatives Considered
* **Flat JWTs:** Rejected because they don't capture the recursive hierarchy and are susceptible to replay if the session token is leaked.
* **Pure LLM-based Validation:** Rejected due to "Reasoning Drift" and the potential for subagents to hallucinate their own authorization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mission Receipts must be hardware-bound to prevent "Token Smuggling" between different physical hosts or Docker containers.
* **Observability:** All issuance and validation failures are logged with high-entropy trace IDs linked to the Mission-Receipt Logging Service.

## 7. Evolutionary Changelog
* **2026-06-07:** Initial Document Creation.
