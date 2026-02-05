/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect, useMemo, useCallback } from 'react';

/**
 * Defines the role of a user in the system.
 */
export type UserRole = 'admin' | 'editor' | 'viewer';

/**
 * Represents a user of the application.
 */
export interface User {
  /** Unique user ID. */
  id: string;
  /** Display name. */
  name: string;
  /** Email address. */
  email: string;
  /** URL to avatar image. */
  avatar?: string;
  /** The user's role. */
  role: UserRole;
}

/**
 * Context interface for user management.
 */
interface UserContextType {
  /** The current user, or null if not logged in. */
  user: User | null;
  /** Whether authentication status is loading. */
  loading: boolean;
  /** Logs in the user with the specified role (mock). */
  login: (role: UserRole) => void;
  /** Logs out the current user. */
  logout: () => void;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

/**
 * Provider component for user authentication context.
 *
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns {JSX.Element} The provider component.
 */
export function UserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Mock initial user for now - default to Admin for development
    // In real app, check session/cookie
    const storedRole = localStorage.getItem('mcp_user_role') as UserRole || 'admin';
    setUser({
      id: '1',
      name: 'Admin User',
      email: 'admin@mcp-any.io',
      role: storedRole, // Default to admin for dev
      avatar: '/avatars/admin.png'
    });
    setLoading(false);
  }, []);

  const login = useCallback((role: UserRole) => {
    const newUser = {
        id: '1',
        name: role === 'admin' ? 'Admin User' : 'Regular User',
        email: role === 'admin' ? 'admin@mcp-any.io' : 'user@mcp-any.io',
        role: role,
        avatar: role === 'admin' ? '/avatars/admin.png' : undefined
    };
    setUser(newUser);
    localStorage.setItem('mcp_user_role', role);
  }, []);

  const logout = useCallback(() => {
    setUser(null);
    localStorage.removeItem('mcp_user_role');
  }, []);

  // ⚡ BOLT: Memoized context value to prevent unnecessary re-renders.
  // Randomized Selection from Top 5 High-Impact Targets
  const value = useMemo(() => ({
    user,
    loading,
    login,
    logout
  }), [user, loading, login, logout]);

  return (
    <UserContext.Provider value={value}>
      {children}
    </UserContext.Provider>
  );
}

/**
 * Hook to access the user context.
 * @returns The user context.
 * @throws Error if used outside of a UserProvider.
 */
export function useUser() {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
}
