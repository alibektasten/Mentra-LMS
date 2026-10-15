package dto

import "time"

type CreateAttachmentRequest struct {
	RefType      string `json:"ref_type" binding:"required"`
	RefID        string `json:"ref_id" binding:"required"`
	StoredPath   string `json:"stored_path" binding:"required"`
	UploadToken  string `json:"upload_token"`
	OriginalName string `json:"original_name" binding:"required"`
	MimeType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
}

type FileAttachmentDTO struct {
	ID           string    `json:"id"`
	RefType      string    `json:"ref_type"`
	RefID        string    `json:"ref_id"`
	StoredPath   string    `json:"stored_path"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	CreatedAt    time.Time `json:"created_at"`
}
