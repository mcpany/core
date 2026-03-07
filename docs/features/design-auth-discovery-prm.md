# Design Doc: MCP Protected Resource Metadata (PRM) Support
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
The MCP Specification (2025-11-25) introduced Protected Resource Metadata (PRM) as a standardized mechanism for servers to expose their authorization requirements. Currently, MCP Any relies on manual configuration for upstream authentication. By implementing PRM support, MCP Any can automatically discover and negotiate the necessary credentials for any compliant upstream service, significantly reducing configuration friction.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically discover PRM documents via "well-known" URLs or `WWW-Authentication` headers.
    * Map PRM requirements to MCP Any's internal `UpstreamAuth` configurations.
    * Support the 2025-11-25 icon and website metadata for tools/resources.
* **Non-Goals:**
    * Implementing a full OAuth2 server (MCP Any remains a client/proxy).
    * Bypassing security; PRM only *discovers* requirements, it does not fulfill them without authorized secrets.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer connecting a new Enterprise MCP Server.
* **Primary Goal:** Connect a secure upstream server without manually hunting for Auth header formats.
* **The Happy Path (Tasks):**
    1. User adds an upstream URL to `mcpany`.
    2. MCP Any probes the URL and detects a `WWW-Authentication` challenge with a PRM link.
    3. MCP Any fetches the PRM document and identifies that "OAuth2 Client Credentials" are required.
    4. MCP Any prompts the user (via UI or CLI) for the necessary `ClientID` and `ClientSecret`.
    5. MCP Any saves the config and successfully connects.

## 4. Design & Architecture
* **System Flow:**
    - **Discovery Probe**: When a new service is added, the `ServiceRegistry` performs an initial HEAD/GET request.
    - **PRM Resolver**: Parses `WWW-Authentication` headers or checks `/.well-known/mcp-server` for the PRM document.
    - **Negotiation Engine**: Matches discovered requirements against available local secrets.
* **APIs / Interfaces:**
    - New `InternalAuthDiscovery` interface in `pkg/upstream`.
* **Data Storage/State:** Discovered PRM metadata is cached alongside the service configuration in the SQLite DB.

## 5. Alternatives Considered
* **Manual Configuration Only**: Status quo. *Rejected* as it doesn't scale with the growing ecosystem of secure MCP servers.
* **Hardcoded Auth Logic**: Writing custom auth logic for every known server. *Rejected* in favor of the standardized PRM protocol.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PRM documents must be fetched over HTTPS. MCP Any will validate the TLS certificate and ensure the PRM domain matches the service domain.
* **Observability:** Log PRM discovery attempts and results in the Service Diagnostic tool.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
