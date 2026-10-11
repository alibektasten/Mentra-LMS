package usecases

import (
	"context"
	"net/http"
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

func TestCourseUseCase_CreateCourse_Success_DraftStatus(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	courseTeacherRepo := new(mocks.MockCourseTeacherRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, new(mocks.MockGroupRepository), courseTeacherRepo)

	orgID := uuid.New()
	creatorID := uuid.New()

	courseRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Course")).Return(nil)
	courseTeacherRepo.On("Add", mock.Anything, mock.MatchedBy(func(assignment *entities.CourseTeacher) bool {
		return assignment.CourseID != uuid.Nil && assignment.TeacherID == creatorID && assignment.OrgID == orgID
	})).Return(nil)

	result, err := uc.CreateCourse(context.Background(), dto.CreateCourseRequest{Title: "Go Programming", Description: "Learn Go"}, creatorID, orgID)
	require.NoError(t, err)
	assert.Equal(t, "Go Programming", result.Title)
	assert.Equal(t, string(entities.StatusDraft), result.Status)
	assert.Equal(t, orgID.String(), result.OrgID)
	assert.Equal(t, creatorID.String(), result.CreatedBy)
	courseTeacherRepo.AssertExpectations(t)
}

func TestCourseUseCase_PublishCourse_Success(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, new(mocks.MockGroupRepository), new(mocks.MockCourseTeacherRepository))

	orgID := uuid.New()
	course := &entities.Course{
		ID:        uuid.New(),
		OrgID:     orgID,
		Title:     "My Course",
		Status:    entities.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	courseRepo.On("FindByID", mock.Anything, course.ID, orgID).Return(course, nil)
	courseRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Course")).Return(nil)

	err := uc.PublishCourse(context.Background(), course.ID, orgID)
	require.NoError(t, err)
	assert.Equal(t, entities.StatusPublished, course.Status)
}

func TestCourseUseCase_PublishCourse_AlreadyPublished_ReturnsConflict(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, new(mocks.MockGroupRepository), new(mocks.MockCourseTeacherRepository))

	orgID := uuid.New()
	course := &entities.Course{
		ID:     uuid.New(),
		OrgID:  orgID,
		Status: entities.StatusPublished,
	}

	courseRepo.On("FindByID", mock.Anything, course.ID, orgID).Return(course, nil)

	err := uc.PublishCourse(context.Background(), course.ID, orgID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.HTTPStatus)
}

func TestCourseUseCase_UpdateCourse_EmptyFields_PreserveExisting(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, new(mocks.MockGroupRepository), new(mocks.MockCourseTeacherRepository))

	orgID := uuid.New()
	course := &entities.Course{
		ID:          uuid.New(),
		OrgID:       orgID,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      entities.StatusDraft,
	}

	courseRepo.On("FindByID", mock.Anything, course.ID, orgID).Return(course, nil)
	courseRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Course")).Return(nil)

	// Pass empty fields — should preserve original
	result, err := uc.UpdateCourse(context.Background(), course.ID, orgID, dto.UpdateCourseRequest{Title: "", Description: ""})
	require.NoError(t, err)
	assert.Equal(t, "Original Title", result.Title)
	assert.Equal(t, "Original Description", result.Description)
}

func TestCourseUseCase_UpdateCourse_UpdatesProvidedFields(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, new(mocks.MockGroupRepository), new(mocks.MockCourseTeacherRepository))

	orgID := uuid.New()
	course := &entities.Course{
		ID:          uuid.New(),
		OrgID:       orgID,
		Title:       "Old Title",
		Description: "Old Description",
	}

	courseRepo.On("FindByID", mock.Anything, course.ID, orgID).Return(course, nil)
	courseRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Course")).Return(nil)

	result, err := uc.UpdateCourse(context.Background(), course.ID, orgID, dto.UpdateCourseRequest{Title: "New Title", Description: "New Description"})
	require.NoError(t, err)
	assert.Equal(t, "New Title", result.Title)
	assert.Equal(t, "New Description", result.Description)
}

func TestCourseUseCase_GetCourse_NotFound(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, new(mocks.MockGroupRepository), new(mocks.MockCourseTeacherRepository))

	orgID := uuid.New()
	courseID := uuid.New()
	courseRepo.On("FindByID", mock.Anything, courseID, orgID).Return(nil, apperrors.NotFoundError("course", courseID.String()))

	_, err := uc.GetCourse(context.Background(), courseID, orgID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.HTTPStatus)
}

func TestCourseUseCase_ListCoursesForStudent_ReturnsOnlyPublishedAssignedCourses(t *testing.T) {
	courseRepo := new(mocks.MockCourseRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	groupRepo := new(mocks.MockGroupRepository)
	uc := NewCourseUseCase(courseRepo, memberRepo, groupRepo, new(mocks.MockCourseTeacherRepository))

	orgID := uuid.New()
	studentID := uuid.New()
	publishedCourseID := uuid.New()
	groupRepo.On("FindCourseIDsByStudent", mock.Anything, studentID, orgID).
		Return([]uuid.UUID{publishedCourseID}, nil).Once()
	courseRepo.On("FindPublishedByIDs", mock.Anything, []uuid.UUID{publishedCourseID}, orgID, 1, 20).
		Return([]entities.Course{{ID: publishedCourseID, OrgID: orgID, Status: entities.StatusPublished}}, int64(1), nil).Once()

	courses, total, err := uc.ListCoursesForRole(context.Background(), studentID, orgID, string(entities.RoleStudent), 1, 20)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, courses, 1)
	assert.Equal(t, publishedCourseID.String(), courses[0].ID)
	assert.Equal(t, string(entities.StatusPublished), courses[0].Status)
	groupRepo.AssertExpectations(t)
	courseRepo.AssertExpectations(t)
	courseRepo.AssertNotCalled(t, "FindByIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
