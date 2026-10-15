package handlers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/infrastructure/storage"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PrivateFileHandler struct {
	storage      *storage.LocalStorage
	attachments  repositories.FileAttachmentRepository
	assignments  repositories.AssignmentRepository
	lessons      repositories.LessonRepository
	modules      repositories.ModuleRepository
	examAttempts repositories.ExamAttemptRepository
	exams        repositories.ExamRepository
	courses      repositories.CourseRepository
	groups       repositories.GroupRepository
	teachers     repositories.CourseTeacherRepository
	users        repositories.UserRepository
}

func NewPrivateFileHandler(store *storage.LocalStorage, attachments repositories.FileAttachmentRepository, assignments repositories.AssignmentRepository, lessons repositories.LessonRepository, modules repositories.ModuleRepository, attempts repositories.ExamAttemptRepository, exams repositories.ExamRepository, courses repositories.CourseRepository, groups repositories.GroupRepository, teachers repositories.CourseTeacherRepository, users repositories.UserRepository) *PrivateFileHandler {
	return &PrivateFileHandler{storage: store, attachments: attachments, assignments: assignments, lessons: lessons, modules: modules, examAttempts: attempts, exams: exams, courses: courses, groups: groups, teachers: teachers, users: users}
}

// Avatar is intentionally public, but only the exact image selected on a
// user's profile can be served. Uploaded course files remain inaccessible.
func (h *PrivateFileHandler) Avatar(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fileError(c, apperrors.ValidationError("invalid user id"))
		return
	}
	user, err := h.users.FindByID(c.Request.Context(), id)
	if err != nil {
		fileError(c, err)
		return
	}
	filePath, err := h.storage.ResolvePublicAvatarPath(id.String(), user.AvatarURL)
	if err != nil {
		fileError(c, apperrors.NotFoundError("avatar", id.String()))
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		fileError(c, apperrors.NotFoundError("avatar", id.String()))
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		fileError(c, apperrors.NotFoundError("avatar", id.String()))
		return
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "public, max-age=300")
	http.ServeContent(c.Writer, c.Request, filepath.Base(filePath), info.ModTime(), file)
}

// UploadAvatar is the only route that can populate the public avatar tree.
func (h *PrivateFileHandler) UploadAvatar(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, storage.MaxUploadBytes+1024*1024)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fileError(c, apperrors.ValidationError("image is required"))
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
		fileError(c, apperrors.ValidationError("avatar must be PNG, JPEG or WebP"))
		return
	}
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		fileError(c, apperrors.UnauthorizedError("invalid user identity"))
		return
	}
	user, err := h.users.FindByID(c.Request.Context(), userID)
	if err != nil {
		fileError(c, err)
		return
	}
	orgID := middleware.GetOrgID(c)
	storedPath, err := h.storage.Save(file, header, orgID)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidUpload) {
			fileError(c, apperrors.ValidationError(err.Error()))
		} else {
			fileError(c, apperrors.InternalError("could not save avatar"))
		}
		return
	}
	avatarPath, err := h.storage.PromoteAvatar(orgID, userID.String(), storedPath)
	if err != nil {
		fileError(c, apperrors.InternalError("could not publish avatar"))
		return
	}
	user.AvatarURL = avatarPath
	if err := h.users.Update(c.Request.Context(), user); err != nil {
		fileError(c, apperrors.InternalError("could not update avatar"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"path": avatarPath, "url": "/api/v1/avatars/" + userID.String()}})
}

func (h *PrivateFileHandler) Lesson(c *gin.Context) {
	id, orgID, ok := fileIdentity(c)
	if !ok {
		return
	}
	lesson, err := h.lessons.FindByID(c.Request.Context(), id, orgID)
	if err != nil {
		fileError(c, err)
		return
	}
	if !middleware.AuthorizeLessonAccess(c, h.teachers, h.groups, h.courses, h.lessons, h.modules, id) {
		return
	}
	h.serve(c, orgID, lesson.FileURL, lesson.Title)
}

