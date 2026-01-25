/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AppSidebar } from '../components/app-sidebar';
import { SidebarProvider } from '../components/ui/sidebar';
import { UserProvider } from '../components/user-context';
import { KeyboardShortcutsProvider } from '../contexts/keyboard-shortcuts-context';

// Mock ResizeObserver
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserver;

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

describe('AppSidebar Navigation', () => {
  it('renders all navigation groups', () => {
    render(
      <UserProvider>
        <KeyboardShortcutsProvider>
          <SidebarProvider>
            <AppSidebar />
          </SidebarProvider>
        </KeyboardShortcutsProvider>
      </UserProvider>
    );

    expect(screen.getByText('Platform')).toBeDefined();
    expect(screen.getByText('Development')).toBeDefined();
    expect(screen.getByText('Configuration')).toBeDefined();
  });

  it('renders key navigation links', () => {
    render(
      <UserProvider>
        <KeyboardShortcutsProvider>
          <SidebarProvider>
             <AppSidebar />
          </SidebarProvider>
        </KeyboardShortcutsProvider>
      </UserProvider>
    );

    const links = [
      'Dashboard',
      'Network Graph',
      'Diagnostics',
      'Live Logs',
      'Playground',
      'Tools',
      'Upstream Services',
      'Diagnostics',
      'Secrets Vault'
    ];

    links.forEach(linkText => {
      expect(screen.getByText(linkText)).toBeDefined();
    });
  });
});
