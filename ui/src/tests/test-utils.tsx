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

export function render(
  ui: React.ReactElement,
  options?: Omit<RenderOptions, 'wrapper'> & { wrapper?: React.ComponentType<{ children: React.ReactNode }> }
) {
  const { wrapper: Wrapper, ...rest } = options ?? {};
  const WrapperComponent = ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter>{Wrapper ? <Wrapper>{children}</Wrapper> : children}</MemoryRouter>
  );
  return originalRender(ui, { ...rest, wrapper: WrapperComponent });
}
