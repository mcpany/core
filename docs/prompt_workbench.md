# Feature: Prompt Engineer's Workbench

## Overview

The **Prompt Engineer's Workbench** is a new "Premium Enterprise" feature for the MCP Any management console. It transforms the basic Prompts list into a fully interactive development environment for prompt engineering.

![Prompt Engineer's Workbench](.audit/ui/2026-01-06/prompt_engineer_workbench.png)

## Key Capabilities

1.  **Master-Detail View:** Quickly browse the prompt library on the left while editing/viewing details on the right.
2.  **Argument Configuration:** Automatically generates a form based on the prompt's defined arguments (`prompt.arguments`).
3.  **Live Preview:** Execute the prompt with the provided arguments to see the exact messages (System, User, Assistant) that will be sent to the LLM.
4.  **Playground Integration:** One-click "Open in Playground" to take the generated context and start a chat session.
5.  **Search:** fast filtering of prompts by name or description.

## Technical Details

-   **Component:** `ui/src/components/prompts/prompt-workbench.tsx`
-   **Client:** `ui/src/lib/client.ts` extended with `executePrompt` (simulation mode available for UI testing).
-   **Tests:** Unit tests in `ui/tests/components/prompt-workbench.test.tsx` and E2E coverage in `ui/tests/e2e/prompts-workbench.spec.ts`.

## Usage

1.  Navigate to the **Prompts** page.
2.  Select a prompt from the sidebar.
3.  Fill in the argument fields (e.g., `topic`, `language`).
4.  Click **Generate Preview**.
5.  Inspect the JSON output or visual message bubble representation.
6.  Click **Open in Playground** to use the result.
