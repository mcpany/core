# Design Doc: Global Agent Activity Map
**Status:** Draft
**Created:** 2026-11-02

## 1. Context and Scope
As agentic swarms grow in complexity and geographic distribution, human supervisors require high-fidelity situational awareness. Standard tables and logs fail to communicate the spatial scale and real-time flow of operations. This feature introduces a "Global Agent Activity Map" to the MCP Any dashboard, providing an "Apple-level" interactive visualization of geolocated tool usage.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide real-time visualization of geolocated agent activity.
    * Use glowing nodes and arcs to represent active tool calls and inter-node coordination.
    * Implement pulsing animations for "Active" vs. "Error" states.
* **Non-Goals:**
    * Providing precise GPS coordinates (limited to city/region level for privacy).
    * Serving as a replacement for detailed execution traces.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Fleet Supervisor
* **Primary Goal:** Identify a regional infrastructure outage affecting agent performance across 5 different frameworks.
* **The Happy Path (Tasks):**
    1. Supervisor opens the MCP Any dashboard.
    2. The Global Activity Map displays nodes across the US, Europe, and Asia.
    3. Several nodes in the `us-east-1` region start pulsing crimson (Error state).
    4. Supervisor hovers over a crimson node to see the frosted-glass tooltip.
    5. Tooltip reveals that multiple Claude Code and OpenClaw agents are failing due to a localized gRPC timeout.
    6. Supervisor initiates a regional failover policy.

## 4. Design & Architecture
* **System Flow:**
    * Frontend: React component using SVG or WebGL for the globe rendering.
    * Backend: Telemetry aggregator in the server collects `location_metadata` from tool calls.
    * WebSocket: Real-time event stream pushes geolocated execution events to the UI.
* **APIs / Interfaces:**
    * New `GET /api/v1/telemetry/spatial-activity` endpoint.
    * WebSocket event type: `AGENT_ACTIVITY_GEOLOCATED`.

## 5. Alternatives Considered
* **Static Image Map**: Rejected due to lack of interactivity and real-time feel.
* **External BI Tool Integration**: Rejected to maintain MCP Any as a self-contained, high-premium infrastructure layer.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Anonymize location data to city/region level. Ensure location metadata is only included if configured in the Privacy Policy.
* **Observability:** Monitor the performance impact of high-frequency telemetry on the frontend UI.

## 7. Evolutionary Changelog
* **2026-11-02:** Initial Document Creation.
