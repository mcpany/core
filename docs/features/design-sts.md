# Design Doc: Sovereign Telemetry Scrubber (STS)
**Status:** Draft
**Created:** 2026-11-02

## 1. Context and Scope
As AI agent swarms scale globally, real-time observability (e.g., Global Agent Activity Maps) has become essential for operational monitoring. However, standard telemetry streams often leak sensitive information, including geolocated mission roots, internal tool metadata, and timing patterns that can be used for stylometric profiling. STS acts as an authoritative privacy proxy that intercept and sanitizes all outbound telemetry from MCP Any nodes before it reaches shared dashboards or third-party monitoring services.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically redact sensitive mission-root identifiers from telemetry payloads.
    * Provide "Spatial Fuzzing" for geolocated agent activity to prevent precise node mapping.
    * Normalize telemetry micro-timing to neutralize side-channel profiling.
    * Enforce cryptographically bound "Privacy Profiles" for different telemetry sinks.
* **Non-Goals:**
    * Redacting data within the agent's internal reasoning loop (handled by RAR Engine).
    * Providing long-term telemetry storage (STS is a passthrough proxy).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Enable global agent monitoring without revealing the exact location or intent of sensitive research swarms.
* **The Happy Path (Tasks):**
    1. Architect configures a `Mission-Sensitive` privacy profile in MCP Any.
    2. Global agent swarms begin executing tasks across 5 nodes.
    3. MCP Any nodes generate telemetry events for every tool call.
    4. STS intercepts these events, redacts the `mission_id`, and adds a 10km Gaussian blur to the geographic coordinates.
    5. The sanitized events are forwarded to the Global Agent Activity Map.
    6. The dashboard shows activity in the general region without exposing the specific laboratory location.

## 4. Design & Architecture
* **System Flow:**
    `Telemetry Source` -> `STS Middleware` -> `Redaction Engine` -> `Fuzzing Layer` -> `Secure Sink`
* **APIs / Interfaces:**
    * `ScrubPayload(event TelemetryEvent, profile PrivacyProfile) (SanitizedEvent, error)`
    * `RegisterSink(url string, profile PrivacyProfile) error`
* **Data Storage/State:**
    * Ephemeral transformation rules stored in memory, anchored to TPM-signed config.

## 5. Alternatives Considered
* **Local-Only Telemetry**: Rejected because it prevents the "Single Pane of Glass" observability required for swarm orchestration.
* **Agent-Level Redaction**: Rejected as it increases subagent complexity and is prone to "Intent Leakage" if the agent is compromised.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: STS ensures that even if the monitoring dashboard is compromised, the leaked data is spatially and semantically neutered.
* **Observability**: STS logs its own redaction metrics (e.g., "12 fields scrubbed") to the local Audit Log.

## 7. Evolutionary Changelog
* **2026-11-02:** Initial Document Creation.
