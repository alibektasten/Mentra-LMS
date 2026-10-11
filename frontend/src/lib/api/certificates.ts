import { apiClient } from './client'
import { CertificateDTO } from '@/types/certificate'

export async function listCertificates(): Promise<CertificateDTO[]> {
  const response = await apiClient.get<CertificateDTO[]>('/certificates')
  return response.data ?? []
}
