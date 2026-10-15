import { apiClient } from './client';

export interface UploadResult {
  path: string;
  name: string;
  upload_token: string;
}

export async function uploadFile(file: File): Promise<UploadResult> {
  const formData = new FormData();
  formData.append('file', file);
  const res = await apiClient.post<UploadResult>('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return res.data;
}
