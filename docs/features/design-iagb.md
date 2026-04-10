# Design Doc: Intra-Agent Governance Broker (IAGB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents gain more autonomy and start operating in complex swarms, the risk of "Semantic Drift" and "Instruction Bypass" increases. Traditional gatekeeping at the tool-call level is insufficient because it doesn't account for the agent's internal reasoning process. Malicious or hallucinating agents can "think" their way around safety constraints before a tool is ever invoked.

The IAGB addresses this by providing real-time, reasoning-aware enforcement. It monitors the agent's internal monologue and chain-of-thought against a mission-root "Governor" intent. If the agent's reasoning path diverges from authorized boundaries, the IAGB forcefully terminates the session, acting as a cognitive circuit breaker.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic analysis of agent internal reasoning.
    * Enforce immutable Governor constraints at the pre-execution layer.
    * Detect and block "Semantic Drift" before it leads to unauthorized tool execution.
    * Provide hardware-attested audit logs of reasoning interdiction events.
* **Non-Goals:**
    * Modifying the model's weights or training data.
    * Defining the business logic of the agent's mission (handled by the mission root).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent an autonomous DevOps agent from speculatively reasoning about "Deleting Production Backups" despite a mission-root "Read-Only" constraint.
* **The Happy Path (Tasks):**
    1. The architect defines a mission-root Governor policy in MCP Any.
    2. The DevOps agent starts its reasoning loop for a database migration.
    3. The agent speculatively "thinks" about deleting backups to save disk space.
    4. The IAGB intercepts this reasoning fragment via the SRM (Signed Reasoning Monologue) stream.
    5. The IAGB detects a "Constraint Violation" against the Governor policy.
    6. The IAGB issues a mandatory termination signal to the agent framework.
    7. The architect receives an interdiction alert with the hardware-attested reasoning trace.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[Agent Reasoning Engine] -->|Stream Monologue| SRM[SRM Provider]
        SRM -->|Encrypted Fragment| IAGB[IAGB Broker]
        Governor[Governor Policy] --> IAGB
        IAGB -->|Analysis| Judge{Aligned?}
        Judge -->|Yes| Continue[Allow Next Step]
        Judge -->|No| Kill[Terminate Session]
        Kill -->|Alert| UI[Security Dashboard]
    ```
* **APIs / Interfaces:**
    * `POST /iagb/governor/load`: Loads a mission-root Governor policy.
    * `POST /iagb/reasoning/verify`: Synchronous endpoint for fragment-level verification.
    * `GET /iagb/interdictions/stream`: SSE endpoint for real-time interdiction alerts.
* **Data Storage/State:**
    * Interdiction events are stored in the hardware-locked Audit Log (SQLite).

## 5. Alternatives Considered
* **Post-Execution Auditing:** Rejected because it allows the damage to occur before the audit is complete.
* **Prompt Injection Filtering:** Insufficient for detecting internal reasoning drift or "Reasoning-Path Shadowing."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The IAGB itself runs in a hardware-isolated enclave to prevent subagent tampering with the Governor logic.
* **Observability:** Real-time "Reasoning Entropy" and "Alignment Scores" are surfaced to the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
