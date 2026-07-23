import { create } from 'zustand'

interface AppState {
  ready: boolean
  setReady: (ready: boolean) => void
}

export const useAppStore = create<AppState>((set) => ({
  ready: true,
  setReady: (ready) => set({ ready }),
}))
