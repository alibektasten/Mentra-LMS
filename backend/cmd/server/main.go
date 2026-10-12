package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infraAI "github.com/ailms/backend/internal/infrastructure/ai"
	"github.com/ailms/backend/internal/infrastructure/config"
	"github.com/ailms/backend/internal/infrastructure/database"
	infraRepos "github.com/ailms/backend/internal/infrastructure/repositories"
	"github.com/ailms/backend/internal/infrastructure/storage"

	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/handlers"
	"github.com/ailms/backend/internal/delivery/http/routes"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logLevel := slog.LevelInfo
	logFormat := "json"
	if cfg.Server.Mode == "debug" {
		logLevel = slog.LevelDebug
		logFormat = "text"
	}
	log := logger.New(logLevel, logFormat)
	logger.SetDefault(log)

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	log.Info("connected to database")

	if err := database.RunMigrations(db); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	log.Info("migrations completed")

	// Init repositories
	userRepo := infraRepos.NewGORMUserRepository(db)
	orgRepo := infraRepos.NewGORMOrganizationRepository(db)
	memberRepo := infraRepos.NewGORMMembershipRepository(db)
	courseRepo := infraRepos.NewGORMCourseRepository(db)
	moduleRepo := infraRepos.NewGORMModuleRepository(db)
	lessonRepo := infraRepos.NewGORMLessonRepository(db)
	quizRepo := infraRepos.NewGORMQuizRepository(db)
	progressRepo := infraRepos.NewGORMProgressRepository(db)
	assignmentRepo := infraRepos.NewGORMAssignmentRepository(db)
	quizAttemptRepo := infraRepos.NewGORMQuizAttemptRepository(db)
	fileAttachmentRepo := infraRepos.NewGORMFileAttachmentRepository(db)
	announcementRepo := infraRepos.NewGORMAnnouncementRepository(db)
	groupRepo := infraRepos.NewGORMGroupRepository(db)
	examRepo := infraRepos.NewGORMExamRepository(db)
	examAttemptRepo := infraRepos.NewGORMExamAttemptRepository(db)
	extraAttemptGrantRepo := infraRepos.NewGORMExtraAttemptGrantRepository(db)
	courseTeacherRepo := infraRepos.NewGORMCourseTeacherRepository(db)
	flashcardDeckRepo := infraRepos.NewGORMFlashcardDeckRepository(db)
	certificateRepo := infraRepos.NewGORMCertificateRepository(db)

	// Init storage
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	localStore := storage.NewLocalStorage(uploadDir)

	// Init AI service. The provider is selected server-side so API keys never
	// reach the browser.
	aiService, err := infraAI.NewService(cfg.AI, cfg.Ollama, cfg.OpenAI)
	if err != nil {
		log.Error("failed to initialize AI service", "error", err)
		os.Exit(1)
	}
	if warmup, ok := aiService.(interface{ Warmup(context.Context) }); ok {
		go warmup.Warmup(context.Background())
	}

	seedSuperAdmin(context.Background(), userRepo, cfg, log)

	// Init use cases
	authUC := usecases.NewAuthUseCase(userRepo, orgRepo, memberRepo, cfg.JWT.Secret)
	courseUC := usecases.NewCourseUseCase(courseRepo, memberRepo, groupRepo, courseTeacherRepo)
	courseTeacherUC := usecases.NewCourseTeacherUseCase(courseTeacherRepo, memberRepo)
	memberUC := usecases.NewMemberUseCase(userRepo, memberRepo)
	moduleUC := usecases.NewModuleUseCase(moduleRepo, courseRepo)
	lessonUC := usecases.NewLessonUseCase(lessonRepo, moduleRepo)
	quizUC := usecases.NewQuizUseCase(quizRepo, lessonRepo)
	progressUC := usecases.NewProgressUseCase(progressRepo, lessonRepo, aiService, moduleRepo, quizRepo, quizAttemptRepo)
	aiUC := usecases.NewAIUseCase(lessonRepo, quizRepo, assignmentRepo, aiService)
	userUC := usecases.NewUserUseCase(userRepo)
	superAdminUC := usecases.NewSuperAdminUseCase(userRepo, orgRepo, memberRepo)
	assignmentUC := usecases.NewAssignmentUseCase(assignmentRepo, moduleRepo, userRepo)
	quizAttemptUC := usecases.NewQuizAttemptUseCase(quizAttemptRepo, quizRepo, lessonRepo, aiService)
	gradeUC := usecases.NewGradeUseCase(assignmentRepo, quizAttemptRepo, quizRepo, memberRepo, examRepo, examAttemptRepo)
	announcementUC := usecases.NewAnnouncementUseCase(announcementRepo, courseRepo)
	groupUC := usecases.NewGroupUseCase(groupRepo, courseRepo)
	examUC := usecases.NewExamUseCase(examRepo, examAttemptRepo, extraAttemptGrantRepo, courseRepo)
	teacherUC := usecases.NewTeacherUseCase(courseRepo, groupRepo, courseTeacherRepo, userRepo, assignmentRepo)
	flashcardDeckUC := usecases.NewFlashcardDeckUseCase(flashcardDeckRepo, lessonRepo)
	certificateUC := usecases.NewCertificateUseCase(certificateRepo, groupRepo, courseRepo, moduleRepo, lessonRepo, progressRepo, userRepo, orgRepo, courseTeacherRepo)
	notificationUC := usecases.NewNotificationUseCase(announcementRepo, courseRepo, groupRepo, courseTeacherRepo)

	fileAttachmentUC := usecases.NewFileAttachmentUseCase(fileAttachmentRepo)

	// Init handlers
	authHandler := handlers.NewAuthHandler(authUC, cfg.GoogleOAuth, cfg.JWT.Secret)
	courseHandler := handlers.NewCourseHandler(courseUC)
	courseTeacherHandler := handlers.NewCourseTeacherHandler(courseTeacherUC)
	memberHandler := handlers.NewMemberHandler(memberUC)
	moduleHandler := handlers.NewModuleHandler(moduleUC)
	lessonHandler := handlers.NewLessonHandler(lessonUC)
	aiHandler := handlers.NewAIHandler(aiUC, courseTeacherRepo, groupRepo, courseRepo, lessonRepo, moduleRepo)
	progressHandler := handlers.NewProgressHandler(progressUC)
	userHandler := handlers.NewUserHandler(userUC)
	superAdminHandler := handlers.NewSuperAdminHandler(superAdminUC, authUC)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentUC, localStore)
	quizAttemptHandler := handlers.NewQuizAttemptHandler(quizAttemptUC)
	quizHandler := handlers.NewQuizHandler(quizUC)
	fileAttachmentHandler := handlers.NewFileAttachmentHandler(fileAttachmentUC, courseTeacherRepo, groupRepo, courseRepo, lessonRepo, moduleRepo, assignmentRepo)
	gradeHandler := handlers.NewGradeHandler(gradeUC)
	uploadHandler := handlers.NewUploadHandler(localStore)
	announcementHandler := handlers.NewAnnouncementHandler(announcementUC)
	groupHandler := handlers.NewGroupHandler(groupUC)
	examHandler := handlers.NewExamHandler(examUC, localStore)
	teacherHandler := handlers.NewTeacherHandler(teacherUC)
	flashcardDeckHandler := handlers.NewFlashcardDeckHandler(flashcardDeckUC)
	certificateHandler := handlers.NewCertificateHandler(certificateUC)
	notificationHandler := handlers.NewNotificationHandler(notificationUC)

	gin.SetMode(cfg.Server.Mode)
	router := routes.NewRouter(routes.Dependencies{
		AuthHandler:           authHandler,
		CourseHandler:         courseHandler,
		CourseTeacherHandler:  courseTeacherHandler,
		MemberHandler:         memberHandler,
		ModuleHandler:         moduleHandler,
		LessonHandler:         lessonHandler,
		AIHandler:             aiHandler,
		ProgressHandler:       progressHandler,
		UserHandler:           userHandler,
		SuperAdminHandler:     superAdminHandler,
		AssignmentHandler:     assignmentHandler,
		QuizAttemptHandler:    quizAttemptHandler,
		QuizHandler:           quizHandler,
		FileAttachmentHandler: fileAttachmentHandler,
		GradeHandler:          gradeHandler,
		UploadHandler:         uploadHandler,
		AnnouncementHandler:   announcementHandler,
		GroupHandler:          groupHandler,
		ExamHandler:           examHandler,
		TeacherHandler:        teacherHandler,
		FlashcardDeckHandler:  flashcardDeckHandler,
		CertificateHandler:    certificateHandler,
		NotificationHandler:   notificationHandler,
		OrgRepo:               orgRepo,
		CourseRepo:            courseRepo,
		CourseTeacherRepo:     courseTeacherRepo,
		GroupRepo:             groupRepo,
		ModuleRepo:            moduleRepo,
		LessonRepo:            lessonRepo,
		AssignmentRepo:        assignmentRepo,
		QuizRepo:              quizRepo,
		ExamRepo:              examRepo,
		ExamAttemptRepo:       examAttemptRepo,
		JWTSecret:             cfg.JWT.Secret,
		AIRequestLimit:        cfg.AI.DailyRequestLimit,
		Logger:                log,
		UploadDir:             uploadDir,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
	}
	log.Info("server exited")
}

func seedSuperAdmin(ctx context.Context, userRepo repositories.UserRepository, cfg *config.Config, log *logger.Logger) {
	email := cfg.SuperAdmin.Email
	password := cfg.SuperAdmin.Password
	if email == "" || password == "" {
		return
	}

	existing, _ := userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash super admin password", "error", err)
		return
	}

	now := time.Now()
	user := &entities.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         "Super Admin",
		Role:         string(entities.RoleSuperAdmin),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		log.Error("failed to seed super admin", "error", err)
		return
	}
	log.Info("super admin seeded", "email", email)
}
