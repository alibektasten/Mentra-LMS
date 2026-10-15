package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/storage"
	"github.com/ailms/backend/internal/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrivateSubmissionDownloadOwnerAndOtherStudent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, ownerID, otherID, assignmentID, submissionID, courseID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	base := t.TempDir()
	orgDir := filepath.Join(base, orgID.String())
	require.NoError(t, os.MkdirAll(orgDir, 0700))
	storedPath := "uploads/" + orgID.String() + "/report.pdf"
	require.NoError(t, os.WriteFile(filepath.Join(orgDir, "report.pdf"), []byte("%PDF-1.4 private"), 0600))
	assignments := new(mocks.MockAssignmentRepository)
	assignments.On("FindSubmissionByID", mock.Anything, submissionID).Return(&entities.AssignmentSubmission{ID: submissionID, OrgID: orgID, AssignmentID: assignmentID, StudentID: ownerID, FilePath: storedPath}, nil)
	assignments.On("FindByID", mock.Anything, assignmentID, orgID).Return(&entities.Assignment{ID: assignmentID, OrgID: orgID, CourseID: courseID}, nil)
	h := NewPrivateFileHandler(storage.NewLocalStorage(base), nil, assignments, nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/files/submissions/:id", func(c *gin.Context) {
		c.Set("org_id", orgID.String())
		c.Set("user_id", c.GetHeader("X-Test-User"))
		c.Set("role", "student")
		c.Next()
	}, h.Submission)
	request := func(userID uuid.UUID) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/files/submissions/"+submissionID.String(), nil)
		req.Header.Set("X-Test-User", userID.String())
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res
	}

	owner := request(ownerID)
	assert.Equal(t, http.StatusOK, owner.Code)
	assert.Equal(t, "private, no-store", owner.Header().Get("Cache-Control"))
	assert.Equal(t, "%PDF-1.4 private", owner.Body.String())
	assert.Equal(t, http.StatusForbidden, request(otherID).Code)
}

func TestPrivateSubmissionDownloadRejectsOtherOrganization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, otherOrgID, ownerID, submissionID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	assignments := new(mocks.MockAssignmentRepository)
	assignments.On("FindSubmissionByID", mock.Anything, submissionID).Return(&entities.AssignmentSubmission{ID: submissionID, OrgID: otherOrgID, StudentID: ownerID, FilePath: "uploads/" + otherOrgID.String() + "/report.pdf"}, nil)
	h := NewPrivateFileHandler(storage.NewLocalStorage(t.TempDir()), nil, assignments, nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/files/submissions/:id", func(c *gin.Context) {
		c.Set("org_id", orgID.String())
		c.Set("user_id", ownerID.String())
		c.Set("role", "student")
		c.Next()
	}, h.Submission)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/files/submissions/"+submissionID.String(), nil))
	assert.Equal(t, http.StatusForbidden, res.Code)
}

func TestPublicAvatarCannotServeCourseUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, userID := uuid.New(), uuid.New()
	base := t.TempDir()
	orgDir := filepath.Join(base, orgID.String())
	require.NoError(t, os.MkdirAll(orgDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(orgDir, "course.png"), []byte("private course image"), 0600))
	users := new(mocks.MockUserRepository)
	users.On("FindByID", mock.Anything, userID).Return(&entities.User{ID: userID, AvatarURL: "uploads/" + orgID.String() + "/course.png"}, nil)
	h := NewPrivateFileHandler(storage.NewLocalStorage(base), nil, nil, nil, nil, nil, nil, nil, nil, nil, users)
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/avatars/:id", h.Avatar)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/avatars/"+userID.String(), nil))
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.NotContains(t, res.Body.String(), "private course image")
}

func TestPublicAvatarServesOnlyPromotedUserImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	base := t.TempDir()
	avatarDir := filepath.Join(base, "avatars", userID.String())
	require.NoError(t, os.MkdirAll(avatarDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(avatarDir, "photo.png"), []byte("public avatar"), 0644))
	users := new(mocks.MockUserRepository)
	users.On("FindByID", mock.Anything, userID).Return(&entities.User{ID: userID, AvatarURL: "avatars/" + userID.String() + "/photo.png"}, nil)
	h := NewPrivateFileHandler(storage.NewLocalStorage(base), nil, nil, nil, nil, nil, nil, nil, nil, nil, users)
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/avatars/:id", h.Avatar)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/avatars/"+userID.String(), nil))
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "public avatar", res.Body.String())
}

func TestPrivateExamAttemptRejectsAnotherStudent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, ownerID, otherID, attemptID, examID, courseID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	attempts := new(mocks.MockExamAttemptRepository)
	exams := new(mocks.MockExamRepository)
	attempts.On("FindByID", mock.Anything, attemptID).Return(&entities.ExamAttempt{ID: attemptID, OrgID: orgID, ExamID: examID, StudentID: ownerID, FilePath: "uploads/" + orgID.String() + "/answer.pdf"}, nil)
	exams.On("FindByID", mock.Anything, examID, orgID).Return(&entities.Exam{ID: examID, OrgID: orgID, CourseID: courseID}, nil)
	h := NewPrivateFileHandler(storage.NewLocalStorage(t.TempDir()), nil, nil, nil, nil, attempts, exams, nil, nil, nil, nil)
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/files/exam-attempts/:id", func(c *gin.Context) {
		c.Set("org_id", orgID.String())
		c.Set("user_id", otherID.String())
		c.Set("role", "student")
		c.Next()
	}, h.ExamAttempt)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/files/exam-attempts/"+attemptID.String(), nil))
	assert.Equal(t, http.StatusForbidden, res.Code)
}
