# Design Doc: Negotiation Storm Circuit Breaker (NSCB)
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
Autonomous swarms using Distributed Capability Bidding (DCA) are increasingly susceptible to **Recursive Negotiation Storms**. In high-density environments, agents can enter infinite bidding loops for low-priority tasks, leading to token exhaustion and cognitive stall.

The NSCB acts as a high-speed stability gate that monitors the frequency and convergence of agent bidding, automatically interdicting "Negotiation Storms" before they impact mesh performance.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect non-convergent bidding patterns in UACO task auctions.
    * Implement autonomous "Cool-down" periods for hyper-active bidders.
    * Enforce bidding-depth limits for recursive delegations.
* **Non-Goals:**
    * Modifying the internal bidding logic of individual agents.
    * Replacing the DCA Negotiation Guard (NSCB is a circuit breaker for the Guard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Performance Engineer
* **Primary Goal:** Prevent a fleet of 100+ subagents from entering a "Spiral of Death" bidding loop.
* **The Happy Path (Tasks):**
    1. Swarm enters a complex task negotiation phase.
    2. NSCB monitors the bid-to-execution ratio on the coordination bus.
    3. NSCB detects a "Storm" pattern (high-frequency bids without task commitment).
    4. NSCB triggers an autonomous **Auction Pause** for the affected mission branch.
    5. NSCB mandates a hardware-attested **Mission Realignment** signal before resuming bidding.
    6. Token exhaustion is averted, and the mesh recovers.

## 4. Design & Architecture
* **System Flow:**
    `[UACO Bus] -> [NSCB Monitor] -> [Entropy Threshold Check] -> [Circuit Trip] -> [Realignment Gate]`
* **APIs / Interfaces:**
    * `GET /v1/mesh/stability`: Returns current bidding health metrics.
    * `POST /v1/mesh/reset/{branch_id}`: Manual override to reset a tripped breaker.
* **Data Storage/State:**
    * Sliding-window counters for bid frequency, stored in volatile memory.

## 5. Alternatives Considered
* **Static Rate Limiting (Rejected):** Too blunt; penalizes healthy high-speed coordination.
* **Human-in-the-Loop Approval (Rejected):** Too slow to prevent machine-speed token spikes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Circuit-breaker triggers are hardware-attested to prevent subagent "Trip Spoofing."
* **Observability:** Real-time visualization of "Negotiation Entropy" in the CSAD Hub.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.
