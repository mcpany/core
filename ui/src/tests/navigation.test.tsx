/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AppSidebar } from '../components/app-sidebar';
import { SidebarProvider } from '../components/ui/sidebar';
import { KeyboardShortcutsProvider } from '../contexts/keyboard-shortcuts-context';
import { UserProvider } from '../components/user-context';
import { FavoritesProvider } from '../contexts/favorites-context';

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
        <FavoritesProvider>
          <KeyboardShortcutsProvider>
            <SidebarProvider>
              <AppSidebar />
            </SidebarProvider>
          </KeyboardShortcutsProvider>
        </FavoritesProvider>
      </UserProvider>
    );

    expect(screen.getByText('Platform')).toBeDefined();
    expect(screen.getByText('Development')).toBeDefined();
    expect(screen.getByText('Configuration')).toBeDefined();
  });

  it('renders key navigation links', () => {
    render(
      <UserProvider>
        <FavoritesProvider>
          <KeyboardShortcutsProvider>
            <SidebarProvider>
               <AppSidebar />
            </SidebarProvider>
          </KeyboardShortcutsProvider>
        </FavoritesProvider>
      </UserProvider>
    );

    const links = [
      'Dashboard',
      'Network Graph',
      'Live Logs',
      'Playground',
      'Tools',
      'Upstream Services',
      'Secrets Vault'
    ];

    links.forEach(linkText => {
      expect(screen.getByText(linkText)).toBeDefined();
    });
  });
});
