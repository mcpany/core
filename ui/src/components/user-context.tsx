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
    const fetchUser = async () => {
      try {
        const userData = await apiClient.getCurrentUser();
        // Map backend user to UI user
        // Note: backend roles is string[], UI expects UserRole
        // We take the first role or default to viewer
        let role: UserRole = 'viewer';
        if (userData.roles && userData.roles.length > 0) {
            if (userData.roles.includes('admin')) role = 'admin';
            else if (userData.roles.includes('editor')) role = 'editor';
        }

        setUser({
          id: userData.id,
          name: userData.name || userData.username || 'User',
          email: userData.email,
          role: role,
          avatar: userData.avatar,
        });
      } catch (e) {
        console.error("Failed to fetch current user", e);
        setUser(null);
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, []);

  const login = (role: UserRole) => {
    // For now, we redirect to login page or just reload to trigger auth check?
    // Since we removed client-side mock login, we should rely on backend auth.
    // If this is a dev helper, we might keep it but warn.
    // Ideally we redirect to /auth/login
    window.location.href = '/auth/login';
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_auth_token'); // Clear token if used
    // Redirect to logout or reload
    window.location.reload();
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
