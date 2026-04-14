# Design Doc: ACP-Native Coordination Bridge
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As OpenClaw stabilizes the **Agent Communication Protocol (ACP)** as its primary coordination tier, a significant gap has emerged between legacy Model Context Protocol (MCP) tools and the new ACP-driven swarms. Current MCP tools cannot natively participate in ACP task auctions or state handoffs without manual wrappers.

The ACP-Native Coordination Bridge is designed to allow legacy MCP, gRPC, and UACO tasks to be represented as first-class ACP messages. This ensures that the vast ecosystem of existing tools remains interoperable with the maturing open-source agent infrastructure.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide bidirectional translation between ACP messages and MCP/gRPC/UACO tool calls.
    * Allow MCP Any to act as a "Virtual ACP Node" for connected legacy tools.
    * Support ACP-compliant task bidding for legacy tools using simulated "Capability Cards."
    * Ensure mission-root intent persistence during protocol translation.
* **Non-Goals:**
    * Rewriting existing MCP tools to use the ACP SDK (the bridge handles the translation).
    * Providing a general-purpose message broker for non-agentic traffic.

## 3. Critical User Journey (CUJ)
* **User Persona:** OpenClaw Swarm Developer
* **Primary Goal:** Integrate a legacy MCP-based Database tool into a horizontal ACP-driven research swarm.
* **The Happy Path (Tasks):**
    1. The developer registers the legacy MCP Database tool in MCP Any.
    2. MCP Any initializes the ACP-Native Coordination Bridge for that tool.
    3. An ACP-based research agent broadcasts a "Data Query" task on the ACP bus.
    4. The Bridge translates the task into a standard MCP `call_tool` request.
    5. The legacy tool executes the query and returns a JSON response.
    6. The Bridge wraps the response in an ACP `TaskCompleted` message, including hardware-attested provenance.
    7. The research agent receives the data as if it came from a native ACP peer.

## 4. Design & Architecture
* **System Flow:**
    `[ACP Agent] <--(ACP Message)--> [Coordination Bridge (Translation Layer)] <--(MCP/JSON-RPC)--> [Legacy Tool]`
* **APIs / Interfaces:**
    * `acp.TranslateInbound(acpMessage) -> internalRequest`: Maps ACP tasks to internal tool calls.
    * `acp.TranslateOutbound(internalResponse) -> acpMessage`: Wraps tool results in ACP envelopes.
    * `acp.SimulateAgentCard(toolMetadata) -> AgentCard`: Generates ACP-compliant discovery cards for legacy tools.
* **Data Storage/State:**
    * **Protocol Mapping Registry:** Maintains stateful mappings between ACP MessageIDs and internal RequestIDs.
    * **Attestation Cache:** Stores session-bound hardware signatures for translated messages.

## 5. Alternatives Considered
* **Manual Wrappers:** Rejected due to "Wrapper Fatigue" and the high maintenance overhead for developers.
* **Direct ACP Implementation in Tools:** Rejected as it requires touching the source code of thousands of existing community tools.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Bridge must enforce "Lineage-Aware Context Signing" to ensure that the translated ACP message carries the same security weight as the original intent.
* **Observability:** Integrated with the "ACP Bridge Monitor" in the UI, visualizing real-time protocol translation and message latency.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
