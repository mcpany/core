/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect } from 'react';

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
    async function loadUser() {
      const token = localStorage.getItem('mcp_auth_token');
      if (!token) {
        setUser(null);
        setLoading(false);
        return;
      }

      try {
        // Extract username from token (Basic Auth: base64(username:password))
        // NOTE: Decoding client side is safe here as we only use the username for lookup.
        // We do NOT use the password.
        const decoded = atob(token);
        const username = decoded.split(':')[0];

        const res = await fetch(`/api/v1/users/${username}`, {
          headers: {
            'Authorization': `Basic ${token}`,
            'X-API-Key': '' // Clear explicit API key if any, rely on Basic Auth
          }
        });

        if (res.ok) {
          const userData = await res.json();
          // Map backend User to UI User
          const role = (userData.roles && userData.roles.includes('admin')) ? 'admin' : 'viewer';
          setUser({
            id: userData.id,
            name: userData.id, // Fallback to ID as name since proto lacks name
            email: `${userData.id}@mcp-any.io`, // Mock email for now
            role: role,
            avatar: undefined // No avatar in backend
          });
        } else {
          // Token invalid or user not found
          logout();
        }
      } catch (error) {
        console.error("Failed to load user", error);
        logout();
      } finally {
        setLoading(false);
      }
    }

    loadUser();
  }, []);

  const logout = () => {
    localStorage.removeItem('mcp_auth_token');
    setUser(null);
    window.location.href = '/login';
  };

  return (
    <UserContext.Provider value={{ user, loading, logout }}>
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
