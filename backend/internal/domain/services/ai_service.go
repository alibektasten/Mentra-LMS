package services

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
)

type AIService interface {
	GenerateLessonSummary(ctx context.Context, lessonContent string) (string, error)
	GenerateQuiz(ctx context.Context, lessonContent string, numQuestions int) ([]entities.QuizQuestion, error)
	GenerateProgressInsights(ctx context.Context, completedLessons []entities.LessonProgress) (string, error)
	GenerateRemediation(ctx context.Context, lessonContent string, wrongQuestions []string) (string, error)
	GenerateAssignmentFeedback(ctx context.Context, assignmentTitle, description, submissionText string) (*entities.AssignmentFeedback, error)
	GenerateFlashcards(ctx context.Context, lessonContent string, numCards int) ([]entities.Flashcard, error)
	GenerateTeachingPlan(ctx context.Context, topic, audience, language string, durationMinutes int) (string, error)
}
