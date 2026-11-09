import { create } from 'zustand';

interface UIState {
  newShell: boolean;
  setNewShell: (v: boolean) => void;
  theme: 'light' | 'dark' | 'system';
  setTheme: (t: 'light' | 'dark' | 'system') => void;
  sidebarOpen: boolean;
  setSidebarOpen: (v: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
  newShell: false,
  setNewShell: (v) => set({ newShell: v }),
  theme: 'system',
  setTheme: (t) => set({ theme: t }),
  sidebarOpen: true,
  setSidebarOpen: (v) => set({ sidebarOpen: v }),
}));
