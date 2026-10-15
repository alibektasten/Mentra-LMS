package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAttendanceUseCase_SaveSheet_RequiresEveryEnrolledStudent(t *testing.T) {
	attendanceRepo := new(mocks.MockAttendanceRepository)
	groupRepo := new(mocks.MockGroupRepository)
	userRepo := new(mocks.MockUserRepository)
	uc := NewAttendanceUseCase(attendanceRepo, groupRepo, userRepo)

	orgID, courseID, groupID := uuid.New(), uuid.New(), uuid.New()
	studentA, studentB := uuid.New(), uuid.New()
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(&entities.Group{ID: groupID, OrgID: orgID, CourseID: &courseID}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: studentA}, {StudentID: studentB}}, nil)

	_, err := uc.SaveSheet(context.Background(), courseID, groupID, orgID, uuid.New(), dto.SaveAttendanceRequest{
		Date: "2026-10-15", Records: []dto.AttendanceMarkInput{{StudentID: studentA.String(), Status: "present"}},
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.HTTPStatus)
	attendanceRepo.AssertNotCalled(t, "UpsertMany", mock.Anything, mock.Anything)
}

func TestAttendanceUseCase_GetSheet_ScopesRecordsToTheCourseGroup(t *testing.T) {
	attendanceRepo := new(mocks.MockAttendanceRepository)
	groupRepo := new(mocks.MockGroupRepository)
	userRepo := new(mocks.MockUserRepository)
	uc := NewAttendanceUseCase(attendanceRepo, groupRepo, userRepo)

	orgID, courseID, groupID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	date := time.Date(2026, 10, 15, 12, 0, 0, 0, time.FixedZone("UTC+6", 6*3600))
	groupRepo.On("GetGroupByID", mock.Anything, groupID, orgID).Return(&entities.Group{ID: groupID, OrgID: orgID, CourseID: &courseID}, nil)
	groupRepo.On("ListMembers", mock.Anything, groupID).Return([]entities.GroupMember{{StudentID: studentID}}, nil)
	userRepo.On("FindByIDs", mock.Anything, []uuid.UUID{studentID}).Return([]entities.User{{ID: studentID, Name: "Aida", Email: "aida@example.com"}}, nil)
	expectedDate := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	attendanceRepo.On("ListByGroupAndDate", mock.Anything, groupID, expectedDate).Return([]entities.AttendanceRecord{{StudentID: studentID, Status: entities.AttendanceLate}}, nil)

	result, err := uc.GetSheet(context.Background(), courseID, groupID, orgID, date)
	require.NoError(t, err)
	require.Len(t, result.Students, 1)
	assert.Equal(t, "late", result.Students[0].Status)
}
