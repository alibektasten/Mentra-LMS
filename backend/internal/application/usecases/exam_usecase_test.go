package usecases

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newExamUC(examRepo *mocks.MockExamRepository, attemptRepo *mocks.MockExamAttemptRepository, grantRepo *mocks.MockExtraAttemptGrantRepository, courseRepo *mocks.MockCourseRepository) *ExamUseCase {
	return NewExamUseCase(examRepo, attemptRepo, grantRepo, courseRepo)
}

func makeCreateExamReq(mcqEnabled, fileEnabled bool) dto.CreateExamRequest {
	req := dto.CreateExamRequest{
		Title:           "Midterm",
		Description:     "desc",
		DurationMinutes: 60,
		MCQEnabled:      mcqEnabled,
		MCQPoints:       100,
		FileEnabled:     fileEnabled,
		FilePoints:      50,
	}
	if mcqEnabled {
		req.Questions = []dto.CreateExamQuestionRequest{
			{
				Question: "Q1",
				Position: 1,
				Answers: []dto.CreateExamAnswerRequest{
					{Answer: "A", IsCorrect: true},
					{Answer: "B", IsCorrect: false},
				},
			},
		}
	}
	return req
}

func TestExamUseCase_CreateExam_Success_MCQOnly(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	course := &entities.Course{ID: courseID, OrgID: orgID}

	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(course, nil)
	examRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Exam")).Return(nil)

	resp, err := uc.CreateExam(context.Background(), courseID, orgID, makeCreateExamReq(true, false))
	require.NoError(t, err)
	assert.True(t, resp.MCQEnabled)
	assert.False(t, resp.FileEnabled)
	assert.Len(t, resp.Questions, 1)
}

func TestExamUseCase_CreateExam_Success_FileOnly(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	course := &entities.Course{ID: courseID, OrgID: orgID}

	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(course, nil)
	examRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Exam")).Return(nil)

	req := dto.CreateExamRequest{Title: "Exam", DurationMinutes: 30, FileEnabled: true, FilePoints: 50}
	resp, err := uc.CreateExam(context.Background(), courseID, orgID, req)
	require.NoError(t, err)
	assert.False(t, resp.MCQEnabled)
	assert.True(t, resp.FileEnabled)
}

func TestExamUseCase_CreateExam_NoSectionEnabled_ReturnsValidation(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	course := &entities.Course{ID: courseID, OrgID: orgID}

	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(course, nil)

	req := dto.CreateExamRequest{Title: "Exam", DurationMinutes: 30}
	_, err := uc.CreateExam(context.Background(), courseID, orgID, req)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestExamUseCase_CreateExam_MCQEnabled_NoQuestions_ReturnsValidation(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	course := &entities.Course{ID: courseID, OrgID: orgID}

	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(course, nil)

	req := dto.CreateExamRequest{Title: "Exam", DurationMinutes: 30, MCQEnabled: true, Questions: nil}
	_, err := uc.CreateExam(context.Background(), courseID, orgID, req)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestExamUseCase_CreateExam_DefaultsMaxAttemptsTo1(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	course := &entities.Course{ID: courseID, OrgID: orgID}
	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(course, nil)

	var capturedExam *entities.Exam
	examRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Exam")).Run(func(args mock.Arguments) {
		capturedExam = args.Get(1).(*entities.Exam)
	}).Return(nil)

	req := makeCreateExamReq(true, false)
	req.MaxAttempts = 0
	_, err := uc.CreateExam(context.Background(), courseID, orgID, req)
	require.NoError(t, err)
	assert.Equal(t, 1, capturedExam.MaxAttempts)
}

func TestExamUseCase_StartAttempt_Success(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	studentID := uuid.New()

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	grantRepo.On("SumByExamAndStudent", mock.Anything, exam.ID, studentID).Return(0, nil)
	attemptRepo.On("StartOrResume", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt"), 1, true, mock.Anything, mock.Anything).
		Return(&entities.ExamAttempt{ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}, nil)

	resp, err := uc.StartAttempt(context.Background(), exam.ID, studentID, orgID)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AttemptID)
	assert.Equal(t, exam.ID.String(), resp.ExamID)
	assert.True(t, resp.ExpiresAt.After(resp.StartedAt))
}

