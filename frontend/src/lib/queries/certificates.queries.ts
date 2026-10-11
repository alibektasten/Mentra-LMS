import { useQuery } from '@tanstack/react-query'
import * as certificatesApi from '@/lib/api/certificates'

export function useCertificates() {
  return useQuery({ queryKey: ['certificates'], queryFn: certificatesApi.listCertificates })
}
