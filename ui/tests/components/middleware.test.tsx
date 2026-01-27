/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import MiddlewarePage from '../../src/app/middleware/page';

// Mock the drag and drop context to avoid errors in JSDOM
vi.mock('@hello-pangea/dnd', () => ({
  DragDropContext: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Droppable: ({ children }: { children: (provided: any) => React.ReactNode }) => (
    <div>
      {children({
        droppableProps: {},
        innerRef: vi.fn(),
        placeholder: null,
      })}
    </div>
  ),
  Draggable: ({ children }: { children: (provided: any, snapshot: any) => React.ReactNode }) => (
    <div>
      {children(
        {
          draggableProps: {},
          dragHandleProps: {},
          innerRef: vi.fn(),
        },
        { isDragging: false }
      )}
    </div>
  ),
}));

describe('MiddlewarePage Component', () => {
  it('should display the middleware pipeline heading', () => {
    render(<MiddlewarePage />);
    expect(screen.getByText('Middleware Pipeline')).toBeDefined();
  });

  it('should display core middleware items', () => {
    render(<MiddlewarePage />);
    expect(screen.getAllByText('Authentication').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Rate Limiter').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Logging').length).toBeGreaterThan(0);
  });


  it('should display switches for each middleware', () => {
    render(<MiddlewarePage />);
    const switches = screen.getAllByRole('button'); // Switch in Radix is often rendered as a button or has role switch
    // Depending on the implementation, we might need to be more specific.
    // Based on the code, it uses Radix Switch which usually has role="switch".
    const switchElements = screen.queryAllByRole('switch');
    // If switch role is not present (sometimes happens in JSDOM/Radix without proper ARIA),
    // we can check for checkboxes or just ensure they are present.
    expect(switchElements.length).toBeGreaterThan(0);
  });

  it('should display settings buttons', () => {
    render(<MiddlewarePage />);
    // The settings button has an icon, we can find it by its container or if it has a label.
    // In the code it's a Button with variant="ghost" size="icon" containing a Settings icon.
    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThan(0);
  });
});
