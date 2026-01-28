import { render, screen, act } from '@testing-library/react';
import { DashboardGrid } from '@/components/dashboard/dashboard-grid';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import React from 'react';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value.toString();
    }),
    clear: vi.fn(() => {
      store = {};
    }),
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Mock dependencies
vi.mock('@hello-pangea/dnd', () => ({
  DragDropContext: ({ children }: any) => <div>{children}</div>,
  Droppable: ({ children }: any) => children({ droppableProps: {}, innerRef: null, placeholder: null }),
  Draggable: ({ children }: any) => children({ draggableProps: {}, dragHandleProps: {}, innerRef: null }, { isDragging: false }),
}));

vi.mock('@/components/dashboard/widget-registry', () => ({
  WIDGET_DEFINITIONS: [
    { type: 'test-widget', title: 'Test Widget', defaultSize: 'third', component: () => <div>Test Widget</div> }
  ],
  getWidgetDefinition: (type: string) => ({ type, title: 'Test Widget', defaultSize: 'third', component: () => <div>Test Widget</div> }),
}));

vi.mock('@/components/dashboard/add-widget-sheet', () => ({
  AddWidgetSheet: ({ onAdd }: any) => (
    <button onClick={() => onAdd('test-widget')} data-testid="add-widget">
      Add Widget
    </button>
  ),
}));

describe('DashboardGrid', () => {
  beforeEach(() => {
    localStorageMock.clear();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it('debounces localStorage writes', async () => {
    render(<DashboardGrid />);

    // Initial load should trigger effect, but let's wait it out
    act(() => {
      vi.runAllTimers();
    });
    localStorageMock.setItem.mockClear();

    const addButton = screen.getByTestId('add-widget');

    // Trigger update 1
    act(() => {
      addButton.click();
    });

    // localStorage should NOT be called yet
    expect(localStorageMock.setItem).not.toHaveBeenCalled();

    // Trigger update 2 immediately
    act(() => {
      addButton.click();
    });

    expect(localStorageMock.setItem).not.toHaveBeenCalled();

    // Fast forward time
    act(() => {
      vi.advanceTimersByTime(500);
    });

    // Now it should be called ONCE (for the latest state)
    expect(localStorageMock.setItem).toHaveBeenCalledTimes(1);

    // Check if the content is correct (roughly)
    const stored = JSON.parse(localStorageMock.setItem.mock.calls[0][1]);
    expect(stored.length).toBeGreaterThan(0);
  });
});
