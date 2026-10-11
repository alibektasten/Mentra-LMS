package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/infrastructure/storage"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExamHandler struct {
	examUC  *usecases.ExamUseCase
	storage *storage.LocalStorage
}

func NewExamHandler(examUC *usecases.ExamUseCase, storage *storage.LocalStorage) *ExamHandler {
	return &ExamHandler{examUC: examUC, storage: storage}
}

func (h *ExamHandler) List(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.examUC.ListExams(c.Request.Context(), courseID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExamHandler) Create(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	var req dto.CreateExamRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.examUC.CreateExam(c.Request.Context(), courseID, orgID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *ExamHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid exam id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.examUC.GetExam(c.Request.Context(), id, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExamHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid exam id"))
		return
	}
	var req dto.UpdateExamRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.examUC.UpdateExam(c.Request.Context(), id, orgID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExamHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid exam id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if err := h.examUC.DeleteExam(c.Request.Context(), id, orgID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *ExamHandler) StartAttempt(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid exam id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	studentID, _ := uuid.Parse(middleware.GetUserID(c))
	result, err := h.examUC.StartAttempt(c.Request.Context(), examID, studentID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExamHandler) SubmitAttempt(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, storage.MaxUploadBytes+1024*1024)
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid attempt id"))
		return
	}
	studentID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))

	// Parse MCQ answers from form field
	var mcqAnswers []dto.ExamMCQAnswerInput
	mcqJSON := c.PostForm("mcq_answers")
	if mcqJSON != "" {
		if err := json.Unmarshal([]byte(mcqJSON), &mcqAnswers); err != nil {
			c.Error(apperrors.ValidationError("invalid mcq_answers format"))
			return
		}
	}

	// Handle optional file upload
	var filePath string
	file, header, fileErr := c.Request.FormFile("file")
	if fileErr == nil {
		defer file.Close()
		filePath, err = h.storage.Save(file, header, orgID.String())
		if err != nil {
			if errors.Is(err, storage.ErrInvalidUpload) {
				c.Error(apperrors.ValidationError(err.Error()))
				return
			}
			c.Error(apperrors.InternalError("failed to save file: " + err.Error()))
			return
		}
	}

	result, err := h.examUC.SubmitAttempt(c.Request.Context(), attemptID, studentID, mcqAnswers, filePath)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExamHandler) SaveDraft(c *gin.Context) {
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid attempt id"))
		return
	}
	var req dto.SaveExamDraftRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	studentID, _ := uuid.Parse(middleware.GetUserID(c))
	if err := h.examUC.SaveDraft(c.Request.Context(), attemptID, studentID, req.MCQAnswers); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ExamHandler) MyAttempts(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid exam id"))
		return
	}
	studentID, _ := uuid.Parse(middleware.GetUserID(c))
	result, err := h.examUC.MyAttempts(c.Request.Context(), examID, studentID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExamHandler) ListAttempts(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid exam id"))
		return
	}
	result, err := h.examUC.ListAttempts(c.Request.Context(), examID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExamHandler) GradeFile(c *gin.Context) {
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid attempt id"))
		return
	}
	var req dto.GradeExamFileRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	graderID, _ := uuid.Parse(middleware.GetUserID(c))
	result, err := h.examUC.GradeFileSection(c.Request.Context(), attemptID, graderID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExamHandler) GrantExtra(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid exam id"))
		return
	}
	var req dto.GrantExtraAttemptRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	grantedByID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if err := h.examUC.GrantExtraAttempt(c.Request.Context(), examID, grantedByID, orgID, req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"success": true}})
}
