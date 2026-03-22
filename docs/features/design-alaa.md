# Design Doc: Attention-Locked Auction Arbiter (ALAA)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
The rise of "Speculative Bidding" in UACO v2.5 (where agents bid based on probable intent) is causing "Negotiation Deadlocks" in horizontal swarms. Disparate agents cannot reach consensus on task priority during high-frequency cycles. MCP Any must provide an authoritative, hardware-locked arbiter that prioritizes mission-root intent fragments to resolve these deadlocks.

## 2. Goals & Non-Goals
* **Goals:**
    * Resolve speculative bidding deadlocks in horizontal swarms.
    * Use hardware-bound attention-locking headers to prioritize mission-root intents.
    * Provide a high-speed, non-blocking auction bus.
* **Non-Goals:**
    * Defining the bidding logic (this is the agent's responsibility).
    * Managing the entire swarm lifecycle (this is handled by the DSM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Speculative Swarm Orchestrator
* **Primary Goal:** Resolve a task priority conflict between 3 specialized teammates within 100ms.
* **The Happy Path (Tasks):**
    1. Teammates submit speculative bids to the ALAA auction bus.
    2. The ALAA detects a negotiation deadlock (no consensus).
    3. The ALAA injects a hardware-bound "Attention-Lock" fragment from the Mission Root.
    4. The ALAA arbitrates the auction based on the locked intent.
    5. The winning agent is assigned the task, and the others roll back.

## 4. Design & Architecture
* **System Flow:**
    `Agents -> Speculative Bids -> ALAA (Deadlock Detection) -> Attention-Locked Arbitration -> Winner Assigned`
* **APIs / Interfaces:**
    * `POST /v1/auction/bid`: Submit a speculative task bid.
    * `GET /v1/auction/{id}/consensus`: Get real-time auction status.
* **Data Storage/State:**
    * Bidding state stored in the Shared KV Store (Blackboard).
    * Arbitration rules defined by the Mission Root manifest.

## 5. Alternatives Considered
* **User-in-the-Loop (HITL):** Rejected due to 5s+ latency causing agentic stall.
* **Round-Robin Assignment:** Rejected as it ignores Mission-Root priority.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All auction arbitration requires hardware-attested Mission Root tokens.
* **Observability:** Real-time metrics for "Negotiation Latency" and "Deadlock Frequency."

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
