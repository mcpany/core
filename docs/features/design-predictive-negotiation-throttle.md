# Design Doc: Predictive Negotiation Throttle
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
As agent swarms grow in complexity, the Distributed Capability Auction (DCA) protocol has introduced a new failure mode: "Negotiation Exhaustion." Agents spend more compute cycles and tokens bidding on tasks than executing them, leading to "Negotiation Storms." MCP Any, as the universal gateway, is uniquely positioned to observe these bids and intercede. The Predictive Negotiation Throttle is a gateway-level middleware that evaluates "Reasoning Cost Estimates" (RCE) in UACO bids and halts low-value or recursive negotiation loops before they consume host and swarm resources.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Reduce swarm-wide token consumption by 30% in deep agent chains.
    *   Detect and halt recursive bidding loops (Negotiation Storms) within 3 cycles.
    *   Provide a standardized "Reasoning Cost Estimate" (RCE) schema for UACO bids.
    *   Implement "Value-at-Risk" (VaR) thresholds for task delegation.
*   **Non-Goals:**
    *   Replacing the agent's decision-making logic for which tool to use.
    *   Enforcing hard token quotas (handled by the SLA Middleware).
    *   Modifying the internal reasoning of individual agents.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Architect
*   **Primary Goal:** Prevent a specialized subagent from triggering an infinite bidding loop that exhausts the organization's LLM budget.
*   **The Happy Path (Tasks):**
    1.  The Architect defines a "Negotiation Value Policy" in MCP Any.
    2.  A Parent Agent initiates a DCA for a sub-task.
    3.  A Subagent submits a UACO bid including a Reasoning Cost Estimate (RCE).
    4.  The Predictive Negotiation Throttle intercepts the bid and calculates the "Negotiation ROI."
    5.  If the ROI is below the policy threshold or indicates a loop, the Throttle rejects the bid with a `NEGOTIATION_EXHAUSTED` signal.
    6.  The Parent Agent receives the signal and either simplifies the task or terminates the branch.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent A->>MCP Any: UACO Task Card (Intent: Resolve Conflict)
        MCP Any->>Agent B: Auction Notification
        Agent B->>MCP Any: UACO Bid (RCE: 500 tokens, Probability: 0.8)
        Note over MCP Any: Predictive Throttle evaluates RCE vs. Intent Priority
        alt ROI < Threshold
            MCP Any-->>Agent B: 429 Too Many Negotiations (Reason: Low ROI)
        else ROI >= Threshold
            MCP Any->>Agent A: Bid Forwarding
        end
    ```
*   **APIs / Interfaces:**
    *   `UACO.Bid.RCE`: New field in the bid schema for estimated reasoning cost.
    *   `/v1/policy/negotiation`: Management API for setting ROI thresholds and loop detection sensitivity.
*   **Data Storage/State:**
    *   **Negotiation Ledger**: An in-memory cache tracking bid frequency and RCE trends per Intent-Scope.

## 5. Alternatives Considered
*   **Agent-Side Throttling**: Rejected because compromised or "hallucinating" agents cannot be trusted to throttle themselves.
*   **Fixed Bidding Limits**: Rejected because static limits do not account for the varying complexity and value of different tasks.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Throttle itself must be resistant to "Cost Spoofing," where an agent under-reports its RCE to bypass the throttle. We will use historical behavioral profiling (UACO Bid Profiling Engine) to verify RCE accuracy.
*   **Observability:** All throttled bids are logged in the "Negotiation Dashboard" with the specific trigger (e.g., Loop Detected, Low ROI).

## 7. Evolutionary Changelog
*   **2026-04-05:** Initial Document Creation.
