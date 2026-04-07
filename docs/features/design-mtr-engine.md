# Design Doc: Moving Target Reasoning (MTR) Engine
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
The massive source code leak of Claude Code (512,000 lines) has introduced a new class of threat: **Blueprint-based Attacks**. Attackers can now use the leaked source to build "Inverse Reasoning Models" that predict how an agent will respond to specific context prompts. This allows for high-fidelity context poisoning and "Coaxing" attacks that bypass static guardrails. The Moving Target Reasoning (MTR) Engine introduces non-deterministic entropy into the agent's cognitive path, making it impossible to predict or blueprint the agent's internal state.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically inject randomized "Reasoning Salt" (invisible metadata) into the agent's context window.
    * Shuffle "Attention Anchors" periodically to ensure reasoning paths are non-linear and non-reproducible.
    * Neutralize "Blueprint-based" context poisoning by breaking the predictability of the LLM's attention mechanism.
    * Provide hardware-attested proof that MTR was active during a tool call.
* **Non-Goals:**
    * Modifying the model weights themselves.
    * Changing the user's primary intent or instructions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Developer.
* **Primary Goal:** Prevent an attacker who knows the agent's source code from crafting a "perfect" prompt injection that evades detection.
* **The Happy Path (Tasks):**
    1. Agent initializes a reasoning loop to solve a coding task.
    2. MTR Engine generates a session-unique "Entropy Seed."
    3. Engine shuffles the order of system instructions and injected context shards without changing their semantic meaning.
    4. Engine injects high-entropy "Reasoning Salt" (e.g., randomized GUIDs in comments) into the prompt.
    5. The LLM generates a response based on this unique cognitive environment.
    6. An attacker attempting a "Blueprint attack" fails because their local replica of the agent lacks the unique entropy injected by MTR.

## 4. Design & Architecture
* **System Flow:**
    `Primary Intent` -> `MTR Engine (Entropy Injection)` -> `Randomized Attention Map` -> `LLM Reasoning` -> `Output Validation`
* **APIs / Interfaces:**
    * `EntropyGenerator`: Generates session-bound non-deterministic seeds.
    * `ContextShuffler`: Randomizes non-dependent prompt fragments.
    * `SaltInjector`: Middleware for injecting invisible reasoning anchors.
* **Data Storage/State:**
    * Session seeds are stored in the "Blackboard" and are cryptographically purged upon session termination.

## 5. Alternatives Considered
* **Static Guardrails**: Rejected because they are easily blueprinted and bypassed once the code is leaked.
* **Frequent Model Fine-tuning**: Too expensive and slow to counter machine-speed agents like "shannon."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MTR ensures that "Possession of the Source" does not equal "Possession of the Session."
* **Observability:** "Entropy Strength" metrics are reported to the Security Dashboard.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
