package mocks

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) GenerateLessonSummary(ctx context.Context, lessonContent string) (string, error) {
	args := m.Called(ctx, lessonContent)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) GenerateQuiz(ctx context.Context, lessonContent string, numQuestions int) ([]entities.QuizQuestion, error) {
	args := m.Called(ctx, lessonContent, numQuestions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.QuizQuestion), args.Error(1)
}

func (m *MockAIService) GenerateProgressInsights(ctx context.Context, completedLessons []entities.LessonProgress) (string, error) {
	args := m.Called(ctx, completedLessons)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) GenerateRemediation(ctx context.Context, lessonContent string, wrongQuestions []string) (string, error) {
	args := m.Called(ctx, lessonContent, wrongQuestions)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) GenerateAssignmentFeedback(ctx context.Context, assignmentTitle, description, submissionText string) (*entities.AssignmentFeedback, error) {
	args := m.Called(ctx, assignmentTitle, description, submissionText)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AssignmentFeedback), args.Error(1)
}

func (m *MockAIService) GenerateFlashcards(ctx context.Context, lessonContent string, numCards int) ([]entities.Flashcard, error) {
	args := m.Called(ctx, lessonContent, numCards)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Flashcard), args.Error(1)
}

func (m *MockAIService) GenerateTeachingPlan(ctx context.Context, topic, audience, language string, durationMinutes int) (string, error) {
	args := m.Called(ctx, topic, audience, language, durationMinutes)
	return args.String(0), args.Error(1)
}
