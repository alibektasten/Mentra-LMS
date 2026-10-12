package dto

import "time"

type CertificateDTO struct {
	ID               string    `json:"id"`
	CourseID         string    `json:"course_id"`
	CourseTitle      string    `json:"course_title"`
	StudentName      string    `json:"student_name"`
	TeacherName      string    `json:"teacher_name"`
	OrganizationName string    `json:"organization_name"`
	VerificationCode string    `json:"verification_code"`
	IssuedAt         time.Time `json:"issued_at"`
}