func TestExamUseCase_StartAttempt_ReturnsExistingInProgressAttempt(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	studentID := uuid.New()

	existingAttempt := &entities.ExamAttempt{
		ID:         uuid.New(),
		ExamID:     exam.ID,
		StudentID:  studentID,
		Status:     "in_progress",
		StartedAt:  time.Now().Add(-5 * time.Minute),
		ExpiresAt:  time.Now().Add(55 * time.Minute),
		MCQAnswers: []entities.ExamMCQAnswer{{QuestionID: exam.Questions[0].ID.String(), AnswerID: exam.Questions[0].Answers[0].ID.String()}},
	}

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	grantRepo.On("SumByExamAndStudent", mock.Anything, exam.ID, studentID).Return(0, nil)
	attemptRepo.On("StartOrResume", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt"), 1, true, mock.Anything, mock.Anything).Return(existingAttempt, nil)

	resp, err := uc.StartAttempt(context.Background(), exam.ID, studentID, orgID)
	require.NoError(t, err)
	assert.Equal(t, existingAttempt.ID.String(), resp.AttemptID)
	assert.Equal(t, []dto.ExamMCQAnswerInput{{QuestionID: existingAttempt.MCQAnswers[0].QuestionID, AnswerID: existingAttempt.MCQAnswers[0].AnswerID}}, resp.SavedAnswers)
	// Create should NOT be called
	attemptRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestExamUseCase_StartAttempt_RetriesWithCommittedExamConfiguration(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	stale := makeExamWithQuestions(orgID, courseID, true, true)
	stale.FilePoints = 100
	stale.UpdatedAt = time.Now().UTC()
	committed := *stale
	committed.FilePoints = 40
	committed.UpdatedAt = stale.UpdatedAt.Add(time.Second)
	committedAttempt := &entities.ExamAttempt{ID: uuid.New(), ExamID: stale.ID, StudentID: studentID, OrgID: orgID, Status: "in_progress", FilePoints: 40, StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}

	examRepo.On("FindByID", mock.Anything, stale.ID, orgID).Return(stale, nil).Once()
	examRepo.On("FindByID", mock.Anything, stale.ID, orgID).Return(&committed, nil).Once()
	grantRepo.On("SumByExamAndStudent", mock.Anything, stale.ID, studentID).Return(0, nil).Once()
	attemptRepo.On("StartOrResume", mock.Anything,
		mock.MatchedBy(func(candidate *entities.ExamAttempt) bool { return candidate.FilePoints == 100 }),
		1, true, mock.Anything, mock.MatchedBy(func(version *time.Time) bool { return version != nil && version.Equal(stale.UpdatedAt) }),
	).Return(nil, repositories.ErrExamConfigurationChanged).Once()
	attemptRepo.On("StartOrResume", mock.Anything,
		mock.MatchedBy(func(candidate *entities.ExamAttempt) bool { return candidate.FilePoints == 40 }),
		1, true, mock.Anything, mock.MatchedBy(func(version *time.Time) bool { return version != nil && version.Equal(committed.UpdatedAt) }),
	).Return(committedAttempt, nil).Once()

	response, err := uc.StartAttempt(context.Background(), stale.ID, studentID, orgID)
	require.NoError(t, err)
	assert.Equal(t, 40, response.Exam.FilePoints)
	assert.Equal(t, committedAttempt.ID.String(), response.AttemptID)
	examRepo.AssertExpectations(t)
	attemptRepo.AssertExpectations(t)
}

func TestExamUseCase_StartAttempt_DueDatePassed_ReturnsValidation(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	pastDue := time.Now().Add(-24 * time.Hour)
	exam.DueDate = &pastDue
	studentID := uuid.New()

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	grantRepo.On("SumByExamAndStudent", mock.Anything, exam.ID, studentID).Return(0, nil)
	attemptRepo.On("StartOrResume", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt"), 1, false, mock.Anything, mock.Anything).
		Return(nil, apperrors.ValidationError("exam is no longer available for new attempts"))

	_, err := uc.StartAttempt(context.Background(), exam.ID, studentID, orgID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestExamUseCase_StartAttempt_DueDatePassed_ResumesExistingAttempt(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	pastDue := time.Now().Add(-time.Hour)
	exam.DueDate = &pastDue
	existing := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID,
		Status: "in_progress", StartedAt: time.Now().Add(-10 * time.Minute), ExpiresAt: time.Now().Add(50 * time.Minute),
	}

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	grantRepo.On("SumByExamAndStudent", mock.Anything, exam.ID, studentID).Return(0, nil)
	attemptRepo.On("StartOrResume", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt"), 1, false, mock.Anything, mock.Anything).Return(existing, nil)

	response, err := uc.StartAttempt(context.Background(), exam.ID, studentID, orgID)
	require.NoError(t, err)
	assert.Equal(t, existing.ID.String(), response.AttemptID)
}

func TestExamUseCase_StartAttempt_MaxAttemptsReached(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	exam.MaxAttempts = 2
	studentID := uuid.New()

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	grantRepo.On("SumByExamAndStudent", mock.Anything, exam.ID, studentID).Return(0, nil)
	attemptRepo.On("StartOrResume", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt"), 2, true, mock.Anything, mock.Anything).
		Return(nil, apperrors.ValidationError("maximum attempts reached"))

	_, err := uc.StartAttempt(context.Background(), exam.ID, studentID, orgID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestExamUseCase_StartAttempt_ExtraGrantsAllowMoreAttempts(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	exam.MaxAttempts = 1
	studentID := uuid.New()

	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	// Used 1, granted 1 extra => total allowed = 2, count=1 < 2 => OK
	grantRepo.On("SumByExamAndStudent", mock.Anything, exam.ID, studentID).Return(1, nil)
	attemptRepo.On("StartOrResume", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt"), 2, true, mock.Anything, mock.Anything).
		Return(&entities.ExamAttempt{ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}, nil)

	resp, err := uc.StartAttempt(context.Background(), exam.ID, studentID, orgID)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AttemptID)
}

func TestExamUseCase_SubmitAttempt_Success_MCQAutoGraded(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	studentID := uuid.New()

	attempt := &entities.ExamAttempt{
		ID:        uuid.New(),
		ExamID:    exam.ID,
		StudentID: studentID,
		OrgID:     orgID,
		Status:    "in_progress",
		StartedAt: time.Now().Add(-5 * time.Minute),
		ExpiresAt: time.Now().Add(55 * time.Minute),
	}

	// Answer correctly
	correctAnswerID := exam.Questions[0].Answers[0].ID.String() // IsCorrect = true
	questionID := exam.Questions[0].ID.String()

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SubmitIfInProgress", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(true, nil)

	mcqAnswers := []dto.ExamMCQAnswerInput{{QuestionID: questionID, AnswerID: correctAnswerID}}
	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID, mcqAnswers, "")
	require.NoError(t, err)
	assert.Equal(t, "submitted", result.Status)
	require.NotNil(t, result.MCQScore)
	// 1 correct out of 1 question, MCQPoints=100 => score = (1 * 100) / 1 = 100
	assert.Equal(t, 100, *result.MCQScore)
	require.NotNil(t, result.TotalScore)
	assert.Equal(t, 100, *result.TotalScore)
}

func TestExamUseCase_SubmitAttempt_NotOwner_ReturnsForbidden(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	studentID := uuid.New()
	otherStudentID := uuid.New()

	attempt := &entities.ExamAttempt{
		ID:        uuid.New(),
		StudentID: studentID,
		Status:    "in_progress",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)

	_, err := uc.SubmitAttempt(context.Background(), attempt.ID, otherStudentID, nil, "")
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusForbidden, appErr.HTTPStatus)
}

func TestExamUseCase_SubmitAttempt_RepeatedRequestReturnsSubmittedAttempt(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	studentID := uuid.New()
	attempt := &entities.ExamAttempt{
		ID:        uuid.New(),
		StudentID: studentID,
		Status:    "submitted",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)

	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID, nil, "")
	require.NoError(t, err)
	assert.Equal(t, "submitted", result.Status)
}

func TestExamUseCase_SubmitAttempt_WithinGracePeriod_Succeeds(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, false, true)
	studentID := uuid.New()

	// ExpiresAt 30s ago — within 60s grace period
	attempt := &entities.ExamAttempt{
		ID:        uuid.New(),
		ExamID:    exam.ID,
		StudentID: studentID,
		OrgID:     orgID,
		Status:    "in_progress",
		StartedAt: time.Now().Add(-90 * time.Second),
		ExpiresAt: time.Now().Add(-30 * time.Second),
	}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SubmitIfInProgress", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(true, nil)

	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID, nil, "")
	require.NoError(t, err)
	assert.Equal(t, "submitted", result.Status)
}

func TestExamUseCase_SubmitAttempt_AfterGracePeriod_ExpiresAttempt(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	studentID := uuid.New()
	attempt := &entities.ExamAttempt{
		ID:        uuid.New(),
		StudentID: studentID,
		Status:    "in_progress",
		StartedAt: time.Now().Add(-130 * time.Second),
		ExpiresAt: time.Now().Add(-70 * time.Second), // 70s ago > 60s grace
	}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	attemptRepo.On("ExpireIfInProgress", mock.Anything, attempt.ID, studentID, mock.Anything).Return(true, nil)

	_, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID, nil, "")
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestExamUseCase_SaveDraftWithinGrace_UsesConditionalUpdate(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	attempt := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID, Status: "in_progress",
		ExpiresAt: time.Now().Add(-30 * time.Second),
	}
	answers := []dto.ExamMCQAnswerInput{{QuestionID: exam.Questions[0].ID.String(), AnswerID: exam.Questions[0].Answers[0].ID.String()}}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SaveDraftIfInProgress", mock.Anything, attempt.ID, studentID, mock.Anything, mock.Anything).Return(true, nil)

	require.NoError(t, uc.SaveDraft(context.Background(), attempt.ID, studentID, answers))
	attemptRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestExamUseCase_SaveDraftAfterGrace_DoesNotOverwriteSubmittedAttempt(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	studentID := uuid.New()
	attempt := &entities.ExamAttempt{ID: uuid.New(), StudentID: studentID, Status: "in_progress", ExpiresAt: time.Now().Add(-2 * time.Minute)}
	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	// A zero-row compare-and-set represents a concurrent submit that won.
	attemptRepo.On("ExpireIfInProgress", mock.Anything, attempt.ID, studentID, mock.Anything).Return(false, nil)

	err := uc.SaveDraft(context.Background(), attempt.ID, studentID, nil)
	require.Error(t, err)
	attemptRepo.AssertNotCalled(t, "SaveDraftIfInProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestExamUseCase_SubmitAttempt_ConcurrentSecondSubmitReturnsCommittedResult(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	attempt := &entities.ExamAttempt{ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID, Status: "in_progress", ExpiresAt: time.Now().Add(time.Hour)}
	committedScore := 100
	committedAt := time.Now().UTC()
	committed := &entities.ExamAttempt{ID: attempt.ID, ExamID: exam.ID, StudentID: studentID, OrgID: orgID, Status: "submitted", MCQScore: &committedScore, TotalScore: &committedScore, SubmittedAt: &committedAt}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil).Once()
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SubmitIfInProgress", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(false, nil)
	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(committed, nil).Once()

	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID, []dto.ExamMCQAnswerInput{{QuestionID: exam.Questions[0].ID.String(), AnswerID: exam.Questions[0].Answers[0].ID.String()}}, "")
	require.NoError(t, err)
	assert.Equal(t, "submitted", result.Status)
	require.NotNil(t, result.TotalScore)
	assert.Equal(t, 100, *result.TotalScore)
}

func TestExamUseCase_UpdateExam_RejectsQuestionChangeAfterAttemptExists(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID, courseID := uuid.New(), uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	guardedQuestion := dto.CreateExamQuestionRequest{
		Question: "Replacement question", Position: 1,
		Answers: []dto.CreateExamAnswerRequest{{Answer: "Correct", IsCorrect: true}, {Answer: "Incorrect", IsCorrect: false}},
	}
	examRepo.On("UpdateWithAttemptGuard", mock.Anything, mock.AnythingOfType("*entities.Exam"), true).Return(false, nil).Once()

	_, err := uc.UpdateExam(context.Background(), exam.ID, orgID, dto.UpdateExamRequest{Questions: []dto.CreateExamQuestionRequest{guardedQuestion}})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
	examRepo.AssertExpectations(t)
}

func TestExamUseCase_UpdateExam_NormalizedEmptyAssessmentTypeUsesAttemptGuard(t *testing.T) {
	for _, rawAssessmentType := range []string{"", " \t "} {
		t.Run("after attempt: "+rawAssessmentType, func(t *testing.T) {
			examRepo := new(mocks.MockExamRepository)
			attemptRepo := new(mocks.MockExamAttemptRepository)
			grantRepo := new(mocks.MockExtraAttemptGrantRepository)
			courseRepo := new(mocks.MockCourseRepository)
			uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

			orgID, courseID := uuid.New(), uuid.New()
			exam := makeExamWithQuestions(orgID, courseID, true, false)
			exam.AssessmentType = "midterm"
			examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil).Once()
			examRepo.On("UpdateWithAttemptGuard", mock.Anything, mock.MatchedBy(func(updated *entities.Exam) bool {
				return updated.AssessmentType == "other"
			}), true).Return(false, nil).Once()

			_, err := uc.UpdateExam(context.Background(), exam.ID, orgID, dto.UpdateExamRequest{AssessmentType: &rawAssessmentType})
			require.Error(t, err)
			examRepo.AssertExpectations(t)
		})
	}
}

func TestExamUseCase_UpdateExam_NormalizedEmptyAssessmentTypeAllowedBeforeAttempt(t *testing.T) {
	for _, rawAssessmentType := range []string{"", " \t "} {
		t.Run("before attempt: "+rawAssessmentType, func(t *testing.T) {
			examRepo := new(mocks.MockExamRepository)
			attemptRepo := new(mocks.MockExamAttemptRepository)
			grantRepo := new(mocks.MockExtraAttemptGrantRepository)
			courseRepo := new(mocks.MockCourseRepository)
			uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

			orgID, courseID := uuid.New(), uuid.New()
			exam := makeExamWithQuestions(orgID, courseID, true, false)
			exam.AssessmentType = "midterm"
			examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil).Once()
			examRepo.On("UpdateWithAttemptGuard", mock.Anything, mock.MatchedBy(func(updated *entities.Exam) bool {
				return updated.AssessmentType == "other"
			}), true).Return(true, nil).Once()

			result, err := uc.UpdateExam(context.Background(), exam.ID, orgID, dto.UpdateExamRequest{AssessmentType: &rawAssessmentType})
			require.NoError(t, err)
			assert.Equal(t, "other", result.AssessmentType)
			examRepo.AssertExpectations(t)
		})
	}
}

