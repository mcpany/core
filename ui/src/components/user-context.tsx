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
  roles: string[]; // Backend uses 'roles' array
}

/**
 * Context interface for user management.
 */
interface UserContextType {
  /** The current user, or null if not logged in. */
  user: User | null;
  /** Whether authentication status is loading. */
  loading: boolean;
  /** Logs in the user (simulated for dev/test, real implementation would trigger OAuth/SSO). */
  login: (role: UserRole) => Promise<void>;
  /** Logs out the current user. */
  logout: () => void;
  /** Refreshes the user session. */
  refresh: () => Promise<void>;
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
      const u = await apiClient.getCurrentUser();
      if (u) {
        setUser({
            id: u.id,
            name: u.name || u.id,
            email: u.email || '',
            roles: u.roles || [],
            avatar: u.avatar
        });
      } else {
        setUser(null);
      }
    } catch (e) {
      console.error("Failed to fetch user:", e);
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUser();
  }, []);

  const login = async (role: UserRole) => {
    // In a real app, this would trigger an OAuth flow or redirect to login page.
    // For development/testing with Basic Auth or API Key, the client is pre-configured via headers/localStorage.
    // We simulate a login by refreshing the user state from the backend.
    console.log(`[UserContext] Simulating login for role: ${role}`);
    await fetchUser();
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_auth_token');
    localStorage.removeItem('mcp_user_role');
    // Force reload to clear client state
    window.location.reload();
  };

  return (
    <UserContext.Provider value={{ user, loading, login, logout, refresh: fetchUser }}>
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
