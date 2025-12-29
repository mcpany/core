# Global Search Feature

## Overview
The **Global Search** feature (activated via `Cmd+K` or `Ctrl+K`) provides a centralized command palette for the MCP Any management console. It allows users to quickly navigate between pages, execute commands, and toggle themes, significantly enhancing the user experience and efficiency.

## Implementation Details
- **Component**: `GlobalSearch` (`ui/src/components/global-search.tsx`)
- **Library**: `cmdk` (via `shadcn/ui` components)
- **Styling**: Tailwind CSS with glassmorphism effects (blurred backdrop).
- **Accessibility**: Fully accessible with keyboard navigation and ARIA attributes (verified with unit tests).

## Verification
- **Unit Tests**: `ui/tests/global-search.test.tsx` (Jest + React Testing Library) covers rendering, interactions, and theme switching.
- **E2E Tests**: `ui/tests/e2e/global-search.spec.ts` (Playwright) covers the full user flow in a real browser environment.
- **Visual Audit**: See screenshot below.

## Screenshot
![Global Search](./global_search_audit.png)
