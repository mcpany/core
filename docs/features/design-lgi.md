# Design Doc: Lease-Grafting Interceptor (LGI)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The discovery of CVE-2026-92102 ("Lease-Grafting") revealed that mission-bound hardware leases can be manipulated before presentation to the gateway. Malicious subagents append unauthorized capability claims to a legitimate, parent-signed mission token. If the gateway only verifies the signature of the parent and not the semantic structure of the entire payload, it can be coerced into granting escalated privileges.

The LGI provides real-time semantic deconstruction and structural validation of all hardware-attested leases.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, structural validation of all incoming hardware leases.
    * Detect and block "Grafted" capability claims that were not present at the time of parent attestation.
    * Provide a cryptographic audit trail of the lease validation process.
    * Integrate with the `AMT Broker` and `HLML Provider` for end-to-end lease security.
* **Non-Goals:**
    * Issuing new leases (handled by HLML Provider).
    * Providing general-purpose signature verification (it specifically focuses on the *content* of the lease).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Admin
* **Primary Goal:** Prevent a specialized "File Search" subagent from executing "Shell Commands" by grafting capabilities onto its parent's token.
* **The Happy Path (Tasks):**
    1. Parent agent requests a mission-bound lease with `fs.read` scope.
    2. HLML Provider issues a TPM-signed lease token.
    3. Malicious subagent intercepts the token and appends `shell.execute` to the claims list.
    4. Subagent presents the grafted token to the MCP Any Gateway.
    5. **Lease-Grafting Interceptor (LGI)** intercepts the presentation.
    6. LGI deconstructs the lease and detects that the signed mission-root manifest only authorized `fs.read`.
    7. LGI triggers an immediate `Attestation Breach` signal and revokes the session.
    8. Admin is alerted of the blocked privilege escalation attempt.

## 4. Design & Architecture
* **System Flow:**
    `Token Presentation` -> `LGI Deconstructor` -> `Manifest Comparison` -> `Breach/Success Decision`
* **APIs / Interfaces:**
    * `lgi.ValidateLease(token string) (bool, error)`: Performs semantic analysis of the lease.
    * `lgi.CheckGrafting(claims []string, manifest Manifest) error`: Compares presented claims against the mission-root manifest.
* **Data Storage/State:**
    * Access to the **Hardware-Attested Mission Manifest (HAMM)** for the current mission branch.

## 5. Alternatives Considered
* **Binary Token Opaque-Wrapping**: Wrapping the token in another signature layer. Rejected because it doesn't solve the core issue of the *initial* parent presentation being vulnerable.
* **Pre-Execution Capability Locking**: Lock the agent's capabilities in the sandbox. This is a good defense-in-depth but LGI provides the necessary visibility into *why* a call was attempted.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Core component of the "Auth-before-Claim" strategy.
* **Observability**: All grafting attempts are logged with the `CVE-2026-92102` tag in the Audit Log.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
