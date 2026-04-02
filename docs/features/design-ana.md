# Design Doc: Adaptive Negotiation Arbitrator (ANA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale to 15+ agents, the overhead of task negotiation (bidding, auctions, verification) often exceeds the actual task execution time—a phenomenon known as **Negotiation Drift**. In deep swarms, this leads to token exhaustion and latency spikes, as agents spend more time "talking about work" than "doing work".

The **Adaptive Negotiation Arbitrator (ANA)** optimizes swarm performance by dynamically switching between full auctions and hardware-attested **Capability Matching**. For low-risk, deterministic tasks, ANA bypasses the auction phase and assigns work based on cryptographically signed "Agent Skill Cards", drastically reducing the coordination tax.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce the Auction-to-Execution ratio in large swarms.
    * Implement "Capability-Based Assignment" for low-risk tasks.
    * Provide real-time "Negotiation Efficiency" metrics.
    * Enforce mission-root sovereignty during work assignment.
* **Non-Goals:**
    * Replacing the UACO auction protocol (ANA is a performance layer on top of it).
    * Bypassing security approvals for high-risk tasks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Performance Engineer
* **Primary Goal:** Scale a developer swarm to 50 agents without hitting token limits on task negotiation.
* **The Happy Path (Tasks):**
    1. The Mission-Root submits a task card for "File Linting".
    2. ANA identifies the task as "Low Risk / Deterministic".
    3. ANA scans the mesh for agents with a hardware-attested "Linting" skill card.
    4. ANA assigns the task directly to the most efficient available agent.
    5. The swarm skips the 3-round bidding process.
    6. Performance metrics show a 90% reduction in coordination tokens for that task.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        TC[Task Card] --> ANA[ANA]
        ANA -- Risk Analysis --> RA{High Risk?}
        RA -- Yes --> AUC[Standard UACO Auction]
        RA -- No --> CM[Capability Matching]
        CM -- Skill Match --> AG[Assign Agent]
        AUC -- Bids --> AG
    ```
* **APIs / Interfaces:**
    * `GET /v1/ana/efficiency`: Retrieve swarm-wide negotiation metrics.
    * `POST /v1/ana/policies`: Define risk thresholds for auction bypass.
* **Data Storage/State:**
    * Real-time skill registry in the Autonomous Service Mesh Gateway.
    * Historical task performance data to inform risk analysis.

## 5. Alternatives Considered
* **Static Assignment:** Fixed roles for agents. *Rejected* because it lacks the flexibility to handle dynamic specialist swarms.
* **Global Rate Limiting on Auctions:** *Rejected* because it causes task backlogs during high-intensity mission phases.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Capability matching relies on hardware-attested skill cards. Any forgery attempt triggers immediate mesh-wide quarantine (MSSQ).
* **Observability:** Negotiation Drift alerts integrated with the System Health Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
