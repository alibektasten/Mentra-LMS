package routes

import (
	"github.com/ailms/backend/internal/delivery/http/handlers"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthHandler           *handlers.AuthHandler
	CourseHandler         *handlers.CourseHandler
	CourseTeacherHandler  *handlers.CourseTeacherHandler
	MemberHandler         *handlers.MemberHandler
	ModuleHandler         *handlers.ModuleHandler
	LessonHandler         *handlers.LessonHandler
	AIHandler             *handlers.AIHandler
	ProgressHandler       *handlers.ProgressHandler
	UserHandler           *handlers.UserHandler
	SuperAdminHandler     *handlers.SuperAdminHandler
	AssignmentHandler     *handlers.AssignmentHandler
	QuizAttemptHandler    *handlers.QuizAttemptHandler
	QuizHandler           *handlers.QuizHandler
	FileAttachmentHandler *handlers.FileAttachmentHandler
	GradeHandler          *handlers.GradeHandler
	UploadHandler         *handlers.UploadHandler
	PrivateFileHandler    *handlers.PrivateFileHandler
	AnnouncementHandler   *handlers.AnnouncementHandler
	GroupHandler          *handlers.GroupHandler
	ExamHandler           *handlers.ExamHandler
	TeacherHandler        *handlers.TeacherHandler
	FlashcardDeckHandler  *handlers.FlashcardDeckHandler
	AttendanceHandler     *handlers.AttendanceHandler
	CertificateHandler    *handlers.CertificateHandler
	NotificationHandler   *handlers.NotificationHandler
	OrgRepo               repositories.OrganizationRepository
	MembershipRepo        repositories.MembershipRepository
	CourseRepo            repositories.CourseRepository
	CourseTeacherRepo     repositories.CourseTeacherRepository
	GroupRepo             repositories.GroupRepository
	ModuleRepo            repositories.ModuleRepository
	LessonRepo            repositories.LessonRepository
	AssignmentRepo        repositories.AssignmentRepository
	QuizRepo              repositories.QuizRepository
	ExamRepo              repositories.ExamRepository
	ExamAttemptRepo       repositories.ExamAttemptRepository
	JWTSecret             string
	AIRequestLimit        int
	Logger                *logger.Logger
	UploadDir             string
}

