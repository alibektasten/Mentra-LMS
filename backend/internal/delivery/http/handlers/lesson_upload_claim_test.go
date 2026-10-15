package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ailms/backend/internal/application/usecases"
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

func TestLessonFileBindingRequiresUploaderClaimButAllowsUnchangedLegacyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, uploaderID, otherID, moduleID, lessonID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	store := storage.NewLocalStorage(t.TempDir(), "lesson-upload-test")
	dir := filepath.Join(store.BaseDir, orgID.String())
	require.NoError(t, os.MkdirAll(dir, 0700))
	path := "uploads/" + orgID.String() + "/new.pdf"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.pdf"), []byte("%PDF-1.4"), 0600))
	claim, err := store.IssueUploadClaim(orgID.String(), uploaderID.String(), path)
	require.NoError(t, err)

	request := func(method string, userID uuid.UUID, fileURL, token, oldURL string, expectedStatus int) {
		t.Helper()
		lessons := new(mocks.MockLessonRepository)
		modules := new(mocks.MockModuleRepository)
		if method == http.MethodPost && expectedStatus == http.StatusCreated {
			modules.On("FindByID", mock.Anything, moduleID, orgID).Return(&entities.Module{ID: moduleID, OrgID: orgID}, nil)
			lessons.On("Create", mock.Anything, mock.MatchedBy(func(lesson *entities.Lesson) bool { return lesson.FileURL == path })).Return(nil)
		}
		if method == http.MethodPut {
			lessons.On("FindByID", mock.Anything, lessonID, orgID).Return(&entities.Lesson{ID: lessonID, ModuleID: moduleID, OrgID: orgID, FileURL: oldURL}, nil)
			if expectedStatus == http.StatusOK {
				lessons.On("Update", mock.Anything, mock.MatchedBy(func(lesson *entities.Lesson) bool { return lesson.FileURL == fileURL })).Return(nil)
			}
		}
		handler := NewLessonHandler(usecases.NewLessonUseCase(lessons, modules), store)
		router := gin.New()
		router.Use(middleware.ErrorHandler())
		identity := func(c *gin.Context) {
			c.Set(string(middleware.ContextKeyOrgID), orgID.String())
			c.Set(string(middleware.ContextKeyUserID), userID.String())
			c.Next()
		}
		router.POST("/modules/:moduleID/lessons", identity, handler.Create)
		router.PUT("/modules/:moduleID/lessons/:lessonID", identity, handler.Update)
		body, err := json.Marshal(map[string]any{"title": "PDF lesson", "type": "pdf", "file_url": fileURL, "upload_token": token})
		require.NoError(t, err)
		url := "/modules/" + moduleID.String() + "/lessons"
		if method == http.MethodPut {
			url += "/" + lessonID.String()
		}
		req := httptest.NewRequest(method, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		assert.Equal(t, expectedStatus, res.Code)
		if expectedStatus == http.StatusForbidden {
			lessons.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			lessons.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		}
	}

	request(http.MethodPost, uploaderID, path, claim, "", http.StatusCreated)
	request(http.MethodPost, otherID, path, claim, "", http.StatusForbidden)
	request(http.MethodPost, uploaderID, path, "", "", http.StatusForbidden)
	for _, alias := range []string{" " + path, path + " ", "\t" + path + "\n", "./" + path, "uploads//" + orgID.String() + "/new.pdf", "uploads/" + orgID.String() + "/../" + orgID.String() + "/new.pdf"} {
		resolved, resolveErr := store.ResolvePrivatePath(orgID.String(), alias)
		require.NoError(t, resolveErr)
		assert.Equal(t, filepath.Join(dir, "new.pdf"), resolved)
		request(http.MethodPost, uploaderID, alias, "", "", http.StatusForbidden)
		request(http.MethodPut, uploaderID, alias, "", "uploads/"+orgID.String()+"/legacy.pdf", http.StatusForbidden)
	}
	request(http.MethodPost, uploaderID, " "+path+" ", claim, "", http.StatusForbidden)
	request(http.MethodPut, uploaderID, path, "", "uploads/"+orgID.String()+"/legacy.pdf", http.StatusForbidden)
	request(http.MethodPut, otherID, path, claim, "uploads/"+orgID.String()+"/legacy.pdf", http.StatusForbidden)
	request(http.MethodPut, uploaderID, path, claim, "uploads/"+orgID.String()+"/legacy.pdf", http.StatusOK)
	legacyPath := "uploads/" + orgID.String() + "/legacy.pdf"
	request(http.MethodPut, uploaderID, legacyPath, "", legacyPath, http.StatusOK)
}
