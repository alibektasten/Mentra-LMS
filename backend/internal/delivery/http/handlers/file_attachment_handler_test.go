package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/storage"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFileAttachmentHandler_ListByRef_RestrictsStudentToEnrolledCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	studentID := uuid.New()
	orgID := uuid.New()
	assignmentID := uuid.New()
	courseID := uuid.New()

	newRouter := func(enrolled bool) (*gin.Engine, *mocks.MockFileAttachmentRepository) {
		attachments := new(mocks.MockFileAttachmentRepository)
		courseTeachers := new(mocks.MockCourseTeacherRepository)
		groups := new(mocks.MockGroupRepository)
		courses := new(mocks.MockCourseRepository)
		lessons := new(mocks.MockLessonRepository)
		modules := new(mocks.MockModuleRepository)
		assignments := new(mocks.MockAssignmentRepository)

		assignments.On("FindByID", mock.Anything, assignmentID, orgID).
			Return(&entities.Assignment{ID: assignmentID, CourseID: courseID, OrgID: orgID}, nil).Once()
		if enrolled {
			groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).
				Return(&entities.Group{ID: uuid.New(), CourseID: &courseID, OrgID: orgID}, nil).Once()
			courses.On("FindByID", mock.Anything, courseID, orgID).
				Return(&entities.Course{ID: courseID, OrgID: orgID, Status: entities.StatusPublished}, nil).Once()
			attachments.On("FindByRef", mock.Anything, "assignment", assignmentID).
				Return([]*entities.FileAttachment{}, nil).Once()
		} else {
			groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).
				Return(nil, apperrors.NotFoundError("group", studentID.String())).Once()
		}

		handler := NewFileAttachmentHandler(
			usecases.NewFileAttachmentUseCase(attachments),
			courseTeachers, groups, courses, lessons, modules, assignments,
		)
		router := gin.New()
		router.Use(middleware.ErrorHandler())
		router.Use(func(c *gin.Context) {
			c.Set(string(middleware.ContextKeyUserID), studentID.String())
			c.Set(string(middleware.ContextKeyOrgID), orgID.String())
			c.Set(string(middleware.ContextKeyRole), string(entities.RoleStudent))
			c.Next()
		})
		router.GET("/attachments", handler.ListByRef)
		return router, attachments
	}

	t.Run("allows enrolled student", func(t *testing.T) {
		router, attachments := newRouter(true)
		request := httptest.NewRequest(http.MethodGet, "/attachments?ref_type=assignment&ref_id="+assignmentID.String(), nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		attachments.AssertExpectations(t)
	})

	t.Run("denies unenrolled student", func(t *testing.T) {
		router, attachments := newRouter(false)
		request := httptest.NewRequest(http.MethodGet, "/attachments?ref_type=assignment&ref_id="+assignmentID.String(), nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusForbidden, response.Code)
		attachments.AssertNotCalled(t, "FindByRef", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestFileAttachmentHandler_CreateRequiresOwnUploadClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, otherOrgID, uploaderID, otherUserID, assignmentID, courseID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	store := storage.NewLocalStorage(t.TempDir(), "attachment-claim-test")
	fileDir := filepath.Join(store.BaseDir, orgID.String())
	require.NoError(t, os.MkdirAll(fileDir, 0700))
	storedPath := "uploads/" + orgID.String() + "/teacher.pdf"
	require.NoError(t, os.WriteFile(filepath.Join(fileDir, "teacher.pdf"), []byte("%PDF-1.4"), 0600))
	claim, err := store.IssueUploadClaim(orgID.String(), uploaderID.String(), storedPath)
	require.NoError(t, err)

	request := func(userID, requestOrgID uuid.UUID, path, token string, expectedStatus int) {
		t.Helper()
		attachments := new(mocks.MockFileAttachmentRepository)
		assignments := new(mocks.MockAssignmentRepository)
		assignments.On("FindByID", mock.Anything, assignmentID, requestOrgID).
			Return(&entities.Assignment{ID: assignmentID, CourseID: courseID, OrgID: requestOrgID}, nil)
		if expectedStatus == http.StatusCreated {
			attachments.On("Create", mock.Anything, mock.MatchedBy(func(file *entities.FileAttachment) bool {
				return file.StoredPath == storedPath && file.UploaderID == uploaderID && file.OrgID == orgID
			})).Return(nil).Once()
		}
		handler := NewFileAttachmentHandler(usecases.NewFileAttachmentUseCase(attachments), nil, nil, nil, nil, nil, assignments, store)
		router := gin.New()
		router.Use(middleware.ErrorHandler())
		router.POST("/attachments", func(c *gin.Context) {
			c.Set(string(middleware.ContextKeyUserID), userID.String())
			c.Set(string(middleware.ContextKeyOrgID), requestOrgID.String())
			c.Set(string(middleware.ContextKeyRole), string(entities.RoleAdmin))
			c.Next()
		}, handler.Create)
		body, err := json.Marshal(map[string]any{"ref_type": "assignment", "ref_id": assignmentID.String(), "stored_path": path, "upload_token": token, "original_name": "teacher.pdf"})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/attachments", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		assert.Equal(t, expectedStatus, res.Code)
		if expectedStatus != http.StatusCreated {
			attachments.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		} else {
			attachments.AssertExpectations(t)
		}
	}

	request(uploaderID, orgID, storedPath, claim, http.StatusCreated)
	request(otherUserID, orgID, storedPath, claim, http.StatusForbidden)
	request(uploaderID, otherOrgID, storedPath, claim, http.StatusForbidden)
	request(uploaderID, orgID, storedPath, "", http.StatusForbidden)
}

func TestUploadThenAttachUsesReturnedClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, userID, assignmentID, courseID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	store := storage.NewLocalStorage(t.TempDir(), "upload-flow-test")
	attachments := new(mocks.MockFileAttachmentRepository)
	assignments := new(mocks.MockAssignmentRepository)
	assignments.On("FindByID", mock.Anything, assignmentID, orgID).Return(&entities.Assignment{ID: assignmentID, OrgID: orgID, CourseID: courseID}, nil)
	attachments.On("Create", mock.Anything, mock.MatchedBy(func(file *entities.FileAttachment) bool {
		return file.UploaderID == userID && file.OrgID == orgID && strings.HasPrefix(file.StoredPath, "uploads/"+orgID.String()+"/")
	})).Return(nil).Once()
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUserID), userID.String())
		c.Set(string(middleware.ContextKeyOrgID), orgID.String())
		c.Set(string(middleware.ContextKeyRole), string(entities.RoleAdmin))
		c.Next()
	})
	router.POST("/upload", NewUploadHandler(store).Upload)
	router.POST("/attachments", NewFileAttachmentHandler(usecases.NewFileAttachmentUseCase(attachments), nil, nil, nil, nil, nil, assignments, store).Create)

	var uploadBody bytes.Buffer
	writer := multipart.NewWriter(&uploadBody)
	part, err := writer.CreateFormFile("file", "lesson.pdf")
	require.NoError(t, err)
	_, err = part.Write([]byte("%PDF-1.4 sample"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	uploadRequest := httptest.NewRequest(http.MethodPost, "/upload", &uploadBody)
	uploadRequest.Header.Set("Content-Type", writer.FormDataContentType())
	uploadResponse := httptest.NewRecorder()
	router.ServeHTTP(uploadResponse, uploadRequest)
	require.Equal(t, http.StatusOK, uploadResponse.Code, uploadResponse.Body.String())
	var uploadResult struct {
		Data struct {
			Path        string `json:"path"`
			UploadToken string `json:"upload_token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(uploadResponse.Body.Bytes(), &uploadResult))
	require.NotEmpty(t, uploadResult.Data.UploadToken)

	attachBody, err := json.Marshal(map[string]any{"ref_type": "assignment", "ref_id": assignmentID.String(), "stored_path": uploadResult.Data.Path, "upload_token": uploadResult.Data.UploadToken, "original_name": "lesson.pdf"})
	require.NoError(t, err)
	attachRequest := httptest.NewRequest(http.MethodPost, "/attachments", bytes.NewReader(attachBody))
	attachRequest.Header.Set("Content-Type", "application/json")
	attachResponse := httptest.NewRecorder()
	router.ServeHTTP(attachResponse, attachRequest)
	assert.Equal(t, http.StatusCreated, attachResponse.Code)
	attachments.AssertExpectations(t)
}
