'use client'

import { create } from 'zustand'

interface ReAuthCallbacks {
  onSuccess: (token: string) => void
  onCancel: () => void
}

interface ReAuthState {
  isOpen: boolean
  prefillEmail: string
  callbacks: ReAuthCallbacks | null
  open: (email: string, callbacks?: ReAuthCallbacks) => void
  close: () => void
  resolve: (token: string) => void
  reject: () => void
}

export const useReAuthStore = create<ReAuthState>()((set, get) => ({
  isOpen: false,
  prefillEmail: '',
  callbacks: null,

  open: (email: string, callbacks?: ReAuthCallbacks) => {
    if (get().isOpen) return
    set({ isOpen: true, prefillEmail: email, callbacks: callbacks ?? null })
  },

  close: () => {
    set({ isOpen: false, prefillEmail: '', callbacks: null })
  },

  resolve: (token: string) => {
    const { callbacks } = get()
    callbacks?.onSuccess(token)
    set({ isOpen: false, prefillEmail: '', callbacks: null })
  },

  reject: () => {
    const { callbacks } = get()
    callbacks?.onCancel()
    set({ isOpen: false, prefillEmail: '', callbacks: null })
  },
}))
