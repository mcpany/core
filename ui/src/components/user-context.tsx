/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect } from 'react';
import { apiClient } from '@/lib/client';

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
  /** The user's roles list */
  roles?: string[];
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
    async function fetchUser() {
      try {
        const userData = await apiClient.getCurrentUser();
        // Map backend user to UI user
        // Backend user: { id, name, email, roles: [] }
        // UI User: { ..., role: 'admin' | ... }
        // Determine primary role
        let role: UserRole = 'viewer';
        if (userData.roles?.includes('admin')) role = 'admin';
        else if (userData.roles?.includes('editor')) role = 'editor';

        setUser({
          id: userData.id,
          name: userData.name || userData.id, // Fallback if name empty
          email: userData.email || '',
          avatar: userData.avatar,
          role: role,
          roles: userData.roles
        });
      } catch (e) {
        console.warn("Failed to fetch current user (likely not logged in):", e);
        setUser(null);
      } finally {
        setLoading(false);
      }
    }
    fetchUser();
  }, []);

  const login = (role: UserRole) => {
    console.warn("Client-side login/role-switching is disabled. Use backend authentication.");
    // No-op or trigger real login flow
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_auth_token');
    // Optional: Redirect to login page
    // window.location.href = '/login';
  };

  return (
    <UserContext.Provider value={{ user, loading, login, logout }}>
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
