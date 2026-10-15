package repositories

import (
	"context"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// These tests deliberately require an explicitly configured, separately named
// PostgreSQL database. They never fall back to DATABASE_URL, so normal local
// development data cannot be mutated by an integration-test run.
func openAcademicIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("MENTRA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("MENTRA_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	databaseName := strings.Trim(parsed.Path, "/")
	require.True(t, strings.HasSuffix(databaseName, "_test"), "MENTRA_TEST_DATABASE_URL must point to a database ending in _test")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, database.RunMigrations(db))
	return db
}

func TestGORMExamAttemptRepository_ConcurrentStartOrResume_UsesDatabaseInvariant(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMExamAttemptRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	examID, studentID, orgID := uuid.New(), uuid.New(), uuid.New()
	t.Cleanup(func() {
		_ = db.Where("exam_id = ?", examID.String()).Delete(&database.ExamAttemptModel{}).Error
	})
	now := time.Now().UTC()

	const starts = 12
	ids := make(chan uuid.UUID, starts)
	errs := make(chan error, starts)
	var wg sync.WaitGroup
	for range starts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			candidate := &entities.ExamAttempt{
				ID: uuid.New(), ExamID: examID, StudentID: studentID, OrgID: orgID, Status: "in_progress",
				StartedAt: now, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now,
			}
			attempt, err := repo.StartOrResume(ctx, candidate, 1, true, now, nil)
			if err != nil {
				errs <- err
				return
			}
			ids <- attempt.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	var first uuid.UUID
	for id := range ids {
		if first == uuid.Nil {
			first = id
		}
		require.Equal(t, first, id, "repeated starts must resume the same attempt")
	}
	count, err := repo.CountByExamAndStudent(ctx, examID, studentID)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestGORMExamAttemptRepository_ConcurrentDraftAndSubmit_CannotRevertSubmission(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMExamAttemptRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()
	attempt := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: uuid.New(), StudentID: uuid.New(), OrgID: uuid.New(), Status: "in_progress",
		StartedAt: now, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, attempt))
	t.Cleanup(func() {
		_ = db.Where("id = ?", attempt.ID.String()).Delete(&database.ExamAttemptModel{}).Error
	})

	submittedAt := time.Now().UTC()
	score := 0
	submission := *attempt
	submission.Status = "submitted"
	submission.SubmittedAt = &submittedAt
	submission.MCQScore = &score
	submission.TotalScore = &score
	submission.UpdatedAt = submittedAt

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = repo.SubmitIfInProgress(ctx, &submission)
	}()
	go func() {
		defer wg.Done()
		_, _ = repo.SaveDraftIfInProgress(ctx, attempt.ID, attempt.StudentID, []entities.ExamMCQAnswer{{QuestionID: "question", AnswerID: "answer"}}, time.Now().UTC())
	}()
	wg.Wait()

	stored, err := repo.FindByID(ctx, attempt.ID)
	require.NoError(t, err)
	require.Equal(t, "submitted", stored.Status)
	require.NotNil(t, stored.SubmittedAt)
}

func TestGORMExamAttemptRepository_RepeatedSubmit_OnlyCommitsOnce(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMExamAttemptRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	attempt := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: uuid.New(), StudentID: uuid.New(), OrgID: uuid.New(), Status: "in_progress",
		StartedAt: now, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, attempt))
	t.Cleanup(func() { _ = db.Where("id = ?", attempt.ID.String()).Delete(&database.ExamAttemptModel{}).Error })

	submittedAt := now.Add(time.Second)
	submission := *attempt
	submission.Status = "submitted"
	submission.SubmittedAt = &submittedAt
	submission.UpdatedAt = submittedAt

	first, err := repo.SubmitIfInProgress(ctx, &submission)
	require.NoError(t, err)
	second, err := repo.SubmitIfInProgress(ctx, &submission)
	require.NoError(t, err)
	require.True(t, first)
	require.False(t, second, "the CAS must reject a duplicate state transition")

	stored, err := repo.FindByID(ctx, attempt.ID)
	require.NoError(t, err)
	require.Equal(t, "submitted", stored.Status)
	require.Equal(t, submittedAt.Unix(), stored.SubmittedAt.Unix())
}

func TestGORMExamAttemptRepository_EnforcesSharedOneMinuteGracePeriod(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMExamAttemptRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	withinGrace := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: uuid.New(), StudentID: uuid.New(), OrgID: uuid.New(), Status: "in_progress",
		StartedAt: now.Add(-time.Hour), ExpiresAt: now.Add(-30 * time.Second), CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
	pastGrace := &entities.ExamAttempt{
		ID: uuid.New(), ExamID: uuid.New(), StudentID: uuid.New(), OrgID: uuid.New(), Status: "in_progress",
		StartedAt: now.Add(-time.Hour), ExpiresAt: now.Add(-61 * time.Second), CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, withinGrace))
	require.NoError(t, repo.Create(ctx, pastGrace))
	t.Cleanup(func() {
		_ = db.Where("id IN ?", []string{withinGrace.ID.String(), pastGrace.ID.String()}).Delete(&database.ExamAttemptModel{}).Error
	})

	saved, err := repo.SaveDraftIfInProgress(ctx, withinGrace.ID, withinGrace.StudentID, []entities.ExamMCQAnswer{{QuestionID: "within", AnswerID: "grace"}}, now)
	require.NoError(t, err)
	require.True(t, saved, "drafts remain valid through the one-minute grace period")

	submittedAt := now
	submission := *pastGrace
	submission.Status = "submitted"
	submission.SubmittedAt = &submittedAt
	submission.UpdatedAt = submittedAt
	submitted, err := repo.SubmitIfInProgress(ctx, &submission)
	require.NoError(t, err)
	require.False(t, submitted, "submission after the grace period must not transition the row")
	expired, err := repo.ExpireIfInProgress(ctx, pastGrace.ID, pastGrace.StudentID, now)
	require.NoError(t, err)
	require.True(t, expired)

	stored, err := repo.FindByID(ctx, pastGrace.ID)
	require.NoError(t, err)
	require.Equal(t, "expired", stored.Status)
}
