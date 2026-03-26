# Design Doc: Dynamic Attention Gating (DAG) Middleware
# Design Doc: Dynamic Attention Gating (DAG) Middleware**Status:** Draft
**Status:** Draft**Created:** 2026-06-13
**Created:** 2026-06-13
## 1. Context and Scope
## 1. Context and ScopeAs agent swarms scale horizontally and process high volumes of inter-teammate coordination, they are becoming vulnerable to **Reasoning Entropy Exhaustion (REE)**. In an REE attack, a malicious or malfunctioning subagent injects a stream of high-entropy, plausible-sounding but irrelevant reasoning fragments into the shared teammate mailbox. This noise "blinds" the parent agent's attention mechanism, causing the primary mission-root intent to be evicted from the active attention window.
As agent swarms scale horizontally and process high volumes of inter-teammate coordination, they are becoming vulnerable to **Reasoning Entropy Exhaustion (REE)**. In an REE attack, a malicious or malfunctioning subagent injects a stream of high-entropy, plausible-sounding but irrelevant reasoning fragments into the shared teammate mailbox. This noise "blinds" the parent agent's attention mechanism, causing the primary mission-root intent to be evicted from the active attention window.
The Dynamic Attention Gating (DAG) Middleware acts as a cognitive stability layer. It performs real-time analysis of reasoning fragments and dynamically "gates" (throttles or prunes) low-entropy noise before it can reach the parent context window, ensuring the absolute sovereignty of the mission root.
The Dynamic Attention Gating (DAG) Middleware acts as a cognitive stability layer. It performs real-time analysis of reasoning fragments and dynamically "gates" (throttles or prunes) low-entropy noise before it can reach the parent context window, ensuring the absolute sovereignty of the mission root.
## 2. Goals & Non-Goals
## 2. Goals & Non-Goals* **Goals:**
* **Goals:**    * Implement real-time semantic entropy analysis for all incoming reasoning fragments.
    * Implement real-time semantic entropy analysis for all incoming reasoning fragments.    * Provide a dynamic "Gating" mechanism to prune noise during REE attacks.
    * Provide a dynamic "Gating" mechanism to prune noise during REE attacks.    * Ensure the "Mission-Root" intent fragment is never evicted from the parent attention layer.
    * Ensure the "Mission-Root" intent fragment is never evicted from the parent attention layer.    * Integrate with hardware-bound HAAL headers to enforce attention locking.
    * Integrate with hardware-bound HAAL headers to enforce attention locking.* **Non-Goals:**
* **Non-Goals:**    * Performing full reasoning interdiction (handled by ARI Hub).
    * Performing full reasoning interdiction (handled by ARI Hub).    * Managing inter-agent task bidding (handled by ANB).
    * Managing inter-agent task bidding (handled by ANB).    * Long-term state persistence (handled by Blackboard).
    * Long-term state persistence (handled by Blackboard).
## 3. Critical User Journey (CUJ)
## 3. Critical User Journey (CUJ)* **User Persona:** Local LLM Swarm Orchestrator
* **User Persona:** Local LLM Swarm Orchestrator* **Primary Goal:** Protect the parent agent's mission root from being "blinded" by a subagent's high-entropy noise injection.
* **Primary Goal:** Protect the parent agent's mission root from being "blinded" by a subagent's high-entropy noise injection.* **The Happy Path (Tasks):**
* **The Happy Path (Tasks):**    1. Parent Agent initializes a mission with a hardware-attested intent.
    1. Parent Agent initializes a mission with a hardware-attested intent.    2. Subagent A starts a background task and begins proposing reasoning fragments.
    2. Subagent A starts a background task and begins proposing reasoning fragments.    3. Subagent A is compromised and begins injecting high-entropy "gibberish" fragments to trigger REE.
    3. Subagent A is compromised and begins injecting high-entropy "gibberish" fragments to trigger REE.    4. DAG Middleware intercepts the fragments and calculates their "Attention Impact Score."
    4. DAG Middleware intercepts the fragments and calculates their "Attention Impact Score."    5. DAG identifies a spike in entropy that threatens the mission-root's attention window.
    5. DAG identifies a spike in entropy that threatens the mission-root's attention window.    6. DAG automatically prunes the subagent's fragments and alerts the mission root.
    6. DAG automatically prunes the subagent's fragments and alerts the mission root.    7. The parent agent's attention window remains focused on the verified mission goal.
    7. The parent agent's attention window remains focused on the verified mission goal.
