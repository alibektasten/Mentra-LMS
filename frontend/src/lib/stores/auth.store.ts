'use client'

import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { AuthProvider, AuthSession, Role, UserDTO } from '@/types/auth'
import { clearSessionQueryCache } from '@/lib/query/session-cache'
import { useReAuthStore } from '@/lib/stores/reauth.store'

interface AuthState {
  token: string | undefined
  user: UserDTO | undefined
  role: Role | undefined
  orgSlug: string | undefined
  expiresAt: string | undefined
  authProvider: AuthProvider | undefined
  sessionGeneration: number
  setSession: (session: AuthSession) => void
  setOrgSlug: (slug: string) => void
  clearSession: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: undefined,
      user: undefined,
      role: undefined,
      orgSlug: undefined,
      expiresAt: undefined,
      authProvider: undefined,
      sessionGeneration: 0,

      setSession: (session: AuthSession) => {
        const previous = useAuthStore.getState()
        const identityChanged = Boolean(
          (previous.user?.id && previous.user.id !== session.user.id) ||
          (previous.orgSlug && previous.orgSlug !== session.orgSlug) ||
          (previous.role && previous.role !== session.role)
        )
        if (identityChanged) {
          clearSessionQueryCache()
        }
        set({
          token: session.token,
          user: session.user,
          role: session.role,
          orgSlug: session.orgSlug,
          expiresAt: session.expiresAt,
          authProvider: session.authProvider ?? 'password',
          sessionGeneration: previous.sessionGeneration + (identityChanged ? 1 : 0),
        })
        // Set cookie so Next.js middleware can read role
        if (typeof document !== 'undefined') {
          document.cookie = `mentra-role=${session.role}; path=/; max-age=${60 * 60 * 24 * 7}`
          document.cookie = `mentra-org=${session.orgSlug}; path=/; max-age=${60 * 60 * 24 * 7}`
        }
      },

      setOrgSlug: (slug: string) => {
        const previous = useAuthStore.getState()
        const changed = previous.orgSlug !== slug
        if (changed) clearSessionQueryCache()
        set({ orgSlug: slug, sessionGeneration: previous.sessionGeneration + (changed ? 1 : 0) })
      },

      clearSession: () => {
        const nextGeneration = useAuthStore.getState().sessionGeneration + 1
        useReAuthStore.getState().reject()
        clearSessionQueryCache()
        set({
          token: undefined,
          user: undefined,
          role: undefined,
          orgSlug: undefined,
          expiresAt: undefined,
          authProvider: undefined,
          sessionGeneration: nextGeneration,
        })
        if (typeof document !== 'undefined') {
          document.cookie = 'mentra-role=; path=/; max-age=0'
          document.cookie = 'mentra-org=; path=/; max-age=0'
        }
      },
    }),
    {
      name: 'mentra-auth',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        role: state.role,
        orgSlug: state.orgSlug,
        expiresAt: state.expiresAt,
        authProvider: state.authProvider,
      }),
    }
  )
)