func NewRouter(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		middleware.CORS(),
		middleware.RequestLogger(deps.Logger),
		middleware.ErrorHandler(),
		middleware.RateLimiter(100),
	)
	// Lightweight liveness probe for a reverse proxy or container platform.
	// It intentionally has no auth requirement and does not disclose internals.
	health := func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	}
	r.GET("/health", health)
	r.GET("/healthz", health)

	api := r.Group("/api/v1")
	api.GET("/avatars/:id", deps.PrivateFileHandler.Avatar)
	api.GET("/certificates/:code", deps.CertificateHandler.Verify)

	// Super admin public login — no tenant middleware
	api.POST("/super-admin/auth/login", middleware.RateLimiter(5), deps.SuperAdminHandler.Login)

	// Super admin protected routes — Auth + RequireRole only, no Tenant middleware
	sa := api.Group("/super-admin")
	sa.Use(middleware.Auth(deps.JWTSecret), middleware.RequireRole(entities.RoleSuperAdmin))
	sa.GET("/stats", deps.SuperAdminHandler.GetStats)
	sa.GET("/orgs", deps.SuperAdminHandler.ListOrgs)
	sa.POST("/orgs", deps.SuperAdminHandler.CreateOrg)
	sa.PUT("/orgs/:id", deps.SuperAdminHandler.UpdateOrg)
	sa.DELETE("/orgs/:id", deps.SuperAdminHandler.DeleteOrg)
	sa.GET("/users", deps.SuperAdminHandler.ListUsers)
	sa.PUT("/users/:membershipID/role", deps.SuperAdminHandler.UpdateUserRole)
	sa.POST("/orgs/invite-admin", deps.SuperAdminHandler.InviteOrgAdmin)

	// Public routes
	auth := api.Group("/auth")
	// Authentication endpoints are intentionally tighter than normal API
	// traffic to reduce credential stuffing and account enumeration attempts.
	auth.Use(middleware.RateLimiter(10))
	auth.POST("/register", deps.AuthHandler.Register)
	auth.POST("/login", deps.AuthHandler.Login)
	auth.GET("/google/status", deps.AuthHandler.GoogleStatus)
	auth.GET("/google/start", deps.AuthHandler.GoogleStart)
	auth.GET("/google/callback", deps.AuthHandler.GoogleCallback)

	// Protected routes
	protected := api.Group("/")
	protected.Use(
		middleware.Auth(deps.JWTSecret),
		middleware.TenantWithMembership(deps.OrgRepo, deps.MembershipRepo),
	)

	teacher := protected.Group("/teacher")
	teacher.Use(middleware.RequireRole(entities.RoleTeacher))
	teacher.GET("/overview", deps.TeacherHandler.Overview)
	teacher.GET("/students", deps.TeacherHandler.Students)
	teacher.GET("/review-queue", deps.TeacherHandler.ReviewQueue)
	teacher.GET("/calendar", deps.TeacherHandler.Calendar)

	// Courses
	courses := protected.Group("/courses")
	courses.Use(middleware.RequireAssignedCourse(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo))
	courses.GET("", deps.CourseHandler.List)
	courses.GET("/:id", deps.CourseHandler.Get)
	courses.POST("",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.CourseHandler.Create,
	)
	courses.PUT("/:id",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.CourseHandler.Update,
	)
	courses.DELETE("/:id",
		middleware.RequireRole(entities.RoleAdmin),
		deps.CourseHandler.Delete,
	)
	courses.POST("/:id/publish",
		middleware.RequireRole(entities.RoleAdmin),
		deps.CourseHandler.Publish,
	)

	// Course teachers
	courses.GET("/:id/teachers", deps.CourseTeacherHandler.List)
	courses.POST("/:id/teachers",
		middleware.RequireRole(entities.RoleAdmin),
		deps.CourseTeacherHandler.Assign,
	)
	courses.DELETE("/:id/teachers/:teacherID",
		middleware.RequireRole(entities.RoleAdmin),
		deps.CourseTeacherHandler.Remove,
	)

	// Members (admin only)
	members := protected.Group("/members")
	members.Use(middleware.RequireRole(entities.RoleAdmin))
	members.GET("", deps.MemberHandler.List)
	members.POST("/invite", deps.MemberHandler.Invite)
	members.POST("/import", deps.MemberHandler.BulkImport)
	members.DELETE("/:id", deps.MemberHandler.Remove)
	members.PUT("/:id/role", deps.MemberHandler.UpdateRole)

	// Modules (inline in courses group — shares :id wildcard for course ID)
	courses.GET("/:id/modules", deps.ModuleHandler.List)
	courses.POST("/:id/modules",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.ModuleHandler.Create,
	)
	courses.GET("/:id/modules/:moduleID", deps.ModuleHandler.Get)
	courses.PUT("/:id/modules/:moduleID",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.ModuleHandler.Update,
	)
	courses.DELETE("/:id/modules/:moduleID",
		middleware.RequireRole(entities.RoleAdmin),
		deps.ModuleHandler.Delete,
	)

	// Lessons (nested under modules — separate tree, no wildcard conflict)
	lessons := protected.Group("/modules/:moduleID/lessons")
	lessons.Use(middleware.RequireAssignedModule(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo, deps.ModuleRepo, "moduleID"))
	lessons.GET("", deps.LessonHandler.List)
	lessons.POST("",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.LessonHandler.Create,
	)
	lessons.GET("/:lessonID", deps.LessonHandler.Get)
	lessons.PUT("/:lessonID",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.LessonHandler.Update,
	)
	lessons.DELETE("/:lessonID",
		middleware.RequireRole(entities.RoleAdmin),
		deps.LessonHandler.Delete,
	)

	// Progress
	progress := protected.Group("/progress")
	progress.GET("", middleware.RequireOptionalProgressCourseAccess(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo), deps.ProgressHandler.GetProgress)
	progress.GET("/insights", deps.ProgressHandler.GetInsights)
	progress.POST("/lessons/:lessonID/complete",
		middleware.RequireAssignedLesson(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo, deps.LessonRepo, deps.ModuleRepo, "lessonID"),
		deps.ProgressHandler.Complete,
	)

	// AI study tools are available to all tenant roles. Quiz authoring is kept
	// to teachers and admins because it exposes answer keys.
	ai := protected.Group("/ai")
	// Model calls are costly and can be abused even by an authenticated user.
	ai.Use(middleware.RateLimiter(6))
	ai.POST("/summarize", deps.AIHandler.SummarizeLesson)
	ai.POST("/generate-quiz",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AIHandler.GenerateQuiz,
	)
	ai.POST("/assignment-feedback", deps.AIHandler.GetAssignmentFeedback)
	ai.POST("/generate-flashcards", deps.AIHandler.GenerateFlashcards)
	ai.POST("/generate-teaching-plan",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AIHandler.GenerateTeachingPlan,
	)

	// Flashcard decks are learner-owned and sync across a user's devices.
	flashcardDecks := protected.Group("/flashcard-decks")
	flashcardDecks.GET("", deps.FlashcardDeckHandler.List)
	flashcardDecks.POST("", deps.FlashcardDeckHandler.Create)
	flashcardDecks.POST("/:id/reviews", deps.FlashcardDeckHandler.Review)
	flashcardDecks.DELETE("/:id", deps.FlashcardDeckHandler.Delete)

	// User profile
	protected.GET("/me", deps.UserHandler.GetMe)
	protected.PUT("/me", deps.UserHandler.UpdateMe)
	protected.POST("/me/avatar", middleware.RateLimiter(2), deps.PrivateFileHandler.UploadAvatar)
	protected.GET("/certificates", deps.CertificateHandler.ListMine)
	protected.GET("/notifications", deps.NotificationHandler.List)

	// Upload
	protected.POST("/upload", middleware.RateLimiter(2), deps.UploadHandler.Upload)
	privateFiles := protected.Group("/files")
	privateFiles.GET("/lessons/:id", deps.PrivateFileHandler.Lesson)
	privateFiles.GET("/attachments/:id", deps.PrivateFileHandler.Attachment)
	privateFiles.GET("/submissions/:id", deps.PrivateFileHandler.Submission)
	privateFiles.GET("/exam-attempts/:id", deps.PrivateFileHandler.ExamAttempt)

	// Assignments (inline in courses group — same :id wildcard for course ID)
	courses.GET("/:id/modules/:moduleID/assignments", deps.AssignmentHandler.ListByModule)
	courses.POST("/:id/modules/:moduleID/assignments",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AssignmentHandler.Create,
	)

	assignmentsStandalone := protected.Group("/assignments")
	assignmentsStandalone.Use(middleware.RequireAssignedAssignment(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo, deps.AssignmentRepo, "id"))
	assignmentsStandalone.GET("/:id", deps.AssignmentHandler.Get)
	assignmentsStandalone.PUT("/:id",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AssignmentHandler.Update,
	)
	assignmentsStandalone.DELETE("/:id",
		middleware.RequireRole(entities.RoleAdmin),
		deps.AssignmentHandler.Delete,
	)
	assignmentsStandalone.POST("/:id/submit", deps.AssignmentHandler.Submit)
	assignmentsStandalone.GET("/:id/my-submission", deps.AssignmentHandler.GetMySubmission)
	assignmentsStandalone.DELETE("/:id/my-submission", deps.AssignmentHandler.DeleteMySubmission)
	assignmentsStandalone.GET("/:id/submissions",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AssignmentHandler.ListSubmissions,
	)

	// Submissions grading
	submissions := protected.Group("/submissions")
	submissions.Use(middleware.RequireAssignedSubmission(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo, deps.AssignmentRepo, "id"))
	submissions.PUT("/:id/grade",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AssignmentHandler.GradeSubmission,
	)

	// Quiz attempts + teacher CRUD
	quizzes := protected.Group("/quizzes")
	quizzes.Use(middleware.RequireAssignedQuiz(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo, deps.QuizRepo, deps.LessonRepo, deps.ModuleRepo, "id"))
	quizzes.POST("/:id/attempt", deps.QuizAttemptHandler.SubmitAttempt)
	quizzes.GET("/:id/my-attempt", deps.QuizAttemptHandler.GetMyAttempt)
	quizzes.PUT("/:id",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.QuizHandler.Update,
	)
	quizzes.DELETE("/:id",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.QuizHandler.Delete,
	)

	// Quiz by lesson (standalone lesson routes)
	lessonRoutes := protected.Group("/lessons")
	lessonRoutes.Use(middleware.RequireAssignedLesson(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo, deps.LessonRepo, deps.ModuleRepo, "lessonID"))
	lessonRoutes.POST("/:lessonID/quiz",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.QuizHandler.Create,
	)
	lessonRoutes.GET("/:lessonID/quiz", deps.QuizHandler.GetByLesson)

	// File attachments
	attachments := protected.Group("/attachments")
	attachments.GET("", deps.FileAttachmentHandler.ListByRef)
	attachments.POST("",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.FileAttachmentHandler.Create,
	)
	attachments.DELETE("/:id",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.FileAttachmentHandler.Delete,
	)

	// Adaptive learning
	courses.GET("/:id/pacing", deps.ProgressHandler.GetCoursePacing)
	quizzes.GET("/:id/remediation", deps.QuizAttemptHandler.GetRemediation)

	// Grades
	courses.GET("/:id/my-grades", deps.GradeHandler.GetMyGrades)
	courses.GET("/:id/deadlines", deps.GradeHandler.GetUpcomingDeadlines)
	courses.GET("/:id/gradebook",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GradeHandler.GetGradebook,
	)
	courses.PATCH("/:id/gradebook/students/:studentID",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GradeHandler.UpdateGradebookStudent,
	)
	courses.PUT("/:id/gradebook/publication",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GradeHandler.SetGradebookPublication,
	)

	// Announcements
	courses.GET("/:id/announcements", deps.AnnouncementHandler.List)
	courses.POST("/:id/announcements",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AnnouncementHandler.Create,
	)
	courses.DELETE("/:id/announcements/:aID",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AnnouncementHandler.Delete,
	)

	// Standalone groups (org-level management, admin only)
	orgGroups := protected.Group("/groups")
	orgGroups.Use(middleware.RequireRole(entities.RoleAdmin))
	orgGroups.GET("", deps.GroupHandler.ListByOrg)
	orgGroups.POST("", deps.GroupHandler.CreateStandalone)
	orgGroups.GET("/:groupID", deps.GroupHandler.Get)
	orgGroups.PUT("/:groupID", deps.GroupHandler.Update)
	orgGroups.DELETE("/:groupID", deps.GroupHandler.Delete)
	orgGroups.POST("/:groupID/assign-course", deps.GroupHandler.AssignToCourse)
	orgGroups.DELETE("/:groupID/assign-course", deps.GroupHandler.UnassignFromCourse)
	orgGroups.GET("/:groupID/members", deps.GroupHandler.ListMembers)
	orgGroups.POST("/:groupID/members", deps.GroupHandler.AddMember)
	orgGroups.DELETE("/:groupID/members/:studentID", deps.GroupHandler.RemoveMember)
	orgGroups.GET("/:groupID/schedules", deps.GroupHandler.ListSchedules)
	orgGroups.POST("/:groupID/schedules", deps.GroupHandler.AddSchedule)
	orgGroups.DELETE("/:groupID/schedules/:schedID", deps.GroupHandler.DeleteSchedule)

	// Groups (course-scoped)
	courses.GET("/:id/groups",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.List,
	)
	courses.POST("/:id/groups",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.Create,
	)
	courses.GET("/:id/groups/:groupID",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.Get,
	)
	courses.PUT("/:id/groups/:groupID",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.Update,
	)
	courses.DELETE("/:id/groups/:groupID",
		middleware.RequireRole(entities.RoleAdmin),
		deps.GroupHandler.Delete,
	)
	courses.GET("/:id/groups/:groupID/members",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.ListMembers,
	)
	courses.GET("/:id/groups/:groupID/attendance",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AttendanceHandler.GetSheet,
	)
	courses.PUT("/:id/groups/:groupID/attendance",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.AttendanceHandler.SaveSheet,
	)
	courses.POST("/:id/groups/:groupID/members",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.AddMember,
	)
	courses.DELETE("/:id/groups/:groupID/members/:studentID",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.RemoveMember,
	)
	courses.GET("/:id/groups/:groupID/schedules",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.ListSchedules,
	)
	courses.POST("/:id/groups/:groupID/schedules",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.AddSchedule,
	)
	courses.DELETE("/:id/groups/:groupID/schedules/:schedID",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.GroupHandler.DeleteSchedule,
	)
	courses.GET("/:id/my-group", deps.GroupHandler.GetMyGroup)

	// Exams (course-level)
	courses.GET("/:id/exams", deps.ExamHandler.List)
	courses.POST("/:id/exams",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.ExamHandler.Create,
	)

	exams := protected.Group("/exams")
	exams.Use(middleware.RequireAssignedExam(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo, deps.ExamRepo, "id"))
	exams.GET("/:id", deps.ExamHandler.Get)
	exams.PUT("/:id",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.ExamHandler.Update,
	)
	exams.DELETE("/:id",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.ExamHandler.Delete,
	)
	exams.POST("/:id/start", deps.ExamHandler.StartAttempt)
	exams.GET("/:id/my-attempts", deps.ExamHandler.MyAttempts)
	exams.GET("/:id/attempts",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.ExamHandler.ListAttempts,
	)
	exams.POST("/:id/grant-attempt",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.ExamHandler.GrantExtra,
	)

	examAttempts := protected.Group("/exam-attempts")
	examAttempts.Use(middleware.RequireAssignedExamAttempt(deps.CourseTeacherRepo, deps.GroupRepo, deps.CourseRepo, deps.ExamAttemptRepo, deps.ExamRepo, "id"))
	examAttempts.PUT("/:id/draft", deps.ExamHandler.SaveDraft)
	examAttempts.POST("/:id/submit", deps.ExamHandler.SubmitAttempt)
	examAttempts.PUT("/:id/grade",
		middleware.RequireRole(entities.RoleTeacher, entities.RoleAdmin),
		deps.ExamHandler.GradeFile,
	)

	return r
}
