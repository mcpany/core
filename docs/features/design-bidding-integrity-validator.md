# Design Doc: Bidding Integrity Validator
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
With the transition to UACO-driven swarms, "Shadow Bidding" has emerged as a significant threat. Malicious agents can submit bids for tasks they are not qualified for, or misrepresent their security posture to capture sensitive data flows. Because current negotiation hubs focus on "Task Allocation" rather than "Capability Verification," these bids can lead to mission failure or data exfiltration.

The Bidding Integrity Validator ensures that every agent participating in a task auction possesses the skills and security credentials it claims. By cross-referencing bids against hardware-attested Capability Cards and historical performance profiles, it protects the integrity of the swarm's coordination mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a validation layer for the UACO Negotiation Hub.
    * Verify agent bids against TPM-signed "Capability Cards."
    * Detect and block "Shadow Bids" that misrepresent skills or security posture.
    * Provide a behavioral profiling score for bidding agents based on historical mission outcomes.
* **Non-Goals:**
    * Managing the economic aspects of task auctions (handled by UACO).
    * Enforcing inter-agent transport security (handled by the Named-Pipe/WebSocket layer).
    * Validating tool metadata (handled by SMS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Ensure that a "Database Admin" task is only delegated to an agent with a hardware-attested DB-access skill.
* **The Happy Path (Tasks):**
    1. Parent Agent publishes a task card requiring `fs:read` and `sql:execute` capabilities.
    2. Multiple subagents submit bids to the UACO Negotiation Hub.
    3. Bidding Integrity Validator intercepts the bids.
    4. Validator retrieves the "Capability Card" for each bidding agent.
    5. Validator verifies the TPM signature of the Capability Cards.
    6. Validator matches the required skills against the attested skills.
    7. One agent claims to have `sql:execute` but its Capability Card only lists `http:get`.
    8. Validator flags the bid as a "Shadow Bid" and excludes it from the auction.
    9. The task is safely delegated to a verified agent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Bid Submission] --> B[UACO Negotiation Hub]
        B --> C[Bidding Integrity Validator]
        C --> D[Capability Card Retriever]
        D --> E[Hardware Attestation Verifier]
        E --> F[Skill Matcher]
        F --> G{Bid Valid?}
        G -- Yes --> H[Surface Bid to Auction]
        G -- No --> I[Quarantine Bid & Log Anomaly]
        J[Behavioral Profile DB] --> F
    ```
* **APIs / Interfaces:**
    * `validator.ValidateBid(bid, agentID) -> ValidationResult`: Evaluates a bid against an agent's attested capabilities.
    * `validator.UpdateBehavioralProfile(agentID, outcome) -> bool`: Updates an agent's reputation score.
* **Data Storage/State:**
    * **Capability Card Registry:** A TPM-protected store of signed agent skills.
    * **Behavioral Profile Store:** A historical record of agent task completions and security incidents.

## 5. Alternatives Considered
* **Self-Attested Bidding:** Rejected due to the high risk of identity and capability spoofing.
* **Centralized Skill Verification:** Rejected to maintain the decentralized nature of the Universal Agent Bus.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The validator must utilize SMI/FSI for non-repudiable agent identification.
* **Observability:** Integrated with the "UACO Bid Safety Analyzer" for real-time visualization of bid anomalies.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation. Countering Shadow Bidding in UACO-driven swarms.
