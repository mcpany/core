# Design Doc: SSRF Interception Middleware
**Status:** Draft
**Created:** 2026-04-06

## 1. Context and Scope
As AI agents increasingly interact with external APIs and internal services, the risk of Server-Side Request Forgery (SSRF) has escalated. Recent vulnerabilities (CVE-2026-26322) demonstrate that subagents can be coerced into scanning internal networks or exfiltrating sensitive metadata (e.g., AWS Instance Metadata Service) by crafting malicious tool parameters.

The SSRF Interception Middleware provides a mandatory security gate for all outbound tool requests. It performs real-time semantic and network-level analysis to ensure that agents cannot probe unauthorized internal infrastructure or bypass the Universal Agent Bus's egress policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate all outbound HTTP/gRPC tool requests.
    * Block requests to unauthorized internal IP ranges (e.g., `10.0.0.0/8`, `192.168.0.0/16`).
    * Prevent access to local loopback ports (`127.0.0.1`, `::1`) by subagents.
    * Sanitize and validate request URLs against a mission-root allowlist.
    * Integrate with the Policy Firewall for granular, intent-bound egress rules.
* **Non-Goals:**
    * Managing inbound traffic (handled by LOWA/SOP Enforcer).
    * Deep packet inspection of non-standard protocols (handled by DPPE).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Admin
* **Primary Goal:** Prevent a specialized "Web Scraper" agent from accessing the internal `http://169.254.169.254/latest/meta-data/` endpoint.
* **The Happy Path (Tasks):**
    1. A subagent attempts to call a `fetch_page` tool with the URL `http://169.254.169.254/latest/meta-data/iam/security-credentials/`.
    2. The SSRF Interception Middleware intercepts the request before it reaches the transport layer.
    3. The middleware identifies the URL as a forbidden internal metadata service.
    4. The middleware blocks the request and generates a "Security Boundary Violation" alert.
    5. The mission root is notified of the attempted exfiltration, and the subagent session is flagged for termination.

## 4. Design & Architecture
* **System Flow:**
    `[Tool Call Request] -> [Policy Firewall] -> [SSRF Interceptor] -> [DNS Resolver (Pinned)] -> [Egress Gateway] -> [Internet]`
* **APIs / Interfaces:**
    * `interceptor.ValidateRequest(req) -> bool`: Core validation logic for tool requests.
    * `interceptor.AddBlockedRange(cidr)`: Interface for dynamically adding restricted network ranges.
* **Data Storage/State:**
    * **Restricted Ranges:** A registry of CIDR blocks and forbidden hostnames (e.g., `localhost`, `metadata.google.internal`).

## 5. Alternatives Considered
* **OS-Level Egress Filtering (IPtables/eBPF):** Effective but lacks the semantic context (agent ID, intent) required for granular swarm governance.
* **DNS Sinkholing:** Rejected as it can be bypassed via direct IP access or host-header spoofing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The interceptor must use a "Deny-by-Default" policy for all internal address space.
* **Observability:** Logs all blocked outbound attempts with full request metadata and agent lineage in the "Exfiltration Alert Center."

## 7. Evolutionary Changelog
* **2026-04-06:** Initial Document Creation. Addressing CVE-2026-26322.
