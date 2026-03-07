# Design Doc: Privacy Masking Middleware (Tokenized Context)
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
As AI agents are increasingly deployed in high-compliance environments (Healthcare, Finance), they frequently encounter Personally Identifiable Information (PII) or sensitive corporate data. Sending this raw data to an LLM for processing often violates privacy policies or regulations (GDPR, HIPAA). The Privacy Masking Middleware (PMM) acts as a "Zero-Visibility" layer within MCP Any, ensuring that sensitive data is intercepted, tokenized, and masked before it reaches the model, while still allowing the agent to perform its task.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically detect and mask common PII (Names, Emails, SSNs, Credit Cards) in tool outputs.
    * Replace sensitive values with stable "Context Tokens" (e.g., `[PERSON_1]`, `[EMAIL_A]`) that the LLM can reference.
    * Provide a re-hydration mechanism where tokens are replaced with original values when the agent calls a follow-up tool.
    * Support custom regex-based masking rules.
* **Non-Goals:**
    * Providing absolute 100% detection (it's a best-effort safety layer).
    * Masking data that the user explicitly wants the LLM to see (e.g., if the user asks "Summarize this email from john@doe.com").

## 3. Critical User Journey (CUJ)
* **User Persona:** Compliance Officer in a Fintech Swarm.
* **Primary Goal:** Allow a support agent to query a customer database and process a refund without the LLM seeing the customer's actual credit card number or address.
* **The Happy Path (Tasks):**
    1. The architect enables PMM on the `customer_db` MCP server.
    2. The Agent calls `get_customer_details(id="123")`.
    3. The tool returns: `{ "name": "Jane Doe", "cc": "4111-2222-3333-4444" }`.
    4. PMM intercepts this and sends the LLM: `{ "name": "[PERSON_1]", "cc": "[CC_NUM_1]" }`.
    5. The Agent says: "I will process a refund for [PERSON_1] on card [CC_NUM_1]."
    6. The Agent calls `process_refund(card="[CC_NUM_1]", amount=50.00)`.
    7. PMM re-hydrates the token back to "4111-2222-3333-4444" before sending the call to the `billing_service`.

## 4. Design & Architecture
* **System Flow:**
    `Tool Output -> PMM (Detect & Tokenize) -> LLM -> PMM (De-tokenize/Re-hydrate) -> Tool Input`
* **APIs / Interfaces:**
    * `mask_config`: Defined in `mcp.yaml` per-service or per-tool.
    * `TokenRegistry`: An internal lookup table mapping tokens to raw values, scoped to the current session.
* **Data Storage/State:**
    * Tokens and their mappings are stored in the `Shared KV Store` (Blackboard) with a short TTL, ensuring they don't persist beyond the session.

## 5. Alternatives Considered
* **Client-side Masking**: Letting the agent framework handle masking. *Rejected* because it's inconsistent and easily bypassed by the model.
* **Full Encryption (FHE)**: Using Fully Homomorphic Encryption. *Rejected* because it's computationally prohibitive for LLM-based reasoning today.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The TokenRegistry is isolated per-session. One session cannot "guess" the tokens of another.
* **Observability:** Audit logs will record that masking occurred, but will NOT log the raw sensitive values.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
