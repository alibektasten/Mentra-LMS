import { useQuery } from '@tanstack/react-query'
import * as notificationsApi from '@/lib/api/notifications'

export function useNotifications() {
  return useQuery({ queryKey: ['notifications'], queryFn: notificationsApi.listNotifications, refetchInterval: 60_000 })
}
