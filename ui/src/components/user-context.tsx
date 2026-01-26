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
  /** User preferences (e.g., dashboard layout). */
  preferences?: Record<string, string>;
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
  /** Updates user preferences. */
  updatePreferences: (prefs: Record<string, string>) => Promise<void>;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

/**
 * Provider component for user authentication context.
 *
 * @param { children - The { children.
 */
export function UserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // Default mock user
  const mockUser: User = {
    id: '1',
    name: 'Admin User',
    email: 'admin@mcp-any.io',
    role: 'admin',
    avatar: '/avatars/admin.png',
    preferences: {}
  };

  useEffect(() => {
    const initUser = async () => {
        // Mock initial user for now - default to Admin for development
        // In real app, check session/cookie
        const storedRole = localStorage.getItem('mcp_user_role') as UserRole || 'admin';

        // Try to fetch real user from backend if possible to get preferences
        // Assuming we are "logged in" as user "1" in this dev/mock setup
        let prefs: Record<string, string> = {};
        try {
            // We use '1' as the ID for the mock admin user
            const backendUserResp = await apiClient.getUser('1'); // This requires a GetUser by ID endpoint or similar
            if (backendUserResp && backendUserResp.user && backendUserResp.user.preferences) {
                prefs = backendUserResp.user.preferences;
            }
        } catch (e) {
            // Ignore backend fetch error in dev/mock mode
            console.warn("Could not fetch user preferences from backend (using mock defaults):", e);
            // Fallback: check localStorage for any locally saved prefs (optional migration)
             const savedLayout = localStorage.getItem("dashboard-layout");
             if (savedLayout) {
                 prefs['dashboard-layout'] = savedLayout;
             }
        }

        setUser({
            ...mockUser,
            role: storedRole,
            preferences: prefs
        });
        setLoading(false);
    };

    initUser();
  }, []);

  const login = (role: UserRole) => {
    const newUser: User = {
        id: '1',
        name: role === 'admin' ? 'Admin User' : 'Regular User',
        email: role === 'admin' ? 'admin@mcp-any.io' : 'user@mcp-any.io',
        role: role,
        avatar: role === 'admin' ? '/avatars/admin.png' : undefined,
        preferences: {}
    };
    setUser(newUser);
    localStorage.setItem('mcp_user_role', role);
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_user_role');
  };

  const updatePreferences = async (newPrefs: Record<string, string>) => {
    if (!user) return;

    const updatedPrefs = { ...(user.preferences || {}), ...newPrefs };

    // Optimistic update
    setUser({
        ...user,
        preferences: updatedPrefs
    });

    // Sync to backend
    try {
        // We first need to get the "real" user object to update it properly (full replace)
        // or we construct a best-effort one.
        // Since backend UpdateUser does a full replace, we should fetch-modify-save or hope partial works?
        // Actually, for this specific task, we'll try to fetch first.
        let backendUser: any;
        try {
            const resp = await apiClient.getUser(user.id);
            backendUser = resp.user;
        } catch (e) {
            // If user doesn't exist in backend, we might need to create it?
            // For now, assume it might fail if using mock.
            // If it fails, we can try to create it?
            // Let's just log and return if we can't sync.
            console.warn("Backend user not found, skipping sync:", e);
            return;
        }

        if (backendUser) {
            backendUser.preferences = updatedPrefs;
            await apiClient.updateUser(backendUser);
        }
    } catch (e) {
        console.error("Failed to sync preferences to backend:", e);
    }
  };

  return (
    <UserContext.Provider value={{ user, loading, login, logout, updatePreferences }}>
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
