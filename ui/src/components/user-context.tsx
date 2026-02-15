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
}

/**
 * Context interface for user management.
 */
interface UserContextType {
  /** The current user, or null if not logged in. */
  user: User | null;
  /** Whether authentication status is loading. */
  loading: boolean;
  /** Logs in the user. Redirects to login page. */
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

  const checkAuth = async () => {
    try {
        const me = await apiClient.getMe();
        let role: UserRole = 'viewer';
        if (me.roles && me.roles.includes('admin')) {
            role = 'admin';
        } else if (me.roles && me.roles.includes('editor')) {
            role = 'editor';
        } else if (me.role) {
            role = me.role as UserRole;
        }

        setUser({
            id: me.id,
            name: me.name,
            email: me.email || '',
            avatar: me.avatar,
            role: role
        });
    } catch (e) {
        // Not authenticated or failed to fetch
        setUser(null);
    } finally {
        setLoading(false);
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  const login = (role: UserRole) => {
      // In a real app, we would redirect to OAuth or login page.
      // For now, redirect to /login
      window.location.href = '/login';
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_auth_token');
    // Also clear the legacy mock key just in case
    localStorage.removeItem('mcp_user_role');
    window.location.href = '/login';
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
