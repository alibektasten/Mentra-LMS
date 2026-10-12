package dto

import "time"

type CreateExamQuestionRequest struct {
	Question string                    `json:"question" binding:"required"`
	Position int                       `json:"position"`
	Answers  []CreateExamAnswerRequest `json:"answers" binding:"required,min=2"`
}

type CreateExamAnswerRequest struct {
	Answer    string `json:"answer" binding:"required"`
	IsCorrect bool   `json:"is_correct"`
}

type CreateExamRequest struct {
	AssessmentType  string                      `json:"assessment_type"`
	Title           string                      `json:"title" binding:"required"`
	Description     string                      `json:"description"`
	DurationMinutes int                         `json:"duration_minutes" binding:"required,min=1"`
	MaxAttempts     int                         `json:"max_attempts"`
	DueDate         *time.Time                  `json:"due_date"`
	MCQEnabled      bool                        `json:"mcq_enabled"`
	MCQPoints       int                         `json:"mcq_points"`
	FileEnabled     bool                        `json:"file_enabled"`
	FilePoints      int                         `json:"file_points"`
	Questions       []CreateExamQuestionRequest `json:"questions"`
}

type UpdateExamRequest struct {
	AssessmentType  *string                     `json:"assessment_type"`
	Title           *string                     `json:"title"`
	Description     *string                     `json:"description"`
	DurationMinutes *int                        `json:"duration_minutes"`
	MaxAttempts     *int                        `json:"max_attempts"`
	DueDate         *time.Time                  `json:"due_date"`
	MCQEnabled      *bool                       `json:"mcq_enabled"`
	MCQPoints       *int                        `json:"mcq_points"`
	FileEnabled     *bool                       `json:"file_enabled"`
	FilePoints      *int                        `json:"file_points"`
	Questions       []CreateExamQuestionRequest `json:"questions"`
}

type ExamAnswerDTO struct {
	ID        string `json:"id"`
	Answer    string `json:"answer"`
	IsCorrect bool   `json:"is_correct"`
}

type ExamQuestionDTO struct {
	ID       string          `json:"id"`
	Question string          `json:"question"`
	Position int             `json:"position"`
	Answers  []ExamAnswerDTO `json:"answers"`
}

type ExamDTO struct {
	ID              string            `json:"id"`
	CourseID        string            `json:"course_id"`
	OrgID           string            `json:"org_id"`
	AssessmentType  string            `json:"assessment_type"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	DurationMinutes int               `json:"duration_minutes"`
	MaxAttempts     int               `json:"max_attempts"`
	TotalPoints     int               `json:"total_points"`
	DueDate         *time.Time        `json:"due_date"`
	MCQEnabled      bool              `json:"mcq_enabled"`
	MCQPoints       int               `json:"mcq_points"`
	FileEnabled     bool              `json:"file_enabled"`
	FilePoints      int               `json:"file_points"`
	Questions       []ExamQuestionDTO `json:"questions"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type ExamListItemDTO struct {
	ID              string     `json:"id"`
	CourseID        string     `json:"course_id"`
	OrgID           string     `json:"org_id"`
	AssessmentType  string     `json:"assessment_type"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	DurationMinutes int        `json:"duration_minutes"`
	MaxAttempts     int        `json:"max_attempts"`
	TotalPoints     int        `json:"total_points"`
	DueDate         *time.Time `json:"due_date"`
	MCQEnabled      bool       `json:"mcq_enabled"`
	MCQPoints       int        `json:"mcq_points"`
	FileEnabled     bool       `json:"file_enabled"`
	FilePoints      int        `json:"file_points"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type StartAttemptResponse struct {
	AttemptID    string               `json:"attempt_id"`
	ExamID       string               `json:"exam_id"`
	StartedAt    time.Time            `json:"started_at"`
	ExpiresAt    time.Time            `json:"expires_at"`
	SavedAnswers []ExamMCQAnswerInput `json:"saved_answers"`
	Exam         *ExamDTO             `json:"exam"`
}

type SaveExamDraftRequest struct {
	MCQAnswers []ExamMCQAnswerInput `json:"mcq_answers"`
}

type ExamMCQAnswerInput struct {
	QuestionID string `json:"question_id"`
	AnswerID   string `json:"answer_id"`
}

type ExamAttemptDTO struct {
	ID           string               `json:"id"`
	ExamID       string               `json:"exam_id"`
	StudentID    string               `json:"student_id"`
	Status       string               `json:"status"`
	StartedAt    time.Time            `json:"started_at"`
	ExpiresAt    time.Time            `json:"expires_at"`
	SubmittedAt  *time.Time           `json:"submitted_at"`
	MCQAnswers   []ExamMCQAnswerInput `json:"mcq_answers"`
	MCQScore     *int                 `json:"mcq_score"`
	MCQMaxScore  int                  `json:"mcq_max_score"`
	FilePath     string               `json:"file_path"`
	FileScore    *int                 `json:"file_score"`
	FilePoints   int                  `json:"file_points"`
	FileFeedback string               `json:"file_feedback"`
	TotalScore   *int                 `json:"total_score"`
	GradedAt     *time.Time           `json:"graded_at"`
}

type GradeExamFileRequest struct {
	Score    int    `json:"score" binding:"gte=0"`
	Feedback string `json:"feedback"`
}

type GrantExtraAttemptRequest struct {
	StudentID  string `json:"student_id" binding:"required"`
	ExtraCount int    `json:"extra_count" binding:"required,min=1"`
}
