package handlers

import (
	"net/http"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AIHandler struct {
	aiUC           *usecases.AIUseCase
	courseTeachers repositories.CourseTeacherRepository
	groups         repositories.GroupRepository
	courses        repositories.CourseRepository
	lessons        repositories.LessonRepository
	modules        repositories.ModuleRepository
}

func NewAIHandler(
	aiUC *usecases.AIUseCase,
	courseTeachers repositories.CourseTeacherRepository,
	groups repositories.GroupRepository,
	courses repositories.CourseRepository,
	lessons repositories.LessonRepository,
	modules repositories.ModuleRepository,
) *AIHandler {
	return &AIHandler{
		aiUC:           aiUC,
		courseTeachers: courseTeachers,
		groups:         groups,
		courses:        courses,
		lessons:        lessons,
		modules:        modules,
	}
}

func (h *AIHandler) authorizeLesson(c *gin.Context, lessonID uuid.UUID) bool {
	return middleware.AuthorizeLessonAccess(c, h.courseTeachers, h.groups, h.courses, h.lessons, h.modules, lessonID)
}

func (h *AIHandler) SummarizeLesson(c *gin.Context) {
	var req dto.SummarizeLessonRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	lessonID, err := uuid.Parse(req.LessonID)
	if err != nil {
		c.Error(apperrors.ValidationError("invalid lesson id"))
		return
	}
	if !h.authorizeLesson(c, lessonID) {
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	resp, err := h.aiUC.SummarizeLesson(c.Request.Context(), lessonID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *AIHandler) GenerateQuiz(c *gin.Context) {
	var req dto.GenerateQuizRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	lessonID, err := uuid.Parse(req.LessonID)
	if err != nil {
		c.Error(apperrors.ValidationError("invalid lesson id"))
		return
	}
	if !h.authorizeLesson(c, lessonID) {
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	quiz, err := h.aiUC.GenerateQuiz(c.Request.Context(), lessonID, orgID, req.NumQuestions)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": quiz})
}

func (h *AIHandler) GetAssignmentFeedback(c *gin.Context) {
	var req dto.AssignmentFeedbackRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	submissionID, err := uuid.Parse(req.SubmissionID)
	if err != nil {
		c.Error(apperrors.ValidationError("invalid submission id"))
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	resp, err := h.aiUC.GetAssignmentFeedback(c.Request.Context(), submissionID, userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *AIHandler) GenerateFlashcards(c *gin.Context) {
	var req dto.GenerateFlashcardsRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	lessonID, err := uuid.Parse(req.LessonID)
	if err != nil {
		c.Error(apperrors.ValidationError("invalid lesson id"))
		return
	}
	if !h.authorizeLesson(c, lessonID) {
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	flashcards, err := h.aiUC.GenerateFlashcards(c.Request.Context(), lessonID, orgID, req.NumCards)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": flashcards})
}
