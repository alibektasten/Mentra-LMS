package repositories

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/domain/entities"
	domainrepos "github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/infrastructure/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGORMExamRepository_UpdatePreservesAssessmentType(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMExamRepository(db)
	ctx := context.Background()
	orgID, courseID := uuid.New(), uuid.New()
	now := time.Now().UTC()
	examIDs := make([]string, 0, 3)
	t.Cleanup(func() { _ = db.Where("id IN ?", examIDs).Delete(&database.ExamModel{}).Error })

	for _, assessmentType := range []string{"midterm", "endterm", "final"} {
		exam := academicIntegrationExam(orgID, courseID, assessmentType, now)
		examIDs = append(examIDs, exam.ID.String())
		require.NoError(t, repo.Create(ctx, exam))
		exam.Title = "Renamed " + assessmentType
		require.NoError(t, repo.Update(ctx, exam))

		reloaded, err := repo.FindByID(ctx, exam.ID, orgID)
		require.NoError(t, err)
		assert.Equal(t, assessmentType, reloaded.AssessmentType)
	}
}

func TestGORMExamRepository_UpdateDueDatePreservesAssessmentType(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMExamRepository(db)
	ctx := context.Background()
	exam := academicIntegrationExam(uuid.New(), uuid.New(), "final", time.Now().UTC())
	require.NoError(t, repo.Create(ctx, exam))
	t.Cleanup(func() { _ = db.Where("id = ?", exam.ID.String()).Delete(&database.ExamModel{}).Error })

	due := time.Now().UTC().Add(72 * time.Hour)
	exam.DueDate = &due
	require.NoError(t, repo.Update(ctx, exam))
	reloaded, err := repo.FindByID(ctx, exam.ID, exam.OrgID)
	require.NoError(t, err)
	assert.Equal(t, "final", reloaded.AssessmentType)
	require.NotNil(t, reloaded.DueDate)
	assert.True(t, reloaded.DueDate.Equal(due))
}

func TestGORMExamRepository_AllowedTypeChangePersistsAndFeedsGradebookComponent(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	examRepo := NewGORMExamRepository(db)
	attemptRepo := NewGORMExamAttemptRepository(db)
	ctx := context.Background()
	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	exam := academicIntegrationExam(orgID, courseID, "other", time.Now().UTC())
	require.NoError(t, examRepo.Create(ctx, exam))
	t.Cleanup(func() {
		_ = db.Where("exam_id = ?", exam.ID.String()).Delete(&database.ExamAttemptModel{}).Error
		_ = db.Where("id = ?", exam.ID.String()).Delete(&database.ExamModel{}).Error
	})

	exam.AssessmentType = "midterm"
	updated, err := examRepo.UpdateWithAttemptGuard(ctx, exam, true)
	require.NoError(t, err)
	require.True(t, updated)
	reloaded, err := examRepo.FindByID(ctx, exam.ID, orgID)
	require.NoError(t, err)
	require.Equal(t, "midterm", reloaded.AssessmentType)

	score := 80
	submittedAt := time.Now().UTC()
	require.NoError(t, attemptRepo.Create(ctx, &entities.ExamAttempt{
		ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID, Status: "submitted",
		StartedAt: submittedAt.Add(-time.Minute), ExpiresAt: submittedAt.Add(time.Hour), SubmittedAt: &submittedAt,
		TotalScore: &score, CreatedAt: submittedAt, UpdatedAt: submittedAt,
	}))

	gradeUC := usecases.NewGradeUseCase(NewGORMAssignmentRepository(db), NewGORMQuizAttemptRepository(db), NewGORMQuizRepository(db), nil, examRepo, attemptRepo)
	grade, err := gradeUC.GetMyGrades(ctx, courseID, studentID, orgID)
	require.NoError(t, err)
	require.NotNil(t, grade.MidtermPercent)
	assert.Equal(t, 80.0, *grade.MidtermPercent)
	assert.Nil(t, grade.EndtermPercent)
	assert.Nil(t, grade.FinalExamPercent)
}

