# Design Doc: Hardware-Attested Cost Attribution (HACA) Provider
**Status:** Draft
**Created:** 2026-03-29

## 1. Context and Scope
As AI agent swarms grow in complexity, particularly in enterprise environments, the "Spiral of Death" (recursive task delegation) has moved from a stability concern to a massive economic risk. Specialist subagents often spawn further sub-processes without parent awareness, leading to "Economic Squatting" where tokens and compute are consumed opaquely.

The Hardware-Attested Cost Attribution (HACA) Provider introduces a cryptographically bound accounting layer to the Universal Agent Bus. It ensures that every token consumed and every millisecond of reasoning time is attributed to a specific mission-root lineage, anchored by hardware (TPM) attestation.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a hardware-bound counter for token and compute consumption.
    * Attribute costs to the full intent-lineage (Mission Root -> Parent -> Subagent).
    * Provide real-time "Economic Circuit Breakers" that halt execution when a lineage budget is exceeded.
    * Integrate with the Reasoning-Budget Firewall (RBF).
* **Non-Goals:**
    * Managing actual financial payments or billing (focus is on attribution and authorization).
    * Optimizing the cost of LLM calls themselves.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Ops Manager
* **Primary Goal:** Track and limit the budget of a "Feature Development" swarm across 10 specialized agents.
* **The Happy Path (Tasks):**
    1. The manager sets a $50 budget for the "Refactor Auth Module" mission root.
    2. The HACA Provider issues a hardware-attested budget token.
    3. As specialist agents (Frontend, Backend, Security) reason, HACA increments the lineage-bound counter.
    4. A subagent attempts a recursive loop that would exceed the $50 cap.
    5. The HACA-integrated RBF interdicts the tool call, provides a "Budget Exhausted" signal, and notifies the manager.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Tool Call] --> RBF[Reasoning-Budget Firewall]
        RBF --> HACA[HACA Provider]
        HACA --> TPM[Hardware TPM Counter]
        TPM --> Result{Within Budget?}
        Result -->|Yes| Execute[Allow Call]
        Result -->|No| Block[Halt & Notify]
    ```
* **APIs / Interfaces:**
    * `GET /haca/attribution/{mission_id}`: Retrieve detailed cost breakdown for a lineage.
    * `POST /haca/budget/assign`: Bind a hardware-attested budget to a new mission root.
* **Data Storage/State:**
    * Cost metadata is stored in the Mesh Audit Log, while active counters are maintained in TPM-bound monotonic registers.

## 5. Alternatives Considered
* **Software-only Counters:** Rejected because subagents in a compromised environment could spoof or reset counters.
* **Centralized Billing Proxy:** Rejected due to latency and the requirement for "Local-Only" sovereignty.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HACA tokens are non-transferable and cryptographically bound to the mission-root identity.
* **Observability:** Real-time cost telemetry is streamed to the Economic Attribution Viewer.

## 7. Evolutionary Changelog
* **2026-03-29:** Initial Document Creation.
