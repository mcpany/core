# Global Search Feature

## Overview
The **Global Search** feature (Cmd+K) provides a unified command palette for quick navigation, theme switching, and accessing resources across the MCP Any management console. It mimics the "Spotlight" or "Alfred" experience found in premium tools.

## Implementation
- **Component**: `GlobalSearch` (`ui/src/components/global-search.tsx`)
- **Primitive**: `ui/src/components/ui/command.tsx` (based on `cmdk` + `shadcn/ui`)
- **Integration**: Added to `RootLayout` in `ui/src/app/layout.tsx`.
- **Keyboard Shortcut**: `Cmd+K` (Mac) or `Ctrl+K` (Windows/Linux).

## Verification
- **Unit/Integration**: Verified via manual interaction and code review.
- **E2E Tests**: Automated Playwright tests (`ui/tests/e2e/global-search.spec.ts`) ensure the menu opens, filters, and navigates correctly.

## Screenshot
![Global Search](global_search.png)
