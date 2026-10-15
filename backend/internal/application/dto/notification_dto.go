package dto

import "time"

type NotificationDTO struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"` // announcement, lesson, assignment, exam, grade
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	CourseID    string     `json:"course_id"`
	CourseTitle string     `json:"course_title"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
