package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGORMCourseRepository_CreateWithAcademicSetup_RollsBackOnAssessmentFailure(t *testing.T) {
	db := openAcademicIntegrationDB(t)
	repo := NewGORMCourseRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	orgID, courseID, teacherID, examID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	// The pre-existing primary key makes the final assessment insert fail. The
	// test then verifies that neither the new course nor its teacher assignment
	// survived the surrounding transaction.
	require.NoError(t, db.Create(&database.ExamModel{
		ID: examID.String(), CourseID: uuid.New().String(), OrgID: orgID.String(), AssessmentType: "other",
		Title: "conflict", DurationMinutes: 60, MaxAttempts: 1, FileEnabled: true, FilePoints: 1, CreatedAt: now, UpdatedAt: now,
	}).Error)
	t.Cleanup(func() {
		_ = db.Where("id = ?", examID.String()).Delete(&database.ExamModel{}).Error
		_ = db.Where("course_id = ?", courseID.String()).Delete(&database.CourseTeacherModel{}).Error
		_ = db.Where("id = ?", courseID.String()).Delete(&database.CourseModel{}).Error
	})

	course := &entities.Course{ID: courseID, OrgID: orgID, Title: "Atomic course", Status: entities.StatusDraft, CreatedBy: teacherID, CreatedAt: now, UpdatedAt: now}
	teacher := &entities.CourseTeacher{ID: uuid.New(), CourseID: courseID, TeacherID: teacherID, OrgID: orgID, AssignedAt: now}
	exam := &entities.Exam{ID: examID, CourseID: courseID, OrgID: orgID, AssessmentType: "midterm", Title: "Midterm", DurationMinutes: 60, MaxAttempts: 1, FileEnabled: true, FilePoints: 100, CreatedAt: now, UpdatedAt: now}

	err := repo.CreateWithAcademicSetup(ctx, course, teacher, []*entities.Exam{exam})
	require.Error(t, err)

	var courseCount, teacherCount int64
	require.NoError(t, db.Model(&database.CourseModel{}).Where("id = ?", courseID.String()).Count(&courseCount).Error)
	require.NoError(t, db.Model(&database.CourseTeacherModel{}).Where("course_id = ?", courseID.String()).Count(&teacherCount).Error)
	require.Zero(t, courseCount)
	require.Zero(t, teacherCount)
}
