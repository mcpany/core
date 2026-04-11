# Design Doc: Multi-Modal Anchor Pinning (MMAP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As LLM context windows expand to 1M+ tokens, agents are increasingly susceptible to "Attention Drift," where critical mission guardrails are evicted or ignored in favor of high-entropy noise. While text-based anchors exist, the rise of multi-modal agents (processing SVG, binary artifacts, and UI traces) creates a gap: visual instructions and logic diagrams are often lost during aggressive context garbage collection (CWGC).

MMAP extends hardware-locked anchor pinning to non-textual fragments, ensuring that visual guardrails remain permanent in the agent's attention window.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mechanism to mark non-textual context fragments (SVG, Image metadata, Binary headers) as "GC-Immune."
    * Integrate with the **Hardware-Attested Attention Locking (HAAL)** middleware to enforce multi-modal priority.
    * Maintain visual instruction continuity in long-running UI-coding or diagnostic swarms.
* **Non-Goals:**
    * Providing a general-purpose image storage or compression service.
    * Pinning every visual artifact (only mission-critical guardrails are pinned).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Modal UI Architect Agent
* **Primary Goal:** Retain a complex SVG-based design system manifest in context while refactoring 50+ components.
* **The Happy Path (Tasks):**
    1. User provides a design manifest as a set of cryptographically signed SVG fragments.
    2. The agent identifies these as "Visual Guardrails" and requests MMAP pinning.
    3. MCP Any's MMAP Manager validates the fragments and issues a `GC_IMMUNE` token.
    4. As the context window fills with component code, standard CWGC prunes text history but respects the MMAP token.
    5. The design manifest remains at the top of the attention window, preventing "Aesthetic Hallucinations."

## 4. Design & Architecture
* **System Flow:**
    * **Anchor Registry**: Tracks `Fragment-ID -> {Mime-Type, Data-Hash, Pin-Strength}`.
    * **Attention Injector**: Periodically re-asserts pinned fragments at the attention layer using provider-specific headers (e.g., `x-gemini-mmap-pin`).
* **APIs / Interfaces:**
    * `mmap.PinFragment(fragmentID, data, strength) -> PinID`
    * `mmap.GetActiveAnchors(missionID) -> []Fragment`
* **Data Storage/State:**
    * Pinned fragments are stored in a hardware-locked sidecar to prevent tampering by subagents.

## 5. Alternatives Considered
* **Repeating Visuals in Every Turn**: Rejected due to prohibitive token costs and "Token Storms."
* **System Prompt Only**: Rejected as 1M+ token windows often "forget" instructions that aren't actively reinforced in the middle-window attention.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Only user-verified or parent-authorized fragments can be pinned via MMAP.
* **Observability:** Integrated with the "Visual Attention Map" in the UI to show which multi-modal anchors are currently active.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
