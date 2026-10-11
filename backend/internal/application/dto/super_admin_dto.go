package dto

import "time"

type OrgDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminUserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type SystemStatsDTO struct {
	TotalOrgs  int64 `json:"total_orgs"`
	TotalUsers int64 `json:"total_users"`
}

type InviteOrgAdminRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Name     string `json:"name"     binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	OrgID    string `json:"org_id"   binding:"required"`
}

type CreateOrganizationRequest struct {
	Name string `json:"name" binding:"required,min=2,max=120"`
	Slug string `json:"slug" binding:"required,min=2,max=63,alphanum"`
}
