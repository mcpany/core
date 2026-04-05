# Design Doc: Ephemeral Pipe Attestation (EPA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Local multi-agent swarms increasingly rely on Docker-bound named pipes (UNIX domain sockets) for high-performance coordination. However, the emergence of "Pipe Replay" attacks—where malicious subagents or local processes capture and replay coordination fragments—threatens the integrity of the mesh.

The Ephemeral Pipe Attestation (EPA) service provides sub-millisecond, message-bound cryptographic security for all local pipe-based communication, ensuring that every interaction is unique, time-bound, and non-reusable.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement per-message cryptographic key rotation for all named pipe coordination.
    * Mandate hardware-attested handshakes for pipe initialization.
    * Provide sub-100ms latency for secure local coordination.
    * Integrate with the Non-Blocking AMS Core for sharded mailbox security.
* **Non-Goals:**
    * EPA will not secure remote network-based transport (handled by mTLS).
    * EPA will not perform semantic analysis of the message content (handled by ISD).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 3 parallel teammates over local pipes without risk of message replay or session hijacking.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes the Docker-bound named pipe mesh.
    2. EPA issues a hardware-attested root key for the session.
    3. Teammate A sends a task-claim message to the shared mailbox shard.
    4. EPA rotates the message-bound key and signs the fragment.
    5. Teammate B receives and validates the fragment using the ephemeral key.
    6. The key is instantly revoked, neutralizing any captured replay attempts.

## 4. Design & Architecture
* **System Flow:**
    `Teammate A -> [Message + Ephemeral Key] -> EPA Provider -> [Validation] -> Named Pipe -> [Encrypted Fragment] -> Teammate B`
* **APIs / Interfaces:**
    * `rotate_pipe_key(pipe_id)`: Issues a new message-bound token.
    * `attest_pipe_message(fragment, token)`: Hardware-bound validation hook.
* **Data Storage/State:**
    * Ephemeral keys are stored in kernel-resident secure memory, inaccessible to subagent process environments.

## 5. Alternatives Considered
* **Static Session Tokens**: Rejected due to vulnerability to replay and "Pipe-Squatting" exploits.
* **Full mTLS over Pipes**: Rejected due to prohibitive handshake latency (20ms+) for high-frequency coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "The Temporal Fence." Every message is a one-time cryptographic event.
* **Observability:** EPA logs all key rotation events and attestation failures to the Local Security Audit Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
