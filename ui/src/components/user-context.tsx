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
  /** Logs in the user with the specified role. */
  login: (role: UserRole) => void;
  /** Logs out the current user. */
  logout: () => void;
  /** Error message if connection failed. */
  error?: string;
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
  const [error, setError] = useState<string | undefined>(undefined);

  useEffect(() => {
    const initUser = async () => {
        try {
            // Verify connection to backend and fetch users
            // We use listUsers to check if we can talk to the backend
            // In a real auth flow, this would be getCurrentUser()
            const users = await apiClient.listUsers();

            const storedRole = localStorage.getItem('mcp_user_role') as UserRole || 'admin';

            if (users && users.length > 0) {
                // Use the first user found in backend, or fallback to a default identity
                // but confirmed by successful API call.
                // Since the backend users might be service accounts, we might still want to
                // simulate a "session" user for the UI, but strictly only if backend is reachable.
                const backendUser = users[0];
                setUser({
                    id: backendUser.id,
                    name: backendUser.id, // Use ID as name for now
                    email: `${backendUser.id}@mcp-any.io`,
                    role: storedRole,
                    avatar: '/avatars/admin.png'
                });
            } else {
                // Backend reachable but no users?
                // Create a default admin context but verified against backend connectivity
                setUser({
                    id: 'admin',
                    name: 'Admin User',
                    email: 'admin@mcp-any.io',
                    role: storedRole,
                    avatar: '/avatars/admin.png'
                });
            }
        } catch (e: any) {
            console.error("Failed to connect to backend:", e);
            setError(e.message || "Failed to connect to backend");
            // Do not set user if backend is unreachable, enforcement of Real Data Policy
            setUser(null);
        } finally {
            setLoading(false);
        }
    };

    initUser();
  }, []);

  const login = (role: UserRole) => {
    // Logic to switch role (simulation for now as we don't have full auth flow)
    const newUser = {
        id: '1',
        name: role === 'admin' ? 'Admin User' : 'Regular User',
        email: role === 'admin' ? 'admin@mcp-any.io' : 'user@mcp-any.io',
        role: role,
        avatar: role === 'admin' ? '/avatars/admin.png' : undefined
    };
    setUser(newUser);
    localStorage.setItem('mcp_user_role', role);
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_user_role');
  };

  return (
    <UserContext.Provider value={{ user, loading, login, logout, error }}>
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
