package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTeacherUseCase_ListCalendar_OnlyIncludesAssignedCourseEvents(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	groupRepo := new(mocks.MockGroupRepository)
	courseTeacherRepo := new(mocks.MockCourseTeacherRepository)
	userRepo := new(mocks.MockUserRepository)
	assignmentRepo := new(mocks.MockAssignmentRepository)
	examRepo := new(mocks.MockExamRepository)
	uc := NewTeacherUseCase(courseRepo, groupRepo, courseTeacherRepo, userRepo, assignmentRepo, examRepo)

	orgID, teacherID, courseID, groupID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	assignmentID, examID, scheduleID := uuid.New(), uuid.New(), uuid.New()
	course := entities.Course{ID: courseID, OrgID: orgID, Title: "Distributed Systems"}
	group := entities.Group{ID: groupID, OrgID: orgID, Name: "DS-01"}
	from := time.Date(2026, time.October, 5, 0, 0, 0, 0, time.UTC) // Monday
	to := time.Date(2026, time.October, 11, 0, 0, 0, 0, time.UTC)  // Sunday
	assignmentDeadline := time.Date(2026, time.October, 8, 14, 0, 0, 0, time.UTC)
	examDeadline := time.Date(2026, time.October, 6, 9, 0, 0, 0, time.UTC)
	outOfRangeDeadline := time.Date(2026, time.October, 12, 9, 0, 0, 0, time.UTC)

	courseTeacherRepo.On("FindCourseIDsByTeacher", mock.Anything, teacherID, orgID).Return([]uuid.UUID{courseID}, nil)
	courseRepo.On("FindByIDs", mock.Anything, []uuid.UUID{courseID}, orgID, 1, 1).Return([]entities.Course{course}, int64(1), nil)
	assignmentRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Assignment{
		{ID: assignmentID, CourseID: courseID, OrgID: orgID, Title: "Lab report", DueDate: &assignmentDeadline},
		{ID: uuid.New(), CourseID: courseID, OrgID: orgID, Title: "Later task", DueDate: &outOfRangeDeadline},
	}, nil)
	examRepo.On("FindByCourse", mock.Anything, courseID, orgID).Return([]*entities.Exam{
		{ID: examID, CourseID: courseID, OrgID: orgID, Title: "Midterm", AssessmentType: "midterm", DueDate: &examDeadline},
	}, nil)
	groupRepo.On("ListGroupsByCourse", mock.Anything, courseID, orgID).Return([]entities.Group{group}, nil)
	groupRepo.On("ListSchedulesByGroup", mock.Anything, groupID).Return([]entities.GroupSchedule{
		{ID: scheduleID, GroupID: groupID, DayOfWeek: int(time.Monday), StartTime: "09:00", EndTime: "10:30", Location: "Room 204"},
		{ID: uuid.New(), GroupID: groupID, DayOfWeek: int(time.Sunday), StartTime: "13:00", EndTime: "14:00"},
	}, nil)

	events, err := uc.ListCalendar(context.Background(), teacherID, orgID, from, to)
	require.NoError(t, err)
	require.Len(t, events, 4)

	assert.Equal(t, "class", events[0].ItemType)
	assert.Equal(t, "2026-10-05T09:00:00Z", events[0].StartAt.Format(time.RFC3339))
	assert.Equal(t, "exam", events[1].ItemType)
	assert.Equal(t, "midterm", events[1].AssessmentType)
	assert.Equal(t, "assignment", events[2].ItemType)
	assert.Equal(t, assignmentID.String(), events[2].ItemID)
	assert.Equal(t, "class", events[3].ItemType)
	assert.Equal(t, "2026-10-11T13:00:00Z", events[3].StartAt.Format(time.RFC3339))
	assert.Equal(t, "DS-01", *events[0].GroupName)
	assert.Equal(t, "Room 204", events[0].Location)
	courseTeacherRepo.AssertExpectations(t)
	courseRepo.AssertExpectations(t)
	assignmentRepo.AssertExpectations(t)
	examRepo.AssertExpectations(t)
	groupRepo.AssertExpectations(t)
}

func TestTeacherUseCase_ListCalendar_NoAssignedCourses(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	groupRepo := new(mocks.MockGroupRepository)
	courseTeacherRepo := new(mocks.MockCourseTeacherRepository)
	userRepo := new(mocks.MockUserRepository)
	assignmentRepo := new(mocks.MockAssignmentRepository)
	examRepo := new(mocks.MockExamRepository)
	uc := NewTeacherUseCase(courseRepo, groupRepo, courseTeacherRepo, userRepo, assignmentRepo, examRepo)

	orgID, teacherID := uuid.New(), uuid.New()
	courseTeacherRepo.On("FindCourseIDsByTeacher", mock.Anything, teacherID, orgID).Return([]uuid.UUID{}, nil)

	events, err := uc.ListCalendar(context.Background(), teacherID, orgID, time.Now(), time.Now())
	require.NoError(t, err)
	assert.Empty(t, events)
	courseTeacherRepo.AssertExpectations(t)
}
