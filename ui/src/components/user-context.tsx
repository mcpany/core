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
  /** Logs in the user with credentials. */
  login: (username: string, password?: string) => Promise<void>;
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

  const fetchUser = async () => {
    try {
      // Check if we have a token first
      const token = localStorage.getItem('mcp_auth_token');
      if (!token) {
        setLoading(false);
        return;
      }

      const u = await apiClient.getCurrentUser();
      setUser({
        id: u.id || u.name, // Use name as ID if ID is empty (common in basic auth)
        name: u.displayName || u.id || u.name,
        email: u.email || '',
        role: (u.roles && u.roles.includes('admin')) ? 'admin' : 'viewer',
        avatar: u.avatarUrl
      });
    } catch (e) {
      console.warn("Failed to fetch user", e);
      setUser(null);
      // Optional: clear invalid token
      // localStorage.removeItem('mcp_auth_token');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUser();
  }, []);

  const login = async (username: string, password?: string) => {
    // Perform login via client which stores token
    await apiClient.login(username, password);
    // Refresh user state
    await fetchUser();
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_auth_token');
    // Optional: Redirect to login
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
