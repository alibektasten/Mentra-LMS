package repositories

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGORMCourseGradeRepository_SetPublication_RollsBackBatchOnDatabaseError(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMCourseGradeRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	orgID, courseID := uuid.New(), uuid.New()
	conflictingID := uuid.New()

	// Reserve an unrelated primary key. The second row in the attempted
	// publication batch will fail, so the first row must be rolled back too.
	require.NoError(t, db.Create(&database.CourseGradeModel{
		ID: conflictingID.String(), CourseID: uuid.New().String(), StudentID: uuid.New().String(), OrgID: orgID.String(), CreatedAt: now, UpdatedAt: now,
	}).Error)
	t.Cleanup(func() {
		_ = db.Where("id = ?", conflictingID.String()).Delete(&database.CourseGradeModel{}).Error
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeModel{}).Error
	})

	first := &entities.CourseGrade{ID: uuid.New(), CourseID: courseID, StudentID: uuid.New(), OrgID: orgID, PublishedAt: &now, CreatedAt: now, UpdatedAt: now}
	second := &entities.CourseGrade{ID: conflictingID, CourseID: courseID, StudentID: uuid.New(), OrgID: orgID, PublishedAt: &now, CreatedAt: now, UpdatedAt: now}
	err := repo.SetPublication(ctx, []*entities.CourseGrade{first, second})
	require.Error(t, err)

	var firstCount int64
	require.NoError(t, db.Model(&database.CourseGradeModel{}).
		Where("course_id = ? AND student_id = ?", courseID.String(), first.StudentID.String()).Count(&firstCount).Error)
	assert.Zero(t, firstCount, "a failed publication must not leave the first student published")
}

func TestGORMCourseGradeRepository_ConcurrentOverrideAndPublishPreserveBothStates(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMCourseGradeRepository(db)
	ctx := context.Background()
	orgID, courseID, studentID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	t.Cleanup(func() {
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeChangeModel{}).Error
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeModel{}).Error
	})

	grade := 88.0
	note := "Published correction"
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := repo.UpdateOverride(ctx, courseID, studentID, orgID, actorID, &grade, &note, false, now)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		<-start
		errs <- repo.SetPublication(ctx, []*entities.CourseGrade{{
			ID: uuid.New(), CourseID: courseID, StudentID: studentID, OrgID: orgID,
			PublishedAt: &now, PublishedBy: &actorID, PublicationChangedAt: &now, PublicationChangedBy: &actorID, CreatedAt: now, UpdatedAt: now,
		}})
	}()
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	stored, err := repo.Find(ctx, courseID, studentID, orgID)
	require.NoError(t, err)
	require.NotNil(t, stored.OverrideGrade)
	assert.Equal(t, 88.0, *stored.OverrideGrade)
	require.NotNil(t, stored.PublishedAt)
	changes := courseGradeChanges(t, db, courseID, studentID)
	require.Len(t, changes, 1)
	assert.Nil(t, changes[0].PreviousOverride)
	require.NotNil(t, changes[0].NewOverride)
	assert.Equal(t, 88.0, *changes[0].NewOverride)
}

func TestGORMCourseGradeRepository_ConcurrentOverrideAndWithdrawPreserveOverride(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMCourseGradeRepository(db)
	ctx := context.Background()
	orgID, courseID, studentID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	previous, replacement := 55.0, 77.0
	require.NoError(t, db.Create(&database.CourseGradeModel{
		ID: uuid.NewString(), CourseID: courseID.String(), StudentID: studentID.String(), OrgID: orgID.String(), OverrideGrade: &previous, OverrideNote: "Initial correction",
		PublishedAt: &now, PublishedBy: stringPointer(actorID.String()), PublicationChangedAt: &now, PublicationChangedBy: stringPointer(actorID.String()), CreatedAt: now, UpdatedAt: now,
	}).Error)
	t.Cleanup(func() {
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeChangeModel{}).Error
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeModel{}).Error
	})

	note := "Replacement correction"
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := repo.UpdateOverride(ctx, courseID, studentID, orgID, actorID, &replacement, &note, false, now.Add(time.Second))
		errs <- err
	}()
	go func() {
		defer wg.Done()
		<-start
		errs <- repo.SetPublication(ctx, []*entities.CourseGrade{{
			ID: uuid.New(), CourseID: courseID, StudentID: studentID, OrgID: orgID,
			PublicationChangedAt: &now, PublicationChangedBy: &actorID, CreatedAt: now, UpdatedAt: now,
		}})
	}()
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	stored, err := repo.Find(ctx, courseID, studentID, orgID)
	require.NoError(t, err)
	require.NotNil(t, stored.OverrideGrade)
	assert.Equal(t, 77.0, *stored.OverrideGrade)
	assert.Nil(t, stored.PublishedAt)
	changes := courseGradeChanges(t, db, courseID, studentID)
	require.Len(t, changes, 1)
	require.NotNil(t, changes[0].PreviousOverride)
	require.NotNil(t, changes[0].NewOverride)
	assert.Equal(t, 55.0, *changes[0].PreviousOverride)
	assert.Equal(t, 77.0, *changes[0].NewOverride)
}

