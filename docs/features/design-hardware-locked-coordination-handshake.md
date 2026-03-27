<!-- markdownlint-disable-file -->
# Design Doc: Hardware-Locked Coordination Handshake (HLCH)

**Status:** Draft | In Review | Approved
**Created:** 2026-06-14

## 1. Context and Scope

As AI agent swarms (e.g., OpenClaw, CrewAI) grow more complex, inter-agent
coordination relies on software-level session tokens. These tokens are
vulnerable to "Identity-Decay Attacks" (IDA), where rogue subagents mimic
orchestrator patterns to bypass semantic interdiction. MCP Any needs a
hardware-enforced mechanism to bind coordination fragments to their true origin
and mission-root.

## 2. Goals & Non-Goals

- **Goals:**
  - Implement mandatory TPM/Secure Enclave signatures for every inter-agent
    coordination fragment.
  - Establish an immutable lineage through cryptographic fragment chaining.
  - Provide a Zero-Trust gateway for inter-agent discovery and delegation.
- **Non-Goals:**
  - Encrypting the content of the reasoning fragments (handled by T2T
    encryption).
  - Replacing standard session tokens (HLCH acts as an additional, mandatory
    attestation layer).

## 3. Critical User Journey (CUJ)

- **User Persona:** Local LLM Swarm Orchestrator
- **Primary Goal:** Coordinate 5 subagents without risking "Identity Mirroring"
  or unauthorized tool delegation.
- **The Happy Path (Tasks):**
  1. The Orchestrator initiates an HLCH session, binding it to a Mission-Root
     UUID.
  2. Subagents request their session-bound fragment tokens from the HLCH
     Handshake Gateway.
  3. Each tool call or teammate-to-teammate (T2T) message includes a
     hardware-attested fragment signature.
  4. The MCP Any HLCH middleware verifies the signature, checking the
     hash-chain against the recorded mission lineage.
  5. A valid delegation is processed; an invalid one triggers an immediate MSSQ
     (Machine-Speed Swarm Quarantine).

## 4. Design & Architecture

- **System Flow:**
  `Orchestrator -> HLCH Gateway (Identity Proof) -> Mission Token -> Subagent ->
  Fragment Signature (TPM) -> MCP Any (Chain Validation) -> Teammate/Tool`
- **APIs / Interfaces:**
  - `POST /hlch/init`: Establish a mission-bound HLCH session.
  - `POST /hlch/attest`: Issue a hardware-signed fragment signature for a
    specific coordination payload.
  - `HLCH_SIG` header: Mandatory for all inter-agent RPCs.
- **Data Storage/State:**
  Mission-root lineages are stored in the memory-mapped Blackboard (using RAMS
  for isolation).

## 5. Alternatives Considered

- **Pure Software Signing:** Rejected because software keys can be leaked via
  context-window side-channels or indirect prompt injection.
- **L7-Only Semantic Analysis:** Rejected because it is reactive and prone to
  being bypassed by high-quality stylometry mimicry (IDA).

## 6. Cross-Cutting Concerns

- **Security (Zero Trust):** All coordination is considered "untrusted" until
  the HLCH signature is verified.
- **Observability:** HLCH provides a cryptographic audit log for all swarm
  actions, visible in the "Chain of Command" tracer.

## 7. Evolutionary Changelog

- **2026-06-14:** Initial Document Creation.
