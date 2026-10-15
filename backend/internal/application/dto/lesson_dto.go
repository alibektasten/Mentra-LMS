package dto

import "time"

type CreateLessonRequest struct {
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content"`
	Type        string `json:"type" binding:"required,oneof=video text quiz pdf link"`
	VideoURL    string `json:"video_url"`
	LinkURL     string `json:"link_url"`
	FileURL     string `json:"file_url"`
	UploadToken string `json:"upload_token"`
	Position    int    `json:"position"`
}

type UpdateLessonRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Type        string `json:"type" binding:"omitempty,oneof=video text quiz pdf link"`
	VideoURL    string `json:"video_url"`
	LinkURL     string `json:"link_url"`
	FileURL     string `json:"file_url"`
	UploadToken string `json:"upload_token"`
	Position    int    `json:"position"`
}

type LessonDTO struct {
	ID        string    `json:"id"`
	ModuleID  string    `json:"module_id"`
	OrgID     string    `json:"org_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	VideoURL  string    `json:"video_url"`
	LinkURL   string    `json:"link_url"`
	FileURL   string    `json:"file_url"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
