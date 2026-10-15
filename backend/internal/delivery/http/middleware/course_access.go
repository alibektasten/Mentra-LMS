package middleware

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireAssignedCourse limits course-scoped routes to resources the current
// tenant role may access. Teachers must be assigned to the course; students
// must belong to a group assigned to a published course. Administrators retain
// organization-wide access.
func RequireAssignedCourse(
	courseTeachers repositories.CourseTeacherRepository,
	groups repositories.GroupRepository,
	courses repositories.CourseRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseIDRaw := c.Param("id")
		if courseIDRaw == "" {
			// The collection endpoint has no course identifier to authorize.
			c.Next()
			return
		}

		courseID, ok := requiredUUID(c, courseIDRaw, "course id")
		if !ok || !AuthorizeCourseAccess(c, courseTeachers, groups, courses, courseID) {
			return
		}
		c.Next()
	}
}

// RequireOptionalProgressCourseAccess keeps the legacy unfiltered progress
// collection, but authorizes a requested course before any scoped records are
// read. A course from another tenant is indistinguishable from a missing one.
func RequireOptionalProgressCourseAccess(courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, present := c.GetQuery("course_id")
		if !present {
			c.Next()
			return
		}
		courseID, ok := requiredUUID(c, raw, "course id")
		if !ok {
			return
		}
		orgID, ok := requestOrgID(c)
		if !ok {
			return
		}
		if _, err := courses.FindByID(c.Request.Context(), courseID, orgID); err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if !AuthorizeCourseAccess(c, courseTeachers, groups, courses, courseID) {
			return
		}
		c.Next()
	}
}

type courseIDResolver func(ctx context.Context, resourceID, orgID uuid.UUID) (uuid.UUID, error)

// RequireAssignedResource resolves a course from a nested resource and then
// applies the same role-specific course rule as a course URL.
func RequireAssignedResource(
	courseTeachers repositories.CourseTeacherRepository,
	groups repositories.GroupRepository,
	courses repositories.CourseRepository,
	parameter string,
	resolve courseIDResolver,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID, ok := requiredUUID(c, c.Param(parameter), parameter)
		if !ok {
			return
		}
		orgID, ok := requestOrgID(c)
		if !ok {
			return
		}
		courseID, err := resolve(c.Request.Context(), resourceID, orgID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if !AuthorizeCourseAccess(c, courseTeachers, groups, courses, courseID) {
			return
		}
		c.Next()
	}
}

func RequireAssignedModule(courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository, modules repositories.ModuleRepository, parameter string) gin.HandlerFunc {
	return RequireAssignedResource(courseTeachers, groups, courses, parameter, func(ctx context.Context, moduleID, orgID uuid.UUID) (uuid.UUID, error) {
		module, err := modules.FindByID(ctx, moduleID, orgID)
		if err != nil {
			return uuid.Nil, err
		}
		return module.CourseID, nil
	})
}

func RequireAssignedLesson(courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository, lessons repositories.LessonRepository, modules repositories.ModuleRepository, parameter string) gin.HandlerFunc {
	return func(c *gin.Context) {
		lessonID, ok := requiredUUID(c, c.Param(parameter), parameter)
		if !ok || !AuthorizeLessonAccess(c, courseTeachers, groups, courses, lessons, modules, lessonID) {
			return
		}
		c.Next()
	}
}

// AuthorizeLessonAccess resolves a lesson's course and applies the same
// role-specific access rules used by course routes. Handlers whose lesson ID
// comes from JSON instead of a URL parameter use this helper after binding.
func AuthorizeLessonAccess(
	c *gin.Context,
	courseTeachers repositories.CourseTeacherRepository,
	groups repositories.GroupRepository,
	courses repositories.CourseRepository,
	lessons repositories.LessonRepository,
	modules repositories.ModuleRepository,
	lessonID uuid.UUID,
) bool {
	orgID, ok := requestOrgID(c)
	if !ok {
		return false
	}
	lesson, err := lessons.FindByID(c.Request.Context(), lessonID, orgID)
	if err != nil {
		c.Error(err)
		c.Abort()
		return false
	}
	module, err := modules.FindByID(c.Request.Context(), lesson.ModuleID, orgID)
	if err != nil {
		c.Error(err)
		c.Abort()
		return false
	}
	return AuthorizeCourseAccess(c, courseTeachers, groups, courses, module.CourseID)
}

func RequireAssignedAssignment(courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository, assignments repositories.AssignmentRepository, parameter string) gin.HandlerFunc {
	return RequireAssignedResource(courseTeachers, groups, courses, parameter, func(ctx context.Context, assignmentID, orgID uuid.UUID) (uuid.UUID, error) {
		assignment, err := assignments.FindByID(ctx, assignmentID, orgID)
		if err != nil {
			return uuid.Nil, err
		}
		return assignment.CourseID, nil
	})
}

