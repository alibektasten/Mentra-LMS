package entities

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleTeacher    Role = "teacher"
	RoleStudent    Role = "student"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	AvatarURL    string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) ValidatePassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain))
	return err == nil
}
