/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '../tests/test-utils';
import { AppSidebar } from '../components/app-sidebar';

// Mock the entire sidebar module so SidebarProvider/Sidebar/useSidebar avoid context issues
vi.mock('@/components/ui/sidebar', () => ({
  SidebarProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Sidebar: ({ children }: { children: React.ReactNode }) => <nav>{children}</nav>,
  SidebarContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarGroup: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarGroupContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarGroupLabel: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarMenu: ({ children }: { children: React.ReactNode }) => <ul>{children}</ul>,
  SidebarMenuItem: ({ children }: { children: React.ReactNode }) => <li>{children}</li>,
  SidebarMenuButton: ({ children, asChild, ...props }: { children: React.ReactNode; asChild?: boolean; [key: string]: unknown }) => <button {...props as object}>{children}</button>,
  SidebarFooter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarRail: () => <div />,
  SidebarSeparator: () => <hr />,
  SidebarTrigger: () => <button />,
  SidebarMenuSub: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarMenuSubItem: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarMenuSubButton: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SidebarMenuBadge: ({ children }: { children: React.ReactNode }) => <span>{children}</span>,
  useSidebar: () => ({ state: 'expanded', open: true, setOpen: vi.fn(), isMobile: false, openMobile: false, setOpenMobile: vi.fn(), toggleSidebar: vi.fn() }),
}));

// Mock user context (alias path used by AppSidebar)
vi.mock('@/components/user-context', () => ({
  UserProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useUser: () => ({
    user: { id: 'admin-user', name: 'Admin', email: 'admin@test.com', roles: ['admin'] },
    loading: false,
    login: vi.fn(),
    logout: vi.fn(),
    refresh: vi.fn(),
  }),
}));

// Mock keyboard shortcuts context
vi.mock('@/contexts/keyboard-shortcuts-context', () => ({
  KeyboardShortcutsProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useKeyboardShortcuts: () => ({
    shortcuts: {},
    overrides: {},
    register: vi.fn(),
    unregister: vi.fn(),
    updateOverride: vi.fn(),
    resetOverride: vi.fn(),
    getKeys: vi.fn().mockReturnValue([]),
  }),
  useShortcut: vi.fn(),
}));

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
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

describe('AppSidebar Navigation', () => {
  it('renders all navigation groups', () => {
    render(<AppSidebar />);

    expect(screen.getByText('Platform')).toBeDefined();
    expect(screen.getByText('Development')).toBeDefined();
    expect(screen.getByText('Configuration')).toBeDefined();
  });

  it('renders key navigation links', () => {
    render(<AppSidebar />);

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

    links.forEach((linkText: string) => {
      expect(screen.getByText(linkText)).toBeDefined();
    });
  });
});
