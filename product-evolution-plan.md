# Product Evolution Plan

## 1. Updated Roadmap

The [Roadmap](docs/roadmap.md) has been synchronized with the current codebase state.

### Key Updates:
- **Completed Items**: Health Checks, Schema Validation, Service Profiles, Message Bus (NATS/Kafka), and Dynamic Registration have been marked as completed and linked to documentation.
- **New Entries**: Semantic Caching and Upstream mTLS have been added to the implemented features list.
- **Clarifications**: Distinctions made between Rate Limiting (Implemented) and Token-Based Quota Management (Planned).

## 2. Top 10 Recommended Features

These recommendations focus on moving MCP Any from a powerful tool to an enterprise-grade platform.

| Priority | Feature Name | Why it matters | Implementation Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Kubernetes Operator** | Enables GitOps, automated scaling, and simplified lifecycle management in production environments. Critical for enterprise adoption. | High |
| 2 | **WASM Plugin Support** | Allows users to extend functionality (custom transformations, validators) safely without recompiling the server. Key for a flexible ecosystem. | High |
| 3 | **Cloud Storage Support (S3/GCS)** | Extends the Filesystem provider to cloud buckets. This allows agents to manipulate cloud assets directly, a massive use-case unlock. | Medium |
| 4 | **Token-Based Quota Management** | Rate limiting protects the server, but Quotas protect the *wallet*. Essential for managing costs when using LLMs. | Medium |
| 5 | **Redis/PgVector for Semantic Cache** | Current SQLite/Memory semantic cache is good for single instances. Redis/PgVector support is needed for distributed, high-scale deployments. | Medium |
| 6 | **Terraform Provider** | "Configuration as Code" is a core philosophy. A Terraform provider allows managing MCP Any resources using standard IaC tools. | Medium |
| 7 | **Downstream mTLS Authentication** | While upstream mTLS is supported, fully enforcing mTLS for *clients* (AI agents) connecting to MCP Any ensures end-to-end zero-trust security. | Medium |
| 8 | **CI/CD Config Validator CLI** | A standalone binary to validate `config.yaml` and `.proto` files in CI pipelines before deployment preventing bad configs from reaching production. | Low |
| 9 | **Streaming Response Support (SSE)** | Ensure full support for Server-Sent Events across all providers to enable real-time feedback from long-running tools to the AI. | Medium |
| 10 | **Data Loss Prevention (DLP)** | Middleware to automatically redact PII (emails, credit cards) from tool inputs/outputs and logs, ensuring compliance. | Medium |

## 3. Codebase Health

### Observations

- **Structure**: The codebase is well-structured with clear separation of concerns (`server/pkg/upstream`, `server/pkg/middleware`, etc.).
- **Documentation**: Documentation is extensive but was slightly fragmented between `docs/` and `server/docs/`. This has been partially reconciled, but a full merge of documentation trees is recommended in the future.
- **Testing**:
    - There is evidence of integration tests (`server/tests/integration`) and unit tests.
    - **Gap**: Coverage for the `filesystem` provider suggests it relies on `afero`, but cloud backend implementations (S3/GCS) are explicitly "not yet supported" in the code, matching the roadmap update.
- **Dependencies**: The project manages dependencies well (`go.work`, `go.mod`).
- **Standardization**: Use of Protocol Buffers (`proto/`) for config definition is a strong architectural choice, ensuring type safety and schema validation.

### Recommendations for Health

1.  **Unify Documentation**: Merge `docs/` and `server/docs/` into a single hierarchy to avoid confusion and drift.
2.  **E2E Testing Suite**: Expand the `server/tests/integration` to cover complex scenarios like "Auth + Rate Limit + Tool Execution" in a single flow.
3.  **Linter Strictness**: Ensure `golangci-lint` is running with a strict ruleset in CI to maintain code quality as the team scales.
