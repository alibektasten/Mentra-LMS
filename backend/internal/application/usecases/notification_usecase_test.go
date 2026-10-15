package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNotificationUseCase_ListStudentIncludesCourseEvents(t *testing.T) {
	studentID, orgID, courseID, moduleID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	dueAt := now.Add(48 * time.Hour)

	announcementRepo := new(mocks.MockAnnouncementRepository)
	courseRepo := new(mocks.MockCourseRepository)
	groupRepo := new(mocks.MockGroupRepository)
	teacherRepo := new(mocks.MockCourseTeacherRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	lessonRepo := new(mocks.MockLessonRepository)
	assignmentRepo := new(mocks.MockAssignmentRepository)
	examRepo := new(mocks.MockExamRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)

	course := &entities.Course{ID: courseID, OrgID: orgID, Title: "Biology"}
	groupRepo.On("FindCourseIDsByStudent", mock.Anything, studentID, orgID).Return([]uuid.UUID{courseID}, nil)
	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(course, nil)
	announcementRepo.On("ListByCourse", mock.Anything, orgID, courseID, 10, 0).Return([]entities.Announcement{}, int64(0), nil)
	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Assignment{{ID: uuid.New(), Title: "Lab report", DueDate: &dueAt, CreatedAt: now}}, nil)
	examRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Exam{{ID: uuid.New(), Title: "Midterm", DueDate: &dueAt, CreatedAt: now}}, nil)
	moduleRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]entities.Module{{ID: moduleID}}, nil)
	lessonRepo.On("FindByModule", mock.Anything, moduleID, orgID).Return([]entities.Lesson{{ID: uuid.New(), Title: "Cells", CreatedAt: now}}, nil)
	gradeRepo.On("Find", mock.Anything, courseID, studentID, orgID).Return(&entities.CourseGrade{ID: uuid.New(), PublishedAt: &now}, nil)

	uc := NewNotificationUseCase(announcementRepo, courseRepo, groupRepo, teacherRepo, moduleRepo, lessonRepo, assignmentRepo, examRepo, gradeRepo)
	items, err := uc.List(context.Background(), studentID, orgID, string(entities.RoleStudent))

	require.NoError(t, err)
	assert.Len(t, items, 4)
	assert.Contains(t, notificationTypes(items), "assignment")
	assert.Contains(t, notificationTypes(items), "exam")
	assert.Contains(t, notificationTypes(items), "lesson")
	assert.Contains(t, notificationTypes(items), "grade")
}

func TestNotificationUseCase_ListStudentSkipsUnscheduledAssessmentShell(t *testing.T) {
	studentID, orgID, courseID := uuid.New(), uuid.New(), uuid.New()
	announcementRepo := new(mocks.MockAnnouncementRepository)
	courseRepo := new(mocks.MockCourseRepository)
	groupRepo := new(mocks.MockGroupRepository)
	teacherRepo := new(mocks.MockCourseTeacherRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	lessonRepo := new(mocks.MockLessonRepository)
	assignmentRepo := new(mocks.MockAssignmentRepository)
	examRepo := new(mocks.MockExamRepository)
	gradeRepo := new(mocks.MockCourseGradeRepository)

	groupRepo.On("FindCourseIDsByStudent", mock.Anything, studentID, orgID).Return([]uuid.UUID{courseID}, nil)
	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(&entities.Course{ID: courseID, OrgID: orgID}, nil)
	announcementRepo.On("ListByCourse", mock.Anything, orgID, courseID, 10, 0).Return([]entities.Announcement{}, int64(0), nil)
	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Assignment{}, nil)
	examRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Exam{{ID: uuid.New(), Title: "Final"}}, nil)
	moduleRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]entities.Module{}, nil)
	gradeRepo.On("Find", mock.Anything, courseID, studentID, orgID).Return(nil, nil)

	uc := NewNotificationUseCase(announcementRepo, courseRepo, groupRepo, teacherRepo, moduleRepo, lessonRepo, assignmentRepo, examRepo, gradeRepo)
	items, err := uc.List(context.Background(), studentID, orgID, string(entities.RoleStudent))

	require.NoError(t, err)
	assert.Empty(t, items)
}

func notificationTypes(items []dto.NotificationDTO) []string {
	types := make([]string, 0, len(items))
	for _, item := range items {
		types = append(types, item.Type)
	}
	return types
}
