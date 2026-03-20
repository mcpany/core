# Design Doc: Verified Skill Auction (VSA)
**Status:** Draft
**Created:** 2026-04-08

## 1. Context and Scope
As agent swarms move toward decentralized coordination (via DCA), the risk of "Malicious Skill Winning" (ClawHavoc incident) increases. MCP Any needs a mechanism that ensures capability negotiation is bound to real-time skill attestation. Verified Skill Auction (VSA) integrates the DCA negotiation bus with the Federated Reputation mesh to ensure only verified, high-trust agents can win sensitive task cards.

## 2. Goals & Non-Goals
* **Goals:**
    * Integrate DCA Auction Broker with real-time Reputation Engine checks.
    * Automate the revocation of high-risk capabilities (e.g., `fs:write`) for agents whose reputation falls below a quorum-defined threshold during an active auction.
    * Implement "Proof-of-Reputation" tokens as part of the UACO bidding process.
* **Non-Goals:**
    * Building a new agent framework (VSA is an infrastructure layer).
    * Managing financial transactions for auctions (it is a capability auction).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Delegate a sensitive file-modification task to the most capable AND trusted subagent.
* **The Happy Path (Tasks):**
    1. Parent agent broadcasts a Task Card for "Refactor sensitive-data-layer.go".
    2. Multiple subagents submit bids via MCP Any's DCA broker.
    3. MCP Any intercepts the bids and cross-references subagent IDs with the Cross-Framework Skill Reputation Engine.
    4. Subagent A has a "High" reputation and a valid attestation token; Subagent B (from an unverified registry) has no reputation history.
    5. MCP Any automatically disqualifies Subagent B or strips its high-risk capabilities before the parent sees the bid.
    6. Subagent A wins the auction and receives a session-bound PoI token for the task.

## 4. Design & Architecture
* **System Flow:**
    `Bid Submission` -> `Reputation Check` -> `Capability Scoping` -> `Auction Resolution` -> `PoI Issuance`
* **APIs / Interfaces:**
    * `AuctionBroker`: `SubmitBid(bid UACOBid) (AttestedBid, error)`
    * `ReputationProvider`: `GetScore(agentID string) (ReputationScore, error)`
* **Data Storage/State:**
    * Reputation scores are cached in the Federated Node mesh.
    * Auction state is transient, managed by the high-speed DCA broker.

## 5. Alternatives Considered
* **Static Block-lists**: Rejected because they cannot handle dynamic "Delayed Payload" behaviors or emerging reputation trends.
* **Manual HITL Approval for Every Bid**: Rejected due to latency bottlenecks in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Continuous validation of reputation during the negotiation lifecycle.
* **Observability:** Visualizing bid attestation status in the VSA Monitor UI.

## 7. Evolutionary Changelog
* **2026-04-08:** Initial Document Creation.
