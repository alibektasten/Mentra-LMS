'use client'

import type { QueryClient } from '@tanstack/react-query'

// React Query intentionally has one client for the browser. Keep the client
// registration here so authentication can cancel and drop private data without
// every feature query having to remember user/org values in its query key.
let activeQueryClient: QueryClient | null = null

export function registerSessionQueryClient(client: QueryClient) {
  activeQueryClient = client
  return () => {
    if (activeQueryClient === client) activeQueryClient = null
  }
}

export function clearSessionQueryCache() {
  const client = activeQueryClient
  if (!client) return
  // Cancelling first prevents a response from the previous identity from
  // populating the fresh session after its cache has been cleared.
  void client.cancelQueries()
  client.clear()
}
