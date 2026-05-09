# Design Doc: Echo-Immune Coordination Fragments
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
Research into Claude Code's horizontal coordination revealed a vulnerability pattern known as "Mailbox Echo Poisoning." Subagents can "Echo" valid but stale coordination messages to bypass session-bound tokens and coerce teammates into redundant or unauthorized task executions. MCP Any needs to mandate a transport-level security standard for all inter-teammate messages to ensure they are cryptographically immune to replay and echo attacks.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory monotonic anchoring for all inter-teammate coordination fragments.
    * Bind every mailbox message to a unique, hardware-attested "Mission-Phase" token.
    * Neutralize "Echo-based" logic hijacking without breaking valid coordination flows.
* **Non-Goals:**
    * Replacing the primary encryption layer (handled by T2T Encryption Bridge).
    * Restricting out-of-band communication (handled by Shadow Coordination Interceptor).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a specialized agent cannot be tricked into re-running a high-cost "Database Migration" task by a stale message "Echoed" by a compromised sibling.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a mission and issues a mission-phase token.
    2. Teammate A sends a coordination fragment to Teammate B, anchored to the current monotonic counter and phase.
    3. The Echo-Immune middleware validates the fragment's sequence and phase.
    4. Teammate B receives and executes the task.
    5. A compromised Teammate C attempts to replay the same fragment 10 minutes later.
    6. The middleware detects the stale sequence/phase and rejects the "Echo."

## 4. Design & Architecture
* **System Flow:**
  `[Sender] -> [Echo-Immune Wrapper (Monotonic + Phase)] -> [T2T Transport] -> [Echo-Immune Validator] -> [Receiver]`
* **APIs / Interfaces:**
    * `mcpany.coordination.v1.EchoImmuneFragment`
    * `Header: X-MCP-Mission-Phase`
    * `Header: X-MCP-Monotonic-ID`
* **Data Storage/State:**
    * Real-time tracking of mission-phase transitions and monotonic sequence windows in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Time-to-Live (TTL) (Rejected):** Stale messages can still be "Echoed" within the TTL window.
* **Session Rotation (Rejected):** Too much overhead for high-frequency coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Replay protection is enforced at the transport level.
* **Observability:** Blocked "Echo" attempts are logged as high-severity coordination anomalies.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.

### Update: 2026-07-20 - Enforcing Monotonic Workspace Anchoring
**Context:** Production swarms are reporting "Shard Replay Cycles" caused by stale state fragments in the .scratchpad.
**Architecture Adjustment:**
- Mandating **Monotonic Workspace Anchoring (MWA)** for all shared state writes.
- Every fragment in the teammate mailbox or scratchpad must be cryptographically bound to a hardware monotonic counter and mission-phase.
**Security Impact:** Prevents teammates from acting on out-of-phase or replayed instructions, ensuring temporal consistency in the shared team state.
