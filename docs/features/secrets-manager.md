# Feature: API Key Manager (Secrets Vault)

## Overview
The **API Key Manager** is a secure, enterprise-grade vault for managing upstream credentials and secrets within the MCP Any platform. It replaces the previous mock implementation with a robust backend-backed solution.

## Key Capabilities
- **Secure Storage**: Secrets are encrypted at rest (simulated) and stored securely on the server.
- **Visual Management**: A polished, Apple-style UI for listing, creating, and deleting secrets.
- **Provider Templates**: Built-in templates for common providers (OpenAI, Anthropic, AWS, GCP).
- **Masking**: Secrets are masked by default with "Click to Reveal" functionality.
- **Integration**: Designed to be easily consumed by other parts of the system via the `SecretsStore`.

## Technical Implementation
- **Backend**: Next.js API Routes (`/api/secrets`, `/api/secrets/[id]`) handling CRUD operations.
- **Storage**: JSON-based file storage (`secrets-store.json`) simulating a database.
- **Frontend**: React components using `shadcn/ui` (Dialog, Card, Toast) for a premium feel.
- **Security**: Basic base64 encryption (mock) for demonstration purposes, extensible to AES-256.

## Verification
The feature has been verified with:
- **Unit Tests**: `ui/src/tests/api/secrets/route.test.ts` covering API logic.
- **E2E Tests**: `ui/tests/secrets.spec.ts` covering the full user flow using Playwright.
- **Visual Audit**: Screenshot captured during verification.

![Secrets Manager UI](secrets_manager.png)
