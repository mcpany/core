/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { WidgetSkeleton } from './widget-skeleton';

describe('WidgetSkeleton Component', () => {
  it('renders correctly with default classes', () => {
    const { container } = render(<WidgetSkeleton data-testid="skeleton" />);
    const skeletonElement = container.firstChild as HTMLElement;

    expect(skeletonElement).toBeInTheDocument();
    expect(skeletonElement).toHaveClass('animate-pulse');
    expect(skeletonElement).toHaveClass('rounded-md');
    expect(skeletonElement).toHaveClass('bg-muted');
  });

  it('merges custom classNames correctly', () => {
    const { container } = render(<WidgetSkeleton className="h-4 w-24 custom-class" />);
    const skeletonElement = container.firstChild as HTMLElement;

    expect(skeletonElement).toHaveClass('animate-pulse');
    expect(skeletonElement).toHaveClass('h-4');
    expect(skeletonElement).toHaveClass('w-24');
    expect(skeletonElement).toHaveClass('custom-class');
  });

  it('passes other props directly to the underlying div', () => {
    const { getByTestId } = render(<WidgetSkeleton data-testid="test-skeleton" id="test-id" />);
    const skeletonElement = getByTestId('test-skeleton');

    expect(skeletonElement).toHaveAttribute('id', 'test-id');
  });
});
