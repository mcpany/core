# Global Search Feature

## Overview
The **Global Search** feature (accessible via `Cmd+K` or `Ctrl+K`) provides a centralized command palette for navigating the MCP Any application. It offers quick access to:
- **Pages:** Dashboard, Services, Resources, Prompts, Tools, Settings.
- **Services:** Direct search and navigation to registered upstream services.
- **Actions:** Contextual actions (future extensibility).
- **Theme:** Rapid switching between Light, Dark, and System themes.

## Verification
The feature has been implemented and verified using Playwright.

![Global Search](global_search_open.png)

## Technical Details
- **Component:** `ui/src/components/global-search.tsx`
- **Library:** `cmdk` (wrapped via `ui/src/components/ui/command.tsx`)
- **Integration:** Added to `ui/src/app/layout.tsx` for global availability.
- **Tests:**
    - Unit: `ui/src/tests/global-search.test.tsx`
    - E2E: `ui/tests/global-search.spec.ts`