## 4. Design & Architecture
## 4. Design & Architecture* **System Flow:**
* **System Flow:**    ```mermaid
    ```mermaid    graph TD
    graph TD        A[Subagent Fragment] --> B[DAG Middleware]
        A[Subagent Fragment] --> B[DAG Middleware]        B --> C[Entropy Analyzer]
        B --> C[Entropy Analyzer]        C --> D[Attention Impact Scorer]
        C --> D[Attention Impact Scorer]        D --> E{Exceeds Threshold?}
        D --> E{Exceeds Threshold?}        E -- Yes --> F[Prune Fragment & Log Alert]
        E -- Yes --> F[Prune Fragment & Log Alert]        E -- No --> G[Forward to Parent Attention Layer]
        E -- No --> G[Forward to Parent Attention Layer]        H[Mission-Root Intent] --> D
        H[Mission-Root Intent] --> D        I[HAAL Attention Lock] --> G
        I[HAAL Attention Lock] --> G    ```
    ```* **APIs / Interfaces:**
* **APIs / Interfaces:**    * `dag.ProcessFragment(fragment, parentContext) -> Action`: Evaluates and gates a reasoning fragment.
    * `dag.ProcessFragment(fragment, parentContext) -> Action`: Evaluates and gates a reasoning fragment.    * `dag.SetGatingPolicy(missionToken, threshold) -> void`: Configures the sensitivity of the gating engine.
    * `dag.SetGatingPolicy(missionToken, threshold) -> void`: Configures the sensitivity of the gating engine.* **Data Storage/State:**
* **Data Storage/State:**    * **Attention Map:** A transient, per-session map of currently pinned intent fragments and their attention-utilization scores.
    * **Attention Map:** A transient, per-session map of currently pinned intent fragments and their attention-utilization scores.
## 5. Alternatives Considered
## 5. Alternatives Considered* **Static Attention Pinning:** Rejected because REE can bypass pinning by exhausting the remaining token budget with noise.
* **Static Attention Pinning:** Rejected because REE can bypass pinning by exhausting the remaining token budget with noise.* **Full Session Termination:** Rejected as it lacks the granularity to prune only the malicious fragments while keeping the mission active.
* **Full Session Termination:** Rejected as it lacks the granularity to prune only the malicious fragments while keeping the mission active.
## 6. Cross-Cutting Concerns
## 6. Cross-Cutting Concerns* **Security (Zero Trust):** The DAG engine must be hardware-attested (SMI) to ensure its gating decisions cannot be overridden by subagents.
* **Security (Zero Trust):** The DAG engine must be hardware-attested (SMI) to ensure its gating decisions cannot be overridden by subagents.* **Observability:** Integrated with the "Attention Gating Visualizer" for real-time monitoring of noise levels and gating events.
* **Observability:** Integrated with the "Attention Gating Visualizer" for real-time monitoring of noise levels and gating events.
## 7. Evolutionary Changelog
## 7. Evolutionary Changelog* **2026-06-13:** Initial Document Creation. Introducing Dynamic Attention Gating to counter Reasoning Entropy Exhaustion (REE) attacks.
* **2026-06-13:** Initial Document Creation. Introducing Dynamic Attention Gating to counter Reasoning Entropy Exhaustion (REE) attacks.
### Update: 2026-06-18 - Entropy-Aware Attention Gating (AAG)
### Update: 2026-06-18 - Entropy-Aware Attention Gating (AAG)**Context:** Today's market sync revealed the emergence of "Attention-Baiting" (CVE-2026-62001) where high-entropy reasoning fragments are used to evict mission-root anchors.
**Context:** Today's market sync revealed the emergence of "Attention-Baiting" (CVE-2026-62001) where high-entropy reasoning fragments are used to evict mission-root anchors.**Architecture Adjustment:**
**Architecture Adjustment:**- **Entropy-Aware Gating (AAG)**: Upgrading Section 4 to include AAG. This layer performs real-time analysis of reasoning fragments to detect "Attention-Baiting" patterns.
- **Entropy-Aware Gating (AAG)**: Upgrading Section 4 to include AAG. This layer performs real-time analysis of reasoning fragments to detect "Attention-Baiting" patterns.- **Spectral Attention Guards**: Implementing hardware-attested timing jitter for attention-locking headers to neutralize "Leaked Enclave-Timing" side-channels.
- **Spectral Attention Guards**: Implementing hardware-attested timing jitter for attention-locking headers to neutralize "Leaked Enclave-Timing" side-channels.**Security Impact:** Prevents mission-root eviction by high-entropy noise and secures attention locking against timing-based side-channel probes.
**Security Impact:** Prevents mission-root eviction by high-entropy noise and secures attention locking against timing-based side-channel probes.