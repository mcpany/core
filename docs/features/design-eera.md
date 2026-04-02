# Design Doc: Enclave-Enforced Resource Allocation (EERA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms become more autonomous and distributed, the risk of "Resource Squatting"—where a specialist agent consumes excessive tokens or compute power without parent oversight—has become a primary economic stability risk. Standard API-level rate limits are insufficient because they can be bypassed by spoofed headers or compromised sub-processes.

The Enclave-Enforced Resource Allocation (EERA) is required to mandate TPM-signed "Budget Tickets" for high-intensity reasoning, ensuring that resources are physically locked to the hardware-attested mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-locked budgeting for LLM tokens and reasoning-effort (ARE).
    * Issue TPM-signed "Budget Tickets" that must be presented by subagents to the Universal Agent Bus.
    * Enable real-time, non-repudiable resource reclamation from dormant missions.
    * Neutralize "Reasoning-Budget Hijacking" by binding tickets to hardware identities.
* **Non-Goals:**
    * Replacing cloud provider billing systems; it acts as a local governor.
    * Managing non-agent system resources (CPU/RAM) unless directly tied to tool execution.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Ops Manager
* **Primary Goal:** Prevent a rogue research agent from exhausting the team's $500 monthly token quota in a single recursive loop.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a "Research" sub-mission with a 50k token budget.
    2. EERA Middleware intercepts the request and generates a "Budget Ticket" signed by the host TPM.
    3. The research subagent receives the ticket and includes it in its reasoning requests.
    4. MCP Any gateway verifies the ticket signature and mission-root ID before proxying to the LLM provider.
    5. EERA tracks consumption in real-time. If the budget is exceeded, the ticket is invalidated at the hardware layer.
    6. Parent agent is notified and can choose to re-attest and top up the budget.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        subgraph Hardware
            A[TPM / Secure Enclave]
        end
        subgraph MCP Any Node
            B[Parent Agent] --> C[EERA Manager]
            C -->|Sign| A
            C --> D[Subagent]
            D -->|Ticket| E[Gateway Proxy]
            E -->|Verify| A
            E --> F[LLM Provider]
        end
    ```
* **APIs / Interfaces:**
    * `eera.IssueTicket(missionRoot, budget) -> TicketID`: Generates a hardware-signed budget.
    * `eera.ValidateTicket(ticketID, requestPayload) -> Success`: Verifies budget and signature.
    * `eera.ReclaimResources(ticketID) -> RemainingBudget`: Forcefully revokes a budget.
* **Data Storage/State:**
    * **Ticket Ledger:** Local encrypted database tracking active tickets and consumption.
    * **Quota Policy Store:** Centralized or local policy file defining default budgets per agent framework.

## 5. Alternatives Considered
* **HTTP Header Injection (ARE):** Rejected as primary enforcement because headers can be stripped or modified in transit. EERA uses them for signaling but relies on TPM for enforcement.
* **Purely Virtual Quotas:** Rejected because they don't provide the non-repudiation required for enterprise auditability.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tickets are cryptographically bound to the mission root. One mission cannot "steal" the ticket of another.
* **Observability:** Integrated with the "Mission Budget Dashboard" in the UI for real-time visualization of budget burn rates.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