func TestGORMExamRepositories_ConcurrentStartAndStructuralUpdateNeverMixVersions(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	examRepo := NewGORMExamRepository(db)
	attemptRepo := NewGORMExamAttemptRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	exam := academicIntegrationExam(orgID, courseID, "other", time.Now().UTC())
	exam.FilePoints = 100
	require.NoError(t, examRepo.Create(ctx, exam))
	// Use the persisted timestamp precision as the version observed by Start.
	// PostgreSQL normalises timestamps, so the pre-insert Go value is not a
	// reliable compare-and-set token.
	exam, err := examRepo.FindByID(ctx, exam.ID, orgID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Where("exam_id = ?", exam.ID.String()).Delete(&database.ExamAttemptModel{}).Error
		_ = db.Where("id = ?", exam.ID.String()).Delete(&database.ExamModel{}).Error
	})

	// Hold the shared lock first. Both operations are released together through
	// a channel barrier; no timing sleep determines their order.
	gate := db.Begin()
	require.NoError(t, gate.Error)
	require.NoError(t, lockExamConfiguration(gate, exam.ID))
	ready := make(chan struct{}, 2)
	type startResult struct {
		attempt *entities.ExamAttempt
		err     error
	}
	startDone := make(chan startResult, 1)
	updateDone := make(chan struct {
		updated bool
		err     error
	}, 1)

	now := time.Now().UTC()
	candidate := &entities.ExamAttempt{ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID, Status: "in_progress", StartedAt: now, ExpiresAt: now.Add(time.Hour), FilePoints: 100, CreatedAt: now, UpdatedAt: now}
	updatedExam := *exam
	updatedExam.FilePoints = 40
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		ready <- struct{}{}
		attempt, err := attemptRepo.StartOrResume(ctx, candidate, 1, true, now, &exam.UpdatedAt)
		startDone <- startResult{attempt: attempt, err: err}
	}()
	go func() {
		defer wg.Done()
		ready <- struct{}{}
		updated, err := examRepo.UpdateWithAttemptGuard(ctx, &updatedExam, true)
		updateDone <- struct {
			updated bool
			err     error
		}{updated: updated, err: err}
	}()
	<-ready
	<-ready
	require.NoError(t, gate.Commit().Error)
	wg.Wait()
	start := <-startDone
	update := <-updateDone
	require.NoError(t, update.err)

	reloaded, err := examRepo.FindByID(ctx, exam.ID, orgID)
	require.NoError(t, err)
	if update.updated {
		require.Error(t, start.err)
		assert.True(t, errors.Is(start.err, domainrepos.ErrExamConfigurationChanged))
		assert.Equal(t, 40, reloaded.FilePoints)
		var attempts int64
		require.NoError(t, db.Model(&database.ExamAttemptModel{}).Where("exam_id = ?", exam.ID.String()).Count(&attempts).Error)
		assert.Zero(t, attempts, "a stale start must not create an attempt after the edit wins")
	} else {
		require.NoError(t, start.err)
		require.NotNil(t, start.attempt)
		assert.Equal(t, 100, start.attempt.FilePoints)
		assert.Equal(t, 100, reloaded.FilePoints, "the structural edit must be rejected after start wins")
	}
}

func TestGORMExamUseCase_StaleTitleUpdateCannotRevertStartedConfiguration(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	examRepo := NewGORMExamRepository(db)
	attemptRepo := NewGORMExamAttemptRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	exam := academicIntegrationExam(orgID, courseID, "other", now)
	exam.MCQEnabled = true
	exam.MCQPoints = 60
	exam.FileEnabled = true
	exam.FilePoints = 40
	exam.Questions = []entities.ExamQuestion{academicIntegrationQuestion(exam.ID, "Original question", now)}
	require.NoError(t, examRepo.Create(ctx, exam))
	t.Cleanup(func() {
		_ = db.Where("exam_id = ?", exam.ID.String()).Delete(&database.ExamAttemptModel{}).Error
		_ = db.Where("question_id IN (SELECT id FROM exam_questions WHERE exam_id = ?)", exam.ID.String()).Delete(&database.ExamAnswerModel{}).Error
		_ = db.Where("exam_id = ?", exam.ID.String()).Delete(&database.ExamQuestionModel{}).Error
		_ = db.Where("id = ?", exam.ID.String()).Delete(&database.ExamModel{}).Error
	})

	staleRead := make(chan struct{})
	releaseTitleUpdate := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(releaseTitleUpdate) })
	staleRepo := &pauseAfterFirstExamReadRepository{
		GORMExamRepository: examRepo,
		staleRead:          staleRead,
		release:            releaseTitleUpdate,
	}
	uc := usecases.NewExamUseCase(staleRepo, attemptRepo, nil, nil)
	title := "Stale title"
	titleUpdateDone := make(chan error, 1)
	go func() {
		_, err := uc.UpdateExam(ctx, exam.ID, orgID, dto.UpdateExamRequest{Title: &title})
		titleUpdateDone <- err
	}()
	<-staleRead // UpdateExam has a stale full model but has not acquired the write lock yet.

	structural, err := examRepo.FindByID(ctx, exam.ID, orgID)
	require.NoError(t, err)
	structural.MCQPoints = 25
	structural.FilePoints = 75
	structural.Questions = []entities.ExamQuestion{academicIntegrationQuestion(exam.ID, "New protected question", now.Add(time.Second))}
	updated, err := examRepo.UpdateWithAttemptGuard(ctx, structural, true)
	require.NoError(t, err)
	require.True(t, updated)

	configured, err := examRepo.FindByID(ctx, exam.ID, orgID)
	require.NoError(t, err)
	candidateTime := time.Now().UTC()
	attempt, err := attemptRepo.StartOrResume(ctx, &entities.ExamAttempt{
		ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID, Status: "in_progress",
		StartedAt: candidateTime, ExpiresAt: candidateTime.Add(time.Hour), FilePoints: configured.FilePoints,
		CreatedAt: candidateTime, UpdatedAt: candidateTime,
	}, configured.MaxAttempts, true, candidateTime, &configured.UpdatedAt)
	require.NoError(t, err)

	releaseOnce.Do(func() { close(releaseTitleUpdate) })
	require.Error(t, <-titleUpdateDone, "a title-only request based on the old version must not save a stale full model")

	reloaded, err := examRepo.FindByID(ctx, exam.ID, orgID)
	require.NoError(t, err)
	assert.Equal(t, "Academic integration exam", reloaded.Title)
	assert.Equal(t, 25, reloaded.MCQPoints)
	assert.Equal(t, 75, reloaded.FilePoints)
	require.Len(t, reloaded.Questions, 1)
	assert.Equal(t, "New protected question", reloaded.Questions[0].Question)
	assert.Equal(t, 75, attempt.FilePoints)
}