func TestExamUseCase_UpdateExam_RejectsStaleVersion(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID, courseID := uuid.New(), uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, false, true)
	title := "Concurrent title"
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil).Once()
	examRepo.On("UpdateWithAttemptGuard", mock.Anything, mock.AnythingOfType("*entities.Exam"), false).
		Return(false, repositories.ErrExamConfigurationChanged).Once()

	_, err := uc.UpdateExam(context.Background(), exam.ID, orgID, dto.UpdateExamRequest{Title: &title})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
	examRepo.AssertExpectations(t)
}

func TestExamUseCase_SubmitAttempt_AllCorrect(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	studentID := uuid.New()

	attempt := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID,
		Status: "in_progress", ExpiresAt: time.Now().Add(time.Hour),
	}

	correctAnswerID := exam.Questions[0].Answers[0].ID.String()
	questionID := exam.Questions[0].ID.String()

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SubmitIfInProgress", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(true, nil)

	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID,
		[]dto.ExamMCQAnswerInput{{QuestionID: questionID, AnswerID: correctAnswerID}}, "")
	require.NoError(t, err)
	assert.Equal(t, 100, *result.MCQScore)
}

func TestExamUseCase_SubmitAttempt_NoneCorrect(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	studentID := uuid.New()

	attempt := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID,
		Status: "in_progress", ExpiresAt: time.Now().Add(time.Hour),
	}

	wrongAnswerID := exam.Questions[0].Answers[1].ID.String() // IsCorrect = false
	questionID := exam.Questions[0].ID.String()

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SubmitIfInProgress", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(true, nil)

	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID,
		[]dto.ExamMCQAnswerInput{{QuestionID: questionID, AnswerID: wrongAnswerID}}, "")
	require.NoError(t, err)
	assert.Equal(t, 0, *result.MCQScore)
}

