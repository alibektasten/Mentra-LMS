import { apiClient } from './client'
import { NotificationDTO } from '@/types/notification'

export async function listNotifications(): Promise<NotificationDTO[]> {
  const response = await apiClient.get<NotificationDTO[]>('/notifications')
  return response.data ?? []
}