func TestGORMExamUseCase_NormalizedEmptyAssessmentTypeHonoursAttemptGuard(t *testing.T) {
	for _, rawAssessmentType := range []string{"", " \t "} {
		t.Run("after attempt: "+rawAssessmentType, func(t *testing.T) {
			db := openAcademicIntegrationDB(t)
			examRepo := NewGORMExamRepository(db)
			attemptRepo := NewGORMExamAttemptRepository(db)
			ctx := context.Background()
			orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
			now := time.Now().UTC()
			exam := academicIntegrationExam(orgID, courseID, "midterm", now)
			require.NoError(t, examRepo.Create(ctx, exam))
			t.Cleanup(func() {
				_ = db.Where("exam_id = ?", exam.ID.String()).Delete(&database.ExamAttemptModel{}).Error
				_ = db.Where("id = ?", exam.ID.String()).Delete(&database.ExamModel{}).Error
			})
			require.NoError(t, attemptRepo.Create(ctx, &entities.ExamAttempt{
				ID: uuid.New(), ExamID: exam.ID, StudentID: studentID, OrgID: orgID, Status: "submitted",
				StartedAt: now, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now,
			}))

			uc := usecases.NewExamUseCase(examRepo, attemptRepo, nil, nil)
			_, err := uc.UpdateExam(ctx, exam.ID, orgID, dto.UpdateExamRequest{AssessmentType: &rawAssessmentType})
			require.Error(t, err)
			reloaded, readErr := examRepo.FindByID(ctx, exam.ID, orgID)
			require.NoError(t, readErr)
			assert.Equal(t, "midterm", reloaded.AssessmentType)
		})

		t.Run("before attempt: "+rawAssessmentType, func(t *testing.T) {
			db := openAcademicIntegrationDB(t)
			examRepo := NewGORMExamRepository(db)
			ctx := context.Background()
			orgID, courseID := uuid.New(), uuid.New()
			exam := academicIntegrationExam(orgID, courseID, "midterm", time.Now().UTC())
			require.NoError(t, examRepo.Create(ctx, exam))
			t.Cleanup(func() { _ = db.Where("id = ?", exam.ID.String()).Delete(&database.ExamModel{}).Error })

			uc := usecases.NewExamUseCase(examRepo, nil, nil, nil)
			updated, err := uc.UpdateExam(ctx, exam.ID, orgID, dto.UpdateExamRequest{AssessmentType: &rawAssessmentType})
			require.NoError(t, err)
			assert.Equal(t, "other", updated.AssessmentType)
			reloaded, readErr := examRepo.FindByID(ctx, exam.ID, orgID)
			require.NoError(t, readErr)
			assert.Equal(t, "other", reloaded.AssessmentType)
		})
	}
}

type pauseAfterFirstExamReadRepository struct {
	*GORMExamRepository
	staleRead chan<- struct{}
	release   <-chan struct{}
	once      sync.Once
}

func (r *pauseAfterFirstExamReadRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*entities.Exam, error) {
	exam, err := r.GORMExamRepository.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	r.once.Do(func() {
		close(r.staleRead)
		select {
		case <-r.release:
		case <-ctx.Done():
		}
	})
	return exam, nil
}

func academicIntegrationQuestion(examID uuid.UUID, text string, now time.Time) entities.ExamQuestion {
	questionID := uuid.New()
	return entities.ExamQuestion{
		ID: questionID, ExamID: examID, Question: text, Position: 1, CreatedAt: now, UpdatedAt: now,
		Answers: []entities.ExamAnswer{
			{ID: uuid.New(), QuestionID: questionID, Answer: "Correct", IsCorrect: true, CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), QuestionID: questionID, Answer: "Incorrect", IsCorrect: false, CreatedAt: now, UpdatedAt: now},
		},
	}
}

func academicIntegrationExam(orgID, courseID uuid.UUID, assessmentType string, now time.Time) *entities.Exam {
	return &entities.Exam{
		ID: uuid.New(), CourseID: courseID, OrgID: orgID, AssessmentType: assessmentType, Title: "Academic integration exam",
		DurationMinutes: 60, MaxAttempts: 1, FileEnabled: true, FilePoints: 100, CreatedAt: now, UpdatedAt: now,
	}
}
