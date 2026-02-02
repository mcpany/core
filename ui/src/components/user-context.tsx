/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect } from 'react';
import { apiClient } from '@/lib/client';
import { useToast } from "@/hooks/use-toast";

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
  /** User preferences. */
  preferences: Record<string, string>;
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

const DEFAULT_USER_ID = '1';

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
  const { toast } = useToast();

  const fetchUser = async () => {
    try {
      // Try to find the default user
      const response = await apiClient.listUsers();
      const users = response.users || [];
      const existingUser = users.find((u: any) => u.id === DEFAULT_USER_ID);

      if (existingUser) {
        // Map backend user to UI user
        setUser({
          id: existingUser.id,
          name: 'Admin User', // Backend User doesn't have name field yet in proto/config/v1/user.proto, maybe strictly ID?
          // Actually configv1.User has only id, authentication, profile_ids, roles.
          // We can infer name/email or just mock it for UI display.
          email: 'admin@mcp-any.io',
          role: (existingUser.roles && existingUser.roles.includes('admin')) ? 'admin' : 'viewer',
          avatar: '/avatars/admin.png',
          preferences: existingUser.preferences || {}
        });
      } else {
        // Create the default user if it doesn't exist
        const newUser = {
          id: DEFAULT_USER_ID,
          roles: ['admin'],
          preferences: {}
        };
        await apiClient.createUser(newUser);
        setUser({
          id: DEFAULT_USER_ID,
          name: 'Admin User',
          email: 'admin@mcp-any.io',
          role: 'admin',
          avatar: '/avatars/admin.png',
          preferences: {}
        });
      }
    } catch (error) {
      console.error("Failed to fetch/create user:", error);
      toast({
        title: "User Error",
        description: "Failed to load user profile. Some features may not work.",
        variant: "destructive"
      });
      // Fallback to local mock so UI doesn't break entirely
      setUser({
        id: DEFAULT_USER_ID,
        name: 'Admin User',
        email: 'admin@mcp-any.io',
        role: 'admin',
        avatar: '/avatars/admin.png',
        preferences: {}
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUser();
  }, []);

  const login = (role: UserRole) => {
    // In a real app, this would redirect to login or exchange credentials.
    // For now, we just reload the user fetching logic which ensures a user exists.
    // Ideally we would update the user's role in backend.
    setLoading(true);
    // Simulate updating role
    if (user) {
        const updatedRoles = [role];
        apiClient.updateUser({
            id: user.id,
            roles: updatedRoles,
            preferences: user.preferences
        }).then(() => fetchUser())
          .catch(() => {
              // Fallback
              const newUser = {
                ...user,
                role: role
             };
             setUser(newUser);
             setLoading(false);
          });
    } else {
        fetchUser();
    }
  };

  const logout = () => {
    setUser(null);
    // In real app, clear tokens.
  };

  const updatePreferences = async (prefs: Record<string, string>) => {
    if (!user) return;
    try {
        const newPreferences = { ...user.preferences, ...prefs };
        // Optimistic update
        setUser({ ...user, preferences: newPreferences });

        // Backend update
        // We need to send the FULL user object usually for PUT, or at least required fields.
        // apiClient.updateUser sends what we give it.
        // Backend UpdateUser usually requires ID.
        // We should send existing fields to avoid overwriting them with nulls if backend is not PATCH.
        // Our backend implementation uses protojson.Unmarshal which handles partials if options set?
        // But the SQL UPDATE replaces config_json. So we must read-modify-write or send full state.
        // Since we have `user` state which should match backend (mostly), we can construct it.
        // BUT `user` state here has extra UI fields (name, email, avatar) not in backend.
        // We should construct the backend object.

        const backendUser = {
            id: user.id,
            roles: user.role === 'admin' ? ['admin'] : ['viewer'], // Simplified mapping
            preferences: newPreferences
            // authentication, profile_ids preserved?
            // If we don't send them, and backend does full replace, we lose them.
            // The `UpdateUser` implementation in `server/pkg/admin/server.go` calls `s.storage.UpdateUser`.
            // The `store.go` implementation does:
            // UPDATE users SET config_json = $2 ...
            // So it IS a full replace.
            // We must ensure we don't lose data.
            // Ideally we fetch fresh user, update prefs, then save.
            // Or we trust our local `user.preferences` is up to date (it should be).
            // But what about `authentication`? We don't store it in UI state.
            // We should fetch first.
        };

        // Safer approach: Fetch, patch, save.
        const response = await apiClient.listUsers(); // Optimisation: get specific user if possible
        const users = response.users || [];
        const existingBackendUser = users.find((u: any) => u.id === user.id);

        if (existingBackendUser) {
             const updatedUser = {
                 ...existingBackendUser,
                 preferences: newPreferences
             };
             await apiClient.updateUser(updatedUser);
        } else {
            // Should not happen if user is logged in
            await apiClient.updateUser(backendUser);
        }

    } catch (e) {
        console.error("Failed to update preferences", e);
        toast({
            title: "Save Failed",
            description: "Failed to save preferences.",
            variant: "destructive"
        });
        // Revert optimistic update?
        // fetchUser();
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
