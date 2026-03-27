import React from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { SmartTable } from './smart-table';

describe('SmartTable Resizing', () => {
    const testData = [
        { id: 1, name: 'Alice', role: 'Engineer' },
        { id: 2, name: 'Bob', role: 'Designer' }
    ];

    beforeEach(() => {
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('renders with default columns without explicit widths', () => {
        render(<SmartTable data={testData} />);

        const nameHeader = screen.getByText('name').closest('th');
        expect(nameHeader).toBeInTheDocument();
        expect(nameHeader?.style.width).toBe('auto');
    });

    it('allows resizing columns via mouse drag', () => {
        render(<SmartTable data={testData} />);

        const nameHeader = screen.getByText('name').closest('th');
        // Find the resizer handle
        const resizer = nameHeader?.querySelector('.cursor-col-resize');
        expect(resizer).toBeInTheDocument();

        // Start dragging
        act(() => {
            fireEvent.mouseDown(resizer!, { clientX: 100 });
        });

        // Move mouse
        act(() => {
            const moveEvent = new MouseEvent('mousemove', { bubbles: true, cancelable: true });
            Object.defineProperty(moveEvent, 'clientX', { value: 150 });
            document.dispatchEvent(moveEvent);
        });

        // Stop dragging
        act(() => {
            const upEvent = new MouseEvent('mouseup', { bubbles: true, cancelable: true });
            document.dispatchEvent(upEvent);
        });

        // Default width was estimated as 150, diff is 50 -> 200px
        expect(nameHeader?.style.width).toBe('200px');
    });

    it('respects minimum column width', () => {
        render(<SmartTable data={testData} />);

        const roleHeader = screen.getByText('role').closest('th');
        const resizer = roleHeader?.querySelector('.cursor-col-resize');

        // Start dragging
        act(() => {
            fireEvent.mouseDown(resizer!, { clientX: 300 });
        });

        // Move mouse massively left
        act(() => {
            const moveEvent = new MouseEvent('mousemove', { bubbles: true, cancelable: true });
            Object.defineProperty(moveEvent, 'clientX', { value: 50 });
            document.dispatchEvent(moveEvent);
        });

        // Stop dragging
        act(() => {
            const upEvent = new MouseEvent('mouseup', { bubbles: true, cancelable: true });
            document.dispatchEvent(upEvent);
        });

        // Min width is 60px
        expect(roleHeader?.style.width).toBe('60px');
    });
});