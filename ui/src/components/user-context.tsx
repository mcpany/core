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
  /** Logs in the user. */
  login: (username: string, password: string) => Promise<void>;
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
          const currentUser = await apiClient.getCurrentUser();
          if (currentUser) {
              setUser({
                  id: currentUser.id || '1',
                  name: currentUser.name || 'User',
                  email: currentUser.email || 'user@example.com',
                  role: (currentUser.role as UserRole) || 'viewer',
                  avatar: currentUser.avatar
              });
          } else {
              setUser(null);
          }
      } catch (e) {
          console.error("Auth check failed", e);
          setUser(null);
      } finally {
          setLoading(false);
      }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  const login = async (username: string, password: string) => {
    try {
        const res = await apiClient.login(username, password);
        if (res.token) {
            localStorage.setItem('mcp_auth_token', res.token);
            await checkAuth();
        }
    } catch (e) {
        console.error("Login failed", e);
        throw e;
    }
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_auth_token');
    localStorage.removeItem('mcp_user_role');
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
