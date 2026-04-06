# Design Doc: Bid-Inflation Interceptor
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In high-density AI Agent Teams utilizing the Distributed Capability Auction (DCA) protocol, a new instability pattern has emerged: "Bid Inflation." In this scenario, subagents competing for a task recursively re-bid with increasingly higher "Reasoning Effort" (ARE) or "Token Budget" promises, attempting to out-compete siblings. This leads to auction deadlocks and rapid exhaustion of mission-root resource pools without improving task outcomes.

The Bid-Inflation Interceptor is a security extension for the Active Negotiation Broker (ANB) that performs real-time analysis of UACO bids to detect and block these recursive re-bidding loops.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a real-time monitor for the Active Negotiation Broker.
    * Detect "Incremental Escalation" patterns in UACO task bidding.
    * Enforce "Cool-down" periods for subagents attempting to re-bid for the same Task UUID.
    * Cap the maximum effort escalation allowed per task auction.
* **Non-Goals:**
    * Changing the underlying DCA auction logic.
    * Determining the "fair market value" of a task (handled by the mission-root policy).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Resource Manager
* **Primary Goal:** Prevent a task auction from entering an infinite re-bidding loop that drains the mission budget.
* **The Happy Path (Tasks):**
    1. Two subagents bid for a "Code Review" task on the ANB.
    2. Subagent A is outbid by Subagent B.
    3. Subagent A immediately submits a new bid with 20% higher ARE.
    4. The Bid-Inflation Interceptor flags this as a "High-Frequency Escalation."
    5. The Interceptor applies a 500ms cool-down to Subagent A and notifies the ANB.
    6. If Subagent A persists after the cool-down without a significant change in its "Skill Attestation," the interceptor blocks the bid and escalates to the Mission Root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Subagent Bids] --> B[ANB Hub]
        B --> C[Bid-Inflation Interceptor]
        C -->|Analyze History| D{Inflation Detected?}
        D -->|Yes| E[Throttle/Block Bid]
        D -->|No| F[Process Bid]
        E --> G[Mission Root Alert]
    ```
* **APIs / Interfaces:**
    * `interceptor.ValidateBid(bid UACOBid, history []UACOBid) -> Result`: Checks if a bid constitutes inflation.
* **Data Storage/State:**
    * **Auction History Cache:** An ephemeral, Task-UUID indexed cache of recent bids, including timestamps and effort metrics.

## 5. Alternatives Considered
* **Hard Caps on ARE:** Rejected because some tasks legitimately require high effort; a static cap would break complex reasoning.
* **Mission-Root Approval for Every Bid:** Rejected due to the latency it would introduce in high-speed coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The interceptor itself must be hardware-attested to ensure it cannot be disabled by a compromised ANB shard.
* **Observability:** Metrics on "Blocked Bids" and "Inflation Events" are surfaced in the `ui/roadmap.md` "Bid-Inflation Alert Center."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation. Addressing "Negotiation Storms" identified in the daily market sync.
