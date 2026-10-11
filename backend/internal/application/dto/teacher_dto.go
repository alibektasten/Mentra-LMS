package dto

import "time"

type TeacherStudentDTO struct {
	UserID  string   `json:"user_id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Courses []string `json:"courses"`
	Groups  []string `json:"groups"`
}

type TeacherOverviewDTO struct {
	TotalCourses     int `json:"total_courses"`
	PublishedCourses int `json:"published_courses"`
	DraftCourses     int `json:"draft_courses"`
	TotalGroups      int `json:"total_groups"`
	TotalStudents    int `json:"total_students"`
}

type TeacherReviewItemDTO struct {
	SubmissionID    string    `json:"submission_id"`
	AssignmentID    string    `json:"assignment_id"`
	CourseID        string    `json:"course_id"`
	CourseTitle     string    `json:"course_title"`
	AssignmentTitle string    `json:"assignment_title"`
	StudentName     string    `json:"student_name"`
	SubmittedAt     time.Time `json:"submitted_at"`
	MaxPoints       int       `json:"max_points"`
}
