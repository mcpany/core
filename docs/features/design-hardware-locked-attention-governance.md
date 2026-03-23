<!-- markdownlint-disable -->
# Design Doc: Hardware-Locked Attention Governance (HLAG)

**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope

As agent swarms become deeper and more autonomous, a new class of exploit has emerged: **Reasoning Entropy Exhaustion (REE)**. In REE, a compromised subagent injects high-entropy "noise" into its reasoning traces, forcing the parent agent's context window to shift and evict the "Mission Root" anchors. MCP Any needs to solve this by providing a hardware-locked mechanism to "pin" critical context fragments at the attention layer of the LLM.

## 2. Goals & Non-Goals

* **Goals:**
  * Provide a mechanism to cryptographically "pin" mission-critical intent fragments.
  * Utilize hardware-bound (TPM/SEP) headers to signal attention priority to LLM providers.
  * Automatically detect and prune high-entropy reasoning noise from subagents.
* **Non-Goals:**
  * Modifying the underlying LLM architecture.
  * Providing absolute protection against all forms of context-window limits.

## 3. Critical User Journey (CUJ)

* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Protect the primary mission objective from being "forgotten" by the lead agent during a complex, multi-agent refactor.
* **The Happy Path (Tasks):**
  1. The User defines a "Mission Root" intent via the MCP Any Gateway.
  2. MCP Any signs the intent fragment using the local TPM and attaches a `X-Attention-Lock` header.
  3. As subagents generate high-frequency reasoning, MCP Any's DAG middleware monitors entropy.
  4. The Lead Agent's context window fills up, but the hardware-locked fragment remains pinned.
  5. The Lead Agent successfully synthesizes the work while remaining bound to the original mission.

## 4. Design & Architecture

* **System Flow:**

```mermaid
graph TD
    A[Mission Root] --> B[HLAG Provider]
    B --> C{TPM Signing}
    C --> D[Attention-Locked Fragment]
    E[Subagent Reasoning] --> F[DAG Middleware]
    F --> G[Entropy Scorer]
    G --> H{Pruning Trigger}
    H -- Low Entropy --> I[Context Window]
    D --> I
    H -- High Entropy --> J[Discarded Noise]

```

## 5. Alternatives Considered

* **Software-only Pinning**: Rejected because it can be bypassed if the subagent gains control of the local transport.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust)**: All attention-lock requests must be accompanied by a valid hardware-attested mission token.

## 7. Evolutionary Changelog

* **2026-06-18:** Initial Document Creation.

### Update: [2026-06-19] - Attention-Locking and HAIL Integration

**Context:** Today's market sync revealed a new exploit pattern in OpenClaw subagent reasoning mimicry.
**Architecture Adjustment:**
* Integrating support for `x-gemini-attention-lock` compatibility in Section 4.
* Mandating **HAIL-attested lineage** for all attention-lock requests.