func TestGORMCourseGradeRepository_ConcurrentFirstOverridesFormExactAuditChain(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMCourseGradeRepository(db)
	ctx := context.Background()
	orgID, courseID, studentID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	t.Cleanup(func() {
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeChangeModel{}).Error
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeModel{}).Error
	})

	values := []float64{61, 84}
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, len(values))
	for index, value := range values {
		value := value
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			note := "Concurrent approved correction"
			_, err := repo.UpdateOverride(ctx, courseID, studentID, orgID, uuid.New(), &value, &note, false, now.Add(time.Duration(index)*time.Second))
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	changes := courseGradeChanges(t, db, courseID, studentID)
	require.Len(t, changes, 2)
	var first, second *database.CourseGradeChangeModel
	for index := range changes {
		change := &changes[index]
		if change.PreviousOverride == nil {
			first = change
		} else {
			second = change
		}
	}
	require.NotNil(t, first)
	require.NotNil(t, second)
	require.NotNil(t, first.NewOverride)
	require.NotNil(t, second.PreviousOverride)
	require.NotNil(t, second.NewOverride)
	assert.Equal(t, *first.NewOverride, *second.PreviousOverride)
	assert.Equal(t, "override_set", first.Action)
	assert.Equal(t, "override_updated", second.Action)
	stored, err := repo.Find(ctx, courseID, studentID, orgID)
	require.NoError(t, err)
	require.NotNil(t, stored.OverrideGrade)
	assert.Equal(t, *second.NewOverride, *stored.OverrideGrade)
}

func TestGORMCourseGradeRepository_AuditFailureRollsBackOverrideTransition(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMCourseGradeRepository(db)
	ctx := context.Background()
	orgID, courseID, studentID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	initial, replacement := 50.0, 90.0
	require.NoError(t, db.Create(&database.CourseGradeModel{
		ID: uuid.NewString(), CourseID: courseID.String(), StudentID: studentID.String(), OrgID: orgID.String(), OverrideGrade: &initial, OverrideNote: "Initial correction", CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Exec(`ALTER TABLE course_grade_changes ADD CONSTRAINT course_grade_changes_reject_reason CHECK (reason <> 'reject-audit')`).Error)
	t.Cleanup(func() {
		_ = db.Exec(`ALTER TABLE course_grade_changes DROP CONSTRAINT IF EXISTS course_grade_changes_reject_reason`).Error
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeChangeModel{}).Error
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseGradeModel{}).Error
	})
	note := "reject-audit"
	_, err := repo.UpdateOverride(ctx, courseID, studentID, orgID, actorID, &replacement, &note, false, now.Add(time.Second))
	require.Error(t, err)

	stored, err := repo.Find(ctx, courseID, studentID, orgID)
	require.NoError(t, err)
	require.NotNil(t, stored.OverrideGrade)
	assert.Equal(t, 50.0, *stored.OverrideGrade)
	assert.Empty(t, courseGradeChanges(t, db, courseID, studentID))
}

func courseGradeChanges(t *testing.T, db *gorm.DB, courseID, studentID uuid.UUID) []database.CourseGradeChangeModel {
	t.Helper()
	var changes []database.CourseGradeChangeModel
	require.NoError(t, db.Where("course_id = ? AND student_id = ?", courseID.String(), studentID.String()).Find(&changes).Error)
	return changes
}

func stringPointer(value string) *string { return &value }
