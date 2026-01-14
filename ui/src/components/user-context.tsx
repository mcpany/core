/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect } from 'react';

export type UserRole = 'admin' | 'editor' | 'viewer';

export interface User {
  id: string;
  name: string;
  email: string;
  avatar?: string;
  role: UserRole;
}

interface UserContextType {
  user: User | null;
  loading: boolean;
  login: (role: UserRole) => void;
  logout: () => void;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export function UserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Mock initial user for now - default to Admin for development
    // In real app, check session/cookie
    const storedRole = localStorage.getItem('mcp_user_role') as UserRole || 'admin';
    setUser({
      id: '1',
      name: 'Admin User',
      email: 'admin@mcp-any.io',
      role: storedRole, // Default to admin for dev
      avatar: '/avatars/admin.png'
    });
    setLoading(false);
  }, []);

  const login = (role: UserRole) => {
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
    <UserContext.Provider value={{ user, loading, login, logout }}>
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
}
