package dto

import "time"

type CertificateDTO struct {
	ID               string    `json:"id"`
	CourseID         string    `json:"course_id"`
	CourseTitle      string    `json:"course_title"`
	VerificationCode string    `json:"verification_code"`
	IssuedAt         time.Time `json:"issued_at"`
}
