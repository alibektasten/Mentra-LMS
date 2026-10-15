package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetStudentCourseProgressUsesScopedRepositoryAndPreservesCompletionState(t *testing.T) {
	repo := new(mocks.MockLessonProgressRepository)
	uc := NewProgressUseCase(repo, new(mocks.MockLessonRepository), new(mocks.MockAIService), new(mocks.MockModuleRepository), new(mocks.MockQuizRepository), new(mocks.MockQuizAttemptRepository))
	userID, orgID, courseA, courseB, lessonB := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	repo.On("FindByUserAndCourse", mock.Anything, userID, courseB, orgID).Return([]entities.LessonProgress{}, nil).Once()
	repo.On("FindByUserAndCourse", mock.Anything, userID, courseB, orgID).Return([]entities.LessonProgress{{ID: uuid.New(), UserID: userID, OrgID: orgID, LessonID: lessonB}}, nil).Once()
	repo.On("FindByUserAndCourse", mock.Anything, userID, courseB, orgID).Return([]entities.LessonProgress{{ID: uuid.New(), UserID: userID, OrgID: orgID, LessonID: lessonB, CompletedAt: &now}}, nil).Once()

	fresh, err := uc.GetStudentCourseProgress(context.Background(), userID, courseB, orgID)
	require.NoError(t, err)
	assert.Empty(t, fresh, "course A's progress must not be returned for fresh course B")
	pending, err := uc.GetStudentCourseProgress(context.Background(), userID, courseB, orgID)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Nil(t, pending[0].CompletedAt)
	completed, err := uc.GetStudentCourseProgress(context.Background(), userID, courseB, orgID)
	require.NoError(t, err)
	require.Len(t, completed, 1)
	assert.NotNil(t, completed[0].CompletedAt)
	repo.AssertNotCalled(t, "FindByUserAndCourse", mock.Anything, userID, courseA, orgID)
	repo.AssertExpectations(t)
}

func TestGetStudentCourseProgressReturnsRepositoryFailure(t *testing.T) {
	repo := new(mocks.MockLessonProgressRepository)
	uc := NewProgressUseCase(repo, new(mocks.MockLessonRepository), new(mocks.MockAIService), new(mocks.MockModuleRepository), new(mocks.MockQuizRepository), new(mocks.MockQuizAttemptRepository))
	userID, orgID, courseID := uuid.New(), uuid.New(), uuid.New()
	failed := errors.New("database unavailable")
	repo.On("FindByUserAndCourse", mock.Anything, userID, courseID, orgID).Return(nil, failed).Once()
	_, err := uc.GetStudentCourseProgress(context.Background(), userID, courseID, orgID)
	require.ErrorIs(t, err, failed)
}

func TestCoursePacingDoesNotCountPendingProgressAsCompleted(t *testing.T) {
	progressRepo := new(mocks.MockLessonProgressRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	lessonRepo := new(mocks.MockLessonRepository)
	quizRepo := new(mocks.MockQuizRepository)
	uc := NewProgressUseCase(progressRepo, lessonRepo, new(mocks.MockAIService), moduleRepo, quizRepo, new(mocks.MockQuizAttemptRepository))
	userID, orgID, courseID, moduleID, lessonID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	moduleRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]entities.Module{{ID: moduleID, CourseID: courseID, OrgID: orgID}}, nil)
	lessonRepo.On("FindByModule", mock.Anything, moduleID, orgID).Return([]entities.Lesson{{ID: lessonID, ModuleID: moduleID, OrgID: orgID}}, nil)
	progressRepo.On("FindByUser", mock.Anything, userID, orgID).Return([]entities.LessonProgress{{ID: uuid.New(), UserID: userID, OrgID: orgID, LessonID: lessonID}}, nil)
	quizRepo.On("FindByLessons", mock.Anything, []uuid.UUID{lessonID}, orgID).Return([]entities.Quiz{}, nil)

	pacing, err := uc.GetCoursePacing(context.Background(), courseID, userID, orgID)
	require.NoError(t, err)
	require.Len(t, pacing.Modules, 1)
	assert.Equal(t, 0, pacing.Modules[0].CompletedLessons)
	assert.Equal(t, 1, pacing.Modules[0].TotalLessons)
}
