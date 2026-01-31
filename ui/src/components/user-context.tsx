/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { apiClient } from "@/lib/client";
import { User as ProtoUser } from "@proto/config/v1/user";

/**
 * Defines the role of a user in the system.
 */
export type UserRole = 'admin' | 'editor' | 'viewer';

/**
 * Represents a user of the application.
 */
export interface User extends ProtoUser {
  /** The user's primary role (derived). */
  role: UserRole;
  /** URL to avatar image (optional). */
  avatar?: string;
  /** Display name (derived). */
  name: string;
  /** Email address (derived). */
  email: string;
}

/**
 * Context interface for user management.
 */
interface UserContextType {
  /** The current user, or null if not logged in. */
  user: User | null;
  /** Whether authentication status is loading. */
  loading: boolean;
  /** Updates the current user preferences. */
  updatePreferences: (prefs: any) => Promise<void>;
  /** Refreshes the user data. */
  refresh: () => Promise<void>;
  /** Logs out the current user. */
  logout: () => void;
  /** Legacy login method (redirects). */
  login: (role: any) => void;
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

  const fetchUser = useCallback(async () => {
    setLoading(true);
    try {
        // Fetch real user from backend
        const protoUser = await apiClient.getCurrentUser() as ProtoUser;

        // Derive role from roles list
        let role: UserRole = 'viewer';
        if (protoUser.roles?.includes('admin')) {
            role = 'admin';
        } else if (protoUser.roles?.includes('editor')) {
            role = 'editor';
        }

        // Add avatar if missing (using Gravatar or default based on ID)
        // For now, keep the simple placeholder logic
        const avatar = `/avatars/${role === 'admin' ? 'admin' : 'user'}.png`;

        // Derive name and email
        let name = protoUser.id;
        let email = "";

        if (protoUser.authentication?.basicAuth?.username) {
            name = protoUser.authentication.basicAuth.username;
        } else if (protoUser.authentication?.oidc?.email) {
            email = protoUser.authentication.oidc.email;
            name = email.split('@')[0];
        }

        if (!email && name.includes('@')) {
            email = name;
        }

        setUser({
            ...protoUser,
            role,
            avatar,
            name,
            email,
            // Ensure preferencesJson is set
            preferencesJson: protoUser.preferencesJson || ""
        });
    } catch (e) {
        console.error("Failed to fetch user:", e);
        setUser(null);
    } finally {
        setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  const updatePreferences = async (prefs: any) => {
      if (!user) return;
      try {
          const json = JSON.stringify(prefs);
          const updated = { ...user, preferencesJson: json };

          // Optimistic update
          setUser(updated);

          // Persist to backend
          // We need to map UI User back to Proto User structure if needed,
          // but apiClient.updateUser handles 'any' currently.
          await apiClient.updateUser(updated);
      } catch (e) {
          console.error("Failed to update preferences:", e);
          // Revert on error?
          fetchUser();
      }
  };

  const login = (role: any) => {
      // In real auth, this would redirect to /auth/login or initiate OIDC
      window.location.href = '/api/v1/auth/login';
  };

  const logout = () => {
    setUser(null);
    // Clear token if any
    localStorage.removeItem('mcp_auth_token');
    // Refresh to clear state/redirect
    window.location.reload();
  };

  return (
    <UserContext.Provider value={{ user, loading, updatePreferences, refresh: fetchUser, login, logout }}>
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
