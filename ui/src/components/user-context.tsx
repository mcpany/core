/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect } from 'react';
import { apiClient } from "@/lib/client";

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
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns {JSX.Element} The provider component.
 */
export function UserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // Helper to sync user state with backend user
  const syncUser = async (role: UserRole) => {
      try {
          const storedRole = role || localStorage.getItem('mcp_user_role') as UserRole || 'admin';
          const userId = "1"; // Fixed ID for single-user mode/dev

          // Try to fetch existing user
          let backendUser: any = null;
          try {
              // We use listUsers because getUser might not be exposed or we want to search
              // Ideally we would use getUser(id)
              // Checking apiClient capability... it has listUsers, createUser, updateUser.
              // It does NOT have getUser(id) explicitly in the client.ts I read earlier,
              // but I can add it or just use listUsers for now if list is small.
              // Actually, I can use a direct fetch if needed, but let's see.
              // Let's assume listUsers is fine for now.
              const usersResponse = await apiClient.listUsers();
              const users = Array.isArray(usersResponse) ? usersResponse : (usersResponse.users || []);
              backendUser = users.find((u: any) => u.id === userId);
          } catch (e) {
              console.error("Failed to list users", e);
          }

          if (!backendUser) {
              // Create default user if not exists
              const newUser = {
                  id: userId,
                  roles: [storedRole], // Map role to roles
                  // preferences: {} // Default empty
              };
              try {
                // apiClient.createUser expects { user: ... } wrapper in some versions,
                // but let's match what client.ts says: body: JSON.stringify({ user })
                const response = await apiClient.createUser(newUser);
                backendUser = response.user || response;
              } catch (e) {
                  console.error("Failed to create user", e);
                  // Fallback to local state if backend fails
                  backendUser = { id: userId, roles: [storedRole] };
              }
          }

          setUser({
              id: userId,
              name: 'Admin User', // Hardcoded display details for now
              email: 'admin@mcp-any.io',
              role: storedRole,
              avatar: '/avatars/admin.png',
              preferences: backendUser.preferences || {}
          });

      } catch (err) {
          console.error("User sync failed", err);
      } finally {
          setLoading(false);
      }
  };

  useEffect(() => {
    syncUser('admin');
  }, []);

  const login = (role: UserRole) => {
    localStorage.setItem('mcp_user_role', role);
    syncUser(role);
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('mcp_user_role');
  };

  const updatePreferences = async (newPrefs: Record<string, string>) => {
      if (!user) return;

      const updatedPrefs = { ...user.preferences, ...newPrefs };
      const updatedUser = { ...user, preferences: updatedPrefs };
      setUser(updatedUser);

      try {
          // Map back to backend structure
          // The backend User proto has `roles`, `id`, `preferences`.
          // apiClient.updateUser expects { id: ... } at least.
          await apiClient.updateUser({
              id: user.id,
              preferences: updatedPrefs,
              // We should probably preserve roles if we knew them from backend,
              // but for now we trust what we have.
              // Be careful not to overwrite other fields if UpdateUser is a full replace.
              // The server implementation of UpdateUser:
              // if err := s.storage.UpdateUser(ctx, req.GetUser()); err != nil
              // It replaces the user config_json.
              // So we MUST send all fields we want to keep.
              roles: [user.role]
          });
      } catch (e) {
          console.error("Failed to save preferences", e);
          // Revert or show toast? For now just log.
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
