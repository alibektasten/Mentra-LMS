package dto

import "time"

type AttendanceStudentDTO struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Status    string `json:"status"`
}

type AttendanceSheetDTO struct {
	GroupID  string                 `json:"group_id"`
	CourseID string                 `json:"course_id"`
	Date     time.Time              `json:"date"`
	Students []AttendanceStudentDTO `json:"students"`
}

type AttendanceMarkInput struct {
	StudentID string `json:"student_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

type SaveAttendanceRequest struct {
	Date    string                `json:"date" binding:"required"`
	Records []AttendanceMarkInput `json:"records" binding:"required"`
}
