/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { WelcomeScreen } from '@/components/onboarding/welcome-screen';
import { vi, describe, it, expect } from 'vitest';

describe('WelcomeScreen', () => {
  it('renders correctly', () => {
    const onRegister = vi.fn();
    const onTemplate = vi.fn();
    render(<WelcomeScreen onRegister={onRegister} onTemplate={onTemplate} />);

    expect(screen.getByText('Welcome to MCP Any')).toBeInTheDocument();
    expect(screen.getByText('Quick Start')).toBeInTheDocument();
    expect(screen.getByText('Connect Manually')).toBeInTheDocument();
  });

  it('calls onTemplate when Quick Start is clicked', () => {
    const onRegister = vi.fn();
    const onTemplate = vi.fn();
    render(<WelcomeScreen onRegister={onRegister} onTemplate={onTemplate} />);

    fireEvent.click(screen.getByText('Browse Templates'));
    expect(onTemplate).toHaveBeenCalled();
  });

  it('calls onRegister when Connect Manually is clicked', () => {
    const onRegister = vi.fn();
    const onTemplate = vi.fn();
    render(<WelcomeScreen onRegister={onRegister} onTemplate={onTemplate} />);

    fireEvent.click(screen.getByText('Configure Service'));
    expect(onRegister).toHaveBeenCalled();
  });
});
