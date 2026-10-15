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

// TeacherCalendarEventDTO represents one dated item on a teacher's calendar.
// Assignment and exam events use their deadline as StartAt. Class events are
// expanded from a group's weekly schedule for the requested date range.
type TeacherCalendarEventDTO struct {
	ID             string     `json:"id"`
	ItemID         string     `json:"item_id"`
	ItemType       string     `json:"item_type"`
	Title          string     `json:"title"`
	CourseID       string     `json:"course_id"`
	CourseTitle    string     `json:"course_title"`
	GroupID        *string    `json:"group_id,omitempty"`
	GroupName      *string    `json:"group_name,omitempty"`
	AssessmentType string     `json:"assessment_type,omitempty"`
	Location       string     `json:"location,omitempty"`
	StartAt        time.Time  `json:"start_at"`
	EndAt          *time.Time `json:"end_at,omitempty"`
}
