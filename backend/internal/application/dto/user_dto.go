package dto

import "time"

type UpdateProfileRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=128"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,max=2048"`
}

type ProfileResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
