export interface FileAttachmentDTO {
  id: string
  ref_type: string
  ref_id: string
  stored_path: string
  original_name: string
  mime_type: string
  size_bytes: number
  created_at: string
}

export interface CreateAttachmentInput {
  ref_type: string
  ref_id: string
  stored_path: string
  upload_token?: string
  original_name: string
  mime_type?: string
  size_bytes?: number
}