func TestExamUseCase_SubmitAttempt_NoFileSection_SetsTotalFromMCQ(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, false)
	studentID := uuid.New()

	attempt := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID,
		Status: "in_progress", ExpiresAt: time.Now().Add(time.Hour),
	}

	correctAnswerID := exam.Questions[0].Answers[0].ID.String()
	questionID := exam.Questions[0].ID.String()

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("SubmitIfInProgress", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(true, nil)

	result, err := uc.SubmitAttempt(context.Background(), attempt.ID, studentID,
		[]dto.ExamMCQAnswerInput{{QuestionID: questionID, AnswerID: correctAnswerID}}, "")
	require.NoError(t, err)
	require.NotNil(t, result.TotalScore)
	assert.Equal(t, *result.MCQScore, *result.TotalScore)
}

func TestExamUseCase_GradeFileSection_Success_RecomputesTotal(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	exam := makeExamWithQuestions(orgID, courseID, true, true)
	graderID := uuid.New()

	mcqScore := 80
	attempt := &entities.ExamAttempt{
		ID:         uuid.New(),
		ExamID:     exam.ID,
		OrgID:      orgID,
		Status:     "submitted",
		MCQScore:   &mcqScore,
		FilePath:   "uploads/file.pdf",
		FilePoints: exam.FilePoints,
	}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)
	examRepo.On("FindByID", mock.Anything, exam.ID, orgID).Return(exam, nil)
	attemptRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.ExamAttempt")).Return(nil)

	result, err := uc.GradeFileSection(context.Background(), attempt.ID, graderID, dto.GradeExamFileRequest{Score: 40, Feedback: "Good"})
	require.NoError(t, err)
	require.NotNil(t, result.TotalScore)
	assert.Equal(t, 120, *result.TotalScore) // 80 (MCQ) + 40 (file)
	assert.Equal(t, 40, *result.FileScore)
	assert.Equal(t, "Good", result.FileFeedback)
}

