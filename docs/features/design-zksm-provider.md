# Design Doc: Zero-Knowledge Shard Masking (ZKSM) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents increasingly collaborate in heterogeneous meshes (e.g., Claude Code teammates working with OpenClaw specialists), the risk of "Context Exposure" has become a primary barrier to enterprise adoption. Specialist agents often need to perform computations or reasoning on data shards without actually needing to "see" the raw, sensitive content (e.g., PII, API keys, or proprietary logic).

The Zero-Knowledge Shard Masking (ZKSM) Provider enables agents to share "Masked Shards." These shards contain cryptographically obscured mission data that can be used for verification, intent alignment, or specific sub-reasoning tasks while remaining unreadable to the recipient agent. This ensures that privacy is maintained across framework boundaries without breaking the collaborative loop.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable privacy-preserving context sharing between agents in a mesh.
    * Provide hardware-attested proofs that a specialist agent processed a masked shard correctly.
    * Support "Just-in-Time" unmasking for verified high-trust supervisors.
    * Integrate with the Universal Multimodal Memory Bus (UMMB) for sharded state management.
* **Non-Goals:**
    * Providing general-purpose homomorphic encryption for all agent computations (ZKSM is mission-specific).
    * Masking the mission-root intent itself (intents must remain visible for governance).

## 3. Critical User Journey (CUJ)
* **User Persona:** Financial Compliance Specialist Agent
* **Primary Goal:** Verify that a transaction subagent followed all regulatory rules without the transaction subagent seeing the raw PII of the account holder.
* **The Happy Path (Tasks):**
    1. The "Mission Root" masks the account holder PII using the ZKSM Provider.
    2. The "Masked Shard" is shared with the Compliance Specialist via the mesh.
    3. The Specialist performs reasoning on the "Regulatory Metadata" and the masked shard's cryptographic signature.
    4. The Specialist issues an "Audit Passed" token, backed by a Zero-Knowledge proof.
    5. The mission proceeds, and the PII was never exposed to the specialist's attention window.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Mission Root] -->|Raw Data| B[ZKSM Provider]
        B -->|Masking Key| C[Hardware Enclave]
        B -->|Masked Shard| D[Specialist Agent]
        D -->|Computation Proof| B
        B -->|Verification| A
    ```
* **APIs / Interfaces:**
    * `POST /v1/zksm/mask`: Obscures a context fragment for mesh sharing.
    * `POST /v1/zksm/verify`: Validates a computation proof from a masked shard.
* **Data Storage/State:**
    * Masking keys are ephemeral and stored only within the TPM/Secure Enclave.

## 5. Alternatives Considered
* **Data Redaction (PII Scrubber):** Rejected as it removes the data entirely, preventing specialists from performing structural verification.
* **Differential Privacy:** Rejected due to the "Reasoning Loss" seen when injecting noise into agent context.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If an agent attempts to "Force Unmask" a shard, the ZKSM Provider triggers an immediate Mission Termination signal via the ARL Provider.
* **Observability:** Masked vs. Raw context usage is visualized in the `Sovereignty Audit Dashboard`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