func (h *PrivateFileHandler) Attachment(c *gin.Context) {
	id, orgID, ok := fileIdentity(c)
	if !ok {
		return
	}
	attachment, err := h.attachments.FindByID(c.Request.Context(), id)
	if err != nil {
		fileError(c, err)
		return
	}
	if attachment.OrgID != orgID {
		fileError(c, apperrors.ForbiddenError("file belongs to another organization"))
		return
	}
	switch attachment.RefType {
	case "lesson":
		if !middleware.AuthorizeLessonAccess(c, h.teachers, h.groups, h.courses, h.lessons, h.modules, attachment.RefID) {
			return
		}
	case "assignment":
		assignment, err := h.assignments.FindByID(c.Request.Context(), attachment.RefID, orgID)
		if err != nil {
			fileError(c, err)
			return
		}
		if !middleware.AuthorizeCourseAccess(c, h.teachers, h.groups, h.courses, assignment.CourseID) {
			return
		}
	default:
		fileError(c, apperrors.ForbiddenError("unsupported attachment reference"))
		return
	}
	h.serve(c, orgID, attachment.StoredPath, attachment.OriginalName)
}

func (h *PrivateFileHandler) Submission(c *gin.Context) {
	id, orgID, ok := fileIdentity(c)
	if !ok {
		return
	}
	submission, err := h.assignments.FindSubmissionByID(c.Request.Context(), id)
	if err != nil {
		fileError(c, err)
		return
	}
	if submission.OrgID != orgID {
		fileError(c, apperrors.ForbiddenError("submission belongs to another organization"))
		return
	}
	assignment, err := h.assignments.FindByID(c.Request.Context(), submission.AssignmentID, orgID)
	if err != nil {
		fileError(c, err)
		return
	}
	if !h.ownerOrCourseStaff(c, submission.StudentID, assignment.CourseID) {
		return
	}
	h.serve(c, orgID, submission.FilePath, "submission"+filepath.Ext(submission.FilePath))
}

func (h *PrivateFileHandler) ExamAttempt(c *gin.Context) {
	id, orgID, ok := fileIdentity(c)
	if !ok {
		return
	}
	attempt, err := h.examAttempts.FindByID(c.Request.Context(), id)
	if err != nil {
		fileError(c, err)
		return
	}
	if attempt.OrgID != orgID {
		fileError(c, apperrors.ForbiddenError("attempt belongs to another organization"))
		return
	}
	exam, err := h.exams.FindByID(c.Request.Context(), attempt.ExamID, orgID)
	if err != nil {
		fileError(c, err)
		return
	}
	if !h.ownerOrCourseStaff(c, attempt.StudentID, exam.CourseID) {
		return
	}
	h.serve(c, orgID, attempt.FilePath, "exam-attempt"+filepath.Ext(attempt.FilePath))
}

func (h *PrivateFileHandler) ownerOrCourseStaff(c *gin.Context, ownerID, courseID uuid.UUID) bool {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err == nil && userID == ownerID {
		return true
	}
	switch middleware.GetRole(c) {
	case string(entities.RoleAdmin), string(entities.RoleTeacher):
		return middleware.AuthorizeCourseAccess(c, h.teachers, h.groups, h.courses, courseID)
	default:
		fileError(c, apperrors.ForbiddenError("file belongs to another student"))
		return false
	}
}

func (h *PrivateFileHandler) serve(c *gin.Context, orgID uuid.UUID, storedPath, downloadName string) {
	if storedPath == "" {
		fileError(c, apperrors.NotFoundError("file", "missing"))
		return
	}
	filePath, err := h.storage.ResolvePrivatePath(orgID.String(), storedPath)
	if err != nil {
		fileError(c, apperrors.ForbiddenError("invalid file path"))
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		fileError(c, apperrors.NotFoundError("file", "missing"))
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		fileError(c, apperrors.NotFoundError("file", "missing"))
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.Header("Content-Disposition", "inline; filename=\""+safeDownloadName(downloadName)+"\"")
	http.ServeContent(c.Writer, c.Request, downloadName, info.ModTime(), file)
}

func safeDownloadName(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r == '"' || r == '\\' || r < 32 {
			return '_'
		}
		return r
	}, name)
	if name == "." || name == "" {
		return "download"
	}
	return name
}

func fileIdentity(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fileError(c, apperrors.ValidationError("invalid file resource id"))
		return uuid.Nil, uuid.Nil, false
	}
	orgID, err := uuid.Parse(middleware.GetOrgID(c))
	if err != nil || orgID == uuid.Nil {
		fileError(c, apperrors.UnauthorizedError("invalid organization identity"))
		return uuid.Nil, uuid.Nil, false
	}
	return id, orgID, true
}

func fileError(c *gin.Context, err error) { c.Error(err); c.Abort() }
