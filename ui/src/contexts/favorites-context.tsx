"use client";

import React, { createContext, useContext, useEffect, useState } from "react";

interface FavoritesContextType {
  pinnedToolNames: string[];
  togglePin: (toolName: string) => void;
  isPinned: (toolName: string) => boolean;
}

const FavoritesContext = createContext<FavoritesContextType | undefined>(undefined);

const STORAGE_KEY = "mcpany_pinned_tools";

export function FavoritesProvider({ children }: { children: React.ReactNode }) {
  const [pinnedToolNames, setPinnedToolNames] = useState<string[]>([]);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      try {
        setPinnedToolNames(JSON.parse(saved));
      } catch (e) {
        console.error("Failed to parse pinned tools", e);
      }
    }
    setLoaded(true);
  }, []);

  useEffect(() => {
    if (loaded) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(pinnedToolNames));
    }
  }, [pinnedToolNames, loaded]);

  const togglePin = (toolName: string) => {
    setPinnedToolNames((prev) => {
      if (prev.includes(toolName)) {
        return prev.filter((name) => name !== toolName);
      } else {
        return [...prev, toolName];
      }
    });
  };

  const isPinned = (toolName: string) => pinnedToolNames.includes(toolName);

  return (
    <FavoritesContext.Provider value={{ pinnedToolNames, togglePin, isPinned }}>
      {children}
    </FavoritesContext.Provider>
  );
}

export function useFavorites() {
  const context = useContext(FavoritesContext);
  if (context === undefined) {
    throw new Error("useFavorites must be used within a FavoritesProvider");
  }
  return context;
}