func TestExamUseCase_GradeFileSection_NotSubmitted_ReturnsValidation(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	attempt := &entities.ExamAttempt{
		ID:     uuid.New(),
		Status: "in_progress",
	}

	attemptRepo.On("FindByID", mock.Anything, attempt.ID).Return(attempt, nil)

	_, err := uc.GradeFileSection(context.Background(), attempt.ID, uuid.New(), dto.GradeExamFileRequest{Score: 10})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestExamUseCase_GrantExtraAttempt_Success(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	examID := uuid.New()
	grantedByID := uuid.New()
	orgID := uuid.New()
	studentID := uuid.New()

	grantRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.ExtraAttemptGrant")).Return(nil)

	err := uc.GrantExtraAttempt(context.Background(), examID, grantedByID, orgID, dto.GrantExtraAttemptRequest{
		StudentID:  studentID.String(),
		ExtraCount: 2,
	})
	require.NoError(t, err)
}

func TestExamUseCase_GrantExtraAttempt_InvalidStudentID_ReturnsValidation(t *testing.T) {
	examRepo := new(mocks.MockExamRepository)
	attemptRepo := new(mocks.MockExamAttemptRepository)
	grantRepo := new(mocks.MockExtraAttemptGrantRepository)
	courseRepo := new(mocks.MockCourseRepository)
	uc := newExamUC(examRepo, attemptRepo, grantRepo, courseRepo)

	err := uc.GrantExtraAttempt(context.Background(), uuid.New(), uuid.New(), uuid.New(), dto.GrantExtraAttemptRequest{
		StudentID:  "not-a-uuid",
		ExtraCount: 1,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}
