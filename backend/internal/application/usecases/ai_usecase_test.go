package usecases

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAIUseCase_SummarizeLesson_Success(t *testing.T) {
	lessonRepo := new(mocks.MockLessonRepository)
	quizRepo := new(mocks.MockQuizRepository)
	assignmentRepo := new(mocks.MockAssignmentRepository)
	aiService := new(mocks.MockAIService)
	uc := NewAIUseCase(lessonRepo, quizRepo, assignmentRepo, aiService)

	orgID := uuid.New()
	lessonID := uuid.New()
	lesson := &entities.Lesson{
		ID:      lessonID,
		OrgID:   orgID,
		Title:   "Go Basics",
		Content: "Go is a compiled language...",
	}

	lessonRepo.On("FindByID", mock.Anything, lessonID, orgID).Return(lesson, nil)
	aiService.On("GenerateLessonSummary", mock.Anything, lesson.Content).Return("Go is fast and compiled.", nil)

	result, err := uc.SummarizeLesson(context.Background(), lessonID, orgID)
	require.NoError(t, err)
	assert.Equal(t, "Go is fast and compiled.", result.Summary)
}

func TestAIUseCase_SummarizeLesson_LessonNotFound(t *testing.T) {
	lessonRepo := new(mocks.MockLessonRepository)
	quizRepo := new(mocks.MockQuizRepository)
	assignmentRepo := new(mocks.MockAssignmentRepository)
	aiService := new(mocks.MockAIService)
	uc := NewAIUseCase(lessonRepo, quizRepo, assignmentRepo, aiService)

	orgID := uuid.New()
	lessonID := uuid.New()
	lessonRepo.On("FindByID", mock.Anything, lessonID, orgID).Return(nil, apperrors.NotFoundError("lesson", lessonID.String()))

	_, err := uc.SummarizeLesson(context.Background(), lessonID, orgID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.HTTPStatus)
}

func TestAIUseCase_SummarizeLesson_AIServiceFails(t *testing.T) {
	lessonRepo := new(mocks.MockLessonRepository)
	quizRepo := new(mocks.MockQuizRepository)
	assignmentRepo := new(mocks.MockAssignmentRepository)
	aiService := new(mocks.MockAIService)
	uc := NewAIUseCase(lessonRepo, quizRepo, assignmentRepo, aiService)

	orgID := uuid.New()
	lessonID := uuid.New()
	lesson := &entities.Lesson{ID: lessonID, OrgID: orgID, Content: "content"}

	lessonRepo.On("FindByID", mock.Anything, lessonID, orgID).Return(lesson, nil)
	aiService.On("GenerateLessonSummary", mock.Anything, lesson.Content).Return("", errors.New("AI unavailable"))

	_, err := uc.SummarizeLesson(context.Background(), lessonID, orgID)
	require.Error(t, err)
}

func TestAIUseCase_GenerateQuiz_Success(t *testing.T) {
	lessonRepo := new(mocks.MockLessonRepository)
	quizRepo := new(mocks.MockQuizRepository)
	assignmentRepo := new(mocks.MockAssignmentRepository)
	aiService := new(mocks.MockAIService)
	uc := NewAIUseCase(lessonRepo, quizRepo, assignmentRepo, aiService)

	orgID := uuid.New()
	lessonID := uuid.New()
	lesson := &entities.Lesson{
		ID:      lessonID,
		OrgID:   orgID,
		Title:   "Go Basics",
		Content: "Go is a compiled language...",
	}

	q1ID := uuid.New()
	a1ID := uuid.New()
	generatedQuestions := []entities.QuizQuestion{
		{
			ID:       q1ID,
			Question: "What type of language is Go?",
			Position: 1,
			Answers: []entities.QuizAnswer{
				{ID: a1ID, QuestionID: q1ID, Answer: "Compiled", IsCorrect: true},
				{ID: uuid.New(), QuestionID: q1ID, Answer: "Interpreted", IsCorrect: false},
			},
		},
	}

	lessonRepo.On("FindByID", mock.Anything, lessonID, orgID).Return(lesson, nil)
	aiService.On("GenerateQuiz", mock.Anything, lesson.Content, 1).Return(generatedQuestions, nil)

	result, err := uc.GenerateQuiz(context.Background(), lessonID, orgID, 1)
	require.NoError(t, err)
	assert.Equal(t, lessonID.String(), result.LessonID)
	assert.Len(t, result.Questions, 1)
	assert.Equal(t, "What type of language is Go?", result.Questions[0].Question)
	assert.Equal(t, "AI Generated Quiz for Go Basics", result.Title)
}
