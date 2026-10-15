package ai

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/domain/services"
	apperrors "github.com/ailms/backend/pkg/errors"
)

// BudgetedService centralizes accounting immediately before a provider call.
// Validation and authorization have already completed in the caller. An
// attempted provider call consumes one unit even if the provider returns an
// error, because upstream providers can charge for failed responses.
type BudgetedService struct {
	underlying services.AIService
	usage      repositories.AIUsageRepository
	limit      int
}

func NewBudgetedService(underlying services.AIService, usage repositories.AIUsageRepository, limit int) *BudgetedService {
	return &BudgetedService{underlying: underlying, usage: usage, limit: limit}
}

func (s *BudgetedService) consume(ctx context.Context) error {
	if s.limit <= 0 {
		return nil
	}
	identity, ok := services.AIIdentityFromContext(ctx)
	if !ok {
		return apperrors.UnauthorizedError("AI request identity is missing")
	}
	_, allowed, err := s.usage.Consume(ctx, identity.OrgID, identity.UserID, time.Now().UTC(), s.limit)
	if err != nil {
		return err
	}
	if !allowed {
		return apperrors.TooManyRequestsError("daily AI request limit reached; try again tomorrow")
	}
	return nil
}

func (s *BudgetedService) GenerateLessonSummary(ctx context.Context, content string) (string, error) {
	if err := s.consume(ctx); err != nil {
		return "", err
	}
	return s.underlying.GenerateLessonSummary(ctx, content)
}
func (s *BudgetedService) GenerateQuiz(ctx context.Context, content string, count int) ([]entities.QuizQuestion, error) {
	if err := s.consume(ctx); err != nil {
		return nil, err
	}
	return s.underlying.GenerateQuiz(ctx, content, count)
}
func (s *BudgetedService) GenerateProgressInsights(ctx context.Context, progresses []entities.LessonProgress) (string, error) {
	if err := s.consume(ctx); err != nil {
		return "", err
	}
	return s.underlying.GenerateProgressInsights(ctx, progresses)
}
func (s *BudgetedService) GenerateRemediation(ctx context.Context, content string, questions []string) (string, error) {
	if err := s.consume(ctx); err != nil {
		return "", err
	}
	return s.underlying.GenerateRemediation(ctx, content, questions)
}
func (s *BudgetedService) GenerateAssignmentFeedback(ctx context.Context, title, description, text string) (*entities.AssignmentFeedback, error) {
	if err := s.consume(ctx); err != nil {
		return nil, err
	}
	return s.underlying.GenerateAssignmentFeedback(ctx, title, description, text)
}
func (s *BudgetedService) GenerateFlashcards(ctx context.Context, content string, count int) ([]entities.Flashcard, error) {
	if err := s.consume(ctx); err != nil {
		return nil, err
	}
	return s.underlying.GenerateFlashcards(ctx, content, count)
}
func (s *BudgetedService) GenerateTeachingPlan(ctx context.Context, topic, audience, language string, minutes int) (string, error) {
	if err := s.consume(ctx); err != nil {
		return "", err
	}
	return s.underlying.GenerateTeachingPlan(ctx, topic, audience, language, minutes)
}