func RequireAssignedQuiz(courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository, quizzes repositories.QuizRepository, lessons repositories.LessonRepository, modules repositories.ModuleRepository, parameter string) gin.HandlerFunc {
	return RequireAssignedResource(courseTeachers, groups, courses, parameter, func(ctx context.Context, quizID, orgID uuid.UUID) (uuid.UUID, error) {
		quiz, err := quizzes.FindByID(ctx, quizID, orgID)
		if err != nil {
			return uuid.Nil, err
		}
		lesson, err := lessons.FindByID(ctx, quiz.LessonID, orgID)
		if err != nil {
			return uuid.Nil, err
		}
		module, err := modules.FindByID(ctx, lesson.ModuleID, orgID)
		if err != nil {
			return uuid.Nil, err
		}
		return module.CourseID, nil
	})
}

func RequireAssignedExam(courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository, exams repositories.ExamRepository, parameter string) gin.HandlerFunc {
	return RequireAssignedResource(courseTeachers, groups, courses, parameter, func(ctx context.Context, examID, orgID uuid.UUID) (uuid.UUID, error) {
		exam, err := exams.FindByID(ctx, examID, orgID)
		if err != nil {
			return uuid.Nil, err
		}
		return exam.CourseID, nil
	})
}

func RequireAssignedSubmission(courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository, assignments repositories.AssignmentRepository, parameter string) gin.HandlerFunc {
	return RequireAssignedResource(courseTeachers, groups, courses, parameter, func(ctx context.Context, submissionID, orgID uuid.UUID) (uuid.UUID, error) {
		submission, err := assignments.FindSubmissionByID(ctx, submissionID)
		if err != nil {
			return uuid.Nil, err
		}
		assignment, err := assignments.FindByID(ctx, submission.AssignmentID, orgID)
		if err != nil {
			return uuid.Nil, err
		}
		return assignment.CourseID, nil
	})
}

func RequireAssignedExamAttempt(courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository, attempts repositories.ExamAttemptRepository, exams repositories.ExamRepository, parameter string) gin.HandlerFunc {
	return RequireAssignedResource(courseTeachers, groups, courses, parameter, func(ctx context.Context, attemptID, orgID uuid.UUID) (uuid.UUID, error) {
		attempt, err := attempts.FindByID(ctx, attemptID)
		if err != nil {
			return uuid.Nil, err
		}
		exam, err := exams.FindByID(ctx, attempt.ExamID, orgID)
		if err != nil {
			return uuid.Nil, err
		}
		return exam.CourseID, nil
	})
}

// AuthorizeCourseAccess applies tenant-role permissions to a resolved course.
// It is shared by route middleware and handlers that receive a resource ID in
// a request payload instead of the URL.
func AuthorizeCourseAccess(c *gin.Context, courseTeachers repositories.CourseTeacherRepository, groups repositories.GroupRepository, courses repositories.CourseRepository, courseID uuid.UUID) bool {
	userID, err := uuid.Parse(GetUserID(c))
	if err != nil {
		c.Error(apperrors.UnauthorizedError("invalid user identity"))
		c.Abort()
		return false
	}
	orgID, ok := requestOrgID(c)
	if !ok {
		return false
	}
	switch GetRole(c) {
	case string(entities.RoleAdmin), string(entities.RoleSuperAdmin):
		return true
	case string(entities.RoleTeacher):
		assigned, err := courseTeachers.Exists(c.Request.Context(), courseID, userID, orgID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return false
		}
		if !assigned {
			c.Error(apperrors.ForbiddenError("you are not assigned to this course"))
			c.Abort()
			return false
		}
		return true
	case string(entities.RoleStudent):
		if _, err := groups.GetStudentGroup(c.Request.Context(), courseID, userID, orgID); err != nil {
			if apperrors.IsNotFound(err) {
				c.Error(apperrors.ForbiddenError("you are not enrolled in this course"))
			} else {
				c.Error(err)
			}
			c.Abort()
			return false
		}

		course, err := courses.FindByID(c.Request.Context(), courseID, orgID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return false
		}
		if course.Status != entities.StatusPublished {
			c.Error(apperrors.ForbiddenError("this course is not published"))
			c.Abort()
			return false
		}
		return true
	default:
		c.Error(apperrors.ForbiddenError("insufficient role"))
		c.Abort()
		return false
	}
}

func requestOrgID(c *gin.Context) (uuid.UUID, bool) {
	orgID, err := uuid.Parse(GetOrgID(c))
	if err != nil {
		c.Error(apperrors.UnauthorizedError("invalid organization identity"))
		c.Abort()
		return uuid.Nil, false
	}
	return orgID, true
}

func requiredUUID(c *gin.Context, value, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(value)
	if err != nil {
		c.Error(apperrors.ValidationError("invalid " + name))
		c.Abort()
		return uuid.Nil, false
	}
	return id, true
}
