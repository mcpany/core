/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Re-export all of @testing-library/react plus a custom `render` that wraps
 * the component under test in a MemoryRouter.  Components migrated from
 * Next.js use react-router-dom hooks (Link, useNavigate, useLocation, …)
 * which require a Router context to be present.
 */
export * from '@testing-library/react';
import { render as originalRender, type RenderOptions } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';

 /**
 * Summary: A custom render function that wraps the component with necessary providers for testing.
 *
 * Params:
 *   - ui (ReactElement): The React component to render.
 *   - options (Omit<RenderOptions, 'wrapper'>): Optional configuration for testing-library render.
 *
 * Returns:
 *   - RenderResult: The result of testing-library's render function.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - Wraps the provided UI component within application context providers (e.g., ThemeProvider).
 */
export function render(
  ui: React.ReactElement,
  options?: Omit<RenderOptions, 'wrapper'> & { wrapper?: React.ComponentType<{ children: React.ReactNode }> }
) {
  const { wrapper: Wrapper, ...rest } = options ?? {};
/**
 * WrapperComponent component.
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns The rendered component.
 */
  const WrapperComponent = ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter>{Wrapper ? <Wrapper>{children}</Wrapper> : children}</MemoryRouter>
  );
  return originalRender(ui, { ...rest, wrapper: WrapperComponent });
}
