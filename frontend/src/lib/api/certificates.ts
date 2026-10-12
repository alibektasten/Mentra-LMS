import axios from 'axios'
import { API_URL, apiClient } from './client'
import { CertificateDTO } from '@/types/certificate'

export async function listCertificates(): Promise<CertificateDTO[]> {
  const response = await apiClient.get<CertificateDTO[]>('/certificates')
  return response.data ?? []
}

export async function verifyCertificate(code: string): Promise<CertificateDTO> {
  const response = await axios.get<{ data: CertificateDTO }>(`${API_URL}/certificates/${encodeURIComponent(code)}`)
  return response.data.data
}
