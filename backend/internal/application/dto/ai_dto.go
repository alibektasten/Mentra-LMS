package dto

type SummarizeLessonRequest struct {
	LessonID string `json:"lesson_id" binding:"required"`
}

type SummarizeLessonResponse struct {
	Summary string `json:"summary"`
}

type GenerateQuizRequest struct {
	LessonID     string `json:"lesson_id" binding:"required"`
	NumQuestions int    `json:"num_questions" binding:"required,min=1,max=20"`
}

type GenerateFlashcardsRequest struct {
	LessonID string `json:"lesson_id" binding:"required"`
	NumCards int    `json:"num_cards" binding:"required,min=1,max=30"`
}

type FlashcardDTO struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
}

type AssignmentFeedbackRequest struct {
	SubmissionID string `json:"submission_id" binding:"required"`
}

type AssignmentFeedbackResponse struct {
	Strengths    []string `json:"strengths"`
	Gaps         []string `json:"gaps"`
	Improvements []string `json:"improvements"`
	Overall      string   `json:"overall"`
}

type GenerateTeachingPlanRequest struct {
	Topic           string `json:"topic" binding:"required,min=3,max=240"`
	Audience        string `json:"audience" binding:"required,min=2,max=120"`
	DurationMinutes int    `json:"duration_minutes" binding:"required,min=10,max=300"`
	Language        string `json:"language" binding:"required,oneof=English Russian Kazakh"`
}

type GenerateTeachingPlanResponse struct {
	Plan string `json:"plan"`
}
