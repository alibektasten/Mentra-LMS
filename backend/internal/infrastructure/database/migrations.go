package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserModel struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Name         string `gorm:"not null"`
	AvatarURL    string
	Role         string `gorm:"default:null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }

type OrganizationModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	Name      string `gorm:"not null"`
	Slug      string `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (OrganizationModel) TableName() string { return "organizations" }

type MembershipModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid;index;not null"`
	OrgID     string `gorm:"type:uuid;index;not null"`
	Role      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (MembershipModel) TableName() string { return "memberships" }

type CourseModel struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	OrgID       string `gorm:"type:uuid;index;not null"`
	Title       string `gorm:"not null"`
	Description string
	Status      string `gorm:"not null;default:draft"`
	CreatedBy   string `gorm:"type:uuid;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (CourseModel) TableName() string { return "courses" }

type ModuleModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	CourseID  string `gorm:"type:uuid;index;not null"`
	OrgID     string `gorm:"type:uuid;index;not null"`
	Title     string `gorm:"not null"`
	Position  int    `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ModuleModel) TableName() string { return "modules" }

type LessonModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	ModuleID  string `gorm:"type:uuid;index;not null"`
	OrgID     string `gorm:"type:uuid;index;not null"`
	Title     string `gorm:"not null"`
	Content   string `gorm:"type:text"`
	Type      string `gorm:"not null"`
	VideoURL  string
	LinkURL   string
	FileURL   string
	Position  int `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (LessonModel) TableName() string { return "lessons" }

type QuizModel struct {
	ID                  string `gorm:"type:uuid;primaryKey"`
	LessonID            string `gorm:"type:uuid;index;not null"`
	OrgID               string `gorm:"type:uuid;index;not null"`
	Title               string `gorm:"not null"`
	MaxPoints           int    `gorm:"default:10"`
	DueDate             *time.Time
	AllowLateSubmission bool                `gorm:"default:false"`
	Questions           []QuizQuestionModel `gorm:"foreignKey:QuizID"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (QuizModel) TableName() string { return "quizzes" }

type QuizQuestionModel struct {
	ID        string            `gorm:"type:uuid;primaryKey"`
	QuizID    string            `gorm:"type:uuid;index;not null"`
	Question  string            `gorm:"not null"`
	Position  int               `gorm:"not null;default:0"`
	Answers   []QuizAnswerModel `gorm:"foreignKey:QuestionID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (QuizQuestionModel) TableName() string { return "quiz_questions" }

type QuizAnswerModel struct {
	ID         string `gorm:"type:uuid;primaryKey"`
	QuestionID string `gorm:"type:uuid;index;not null"`
	Answer     string `gorm:"not null"`
	IsCorrect  bool   `gorm:"not null;default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (QuizAnswerModel) TableName() string { return "quiz_answers" }

type LessonProgressModel struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	UserID      string `gorm:"type:uuid;index;not null"`
	LessonID    string `gorm:"type:uuid;index;not null"`
	OrgID       string `gorm:"type:uuid;index;not null"`
	CompletedAt *time.Time
	Score       *float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (LessonProgressModel) TableName() string { return "lesson_progress" }

type AssignmentModel struct {
	ID                  string `gorm:"type:uuid;primaryKey"`
	OrgID               string `gorm:"type:uuid;not null;index"`
	CourseID            string `gorm:"type:uuid;not null;index"`
	ModuleID            string `gorm:"type:uuid;not null;index"`
	Title               string `gorm:"not null"`
	Description         string
	Rubric              string `gorm:"type:text"`
	MaxPoints           int    `gorm:"not null;default:100"`
	DueDate             *time.Time
	AllowLateSubmission bool `gorm:"default:false"`
	Position            int  `gorm:"default:0"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (AssignmentModel) TableName() string { return "assignments" }

type AssignmentSubmissionModel struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	AssignmentID string `gorm:"type:uuid;not null;index"`
	StudentID    string `gorm:"type:uuid;not null;index"`
	OrgID        string `gorm:"type:uuid;not null;index"`
	TextContent  string
	LinkURL      string
	FilePath     string
	Score        *int
	Feedback     string
	GradedBy     *string `gorm:"type:uuid"`
	GradedAt     *time.Time
	SubmittedAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (AssignmentSubmissionModel) TableName() string { return "assignment_submissions" }

type QuizAttemptModel struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	QuizID      string `gorm:"type:uuid;not null;index"`
	StudentID   string `gorm:"type:uuid;not null;index"`
	OrgID       string `gorm:"type:uuid;not null;index"`
	Score       int
	MaxScore    int
	Answers     string `gorm:"type:jsonb"`
	SubmittedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (QuizAttemptModel) TableName() string { return "quiz_attempts" }

type FileAttachmentModel struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	OrgID        string `gorm:"type:uuid;not null;index"`
	UploaderID   string `gorm:"type:uuid;not null"`
	OriginalName string `gorm:"not null"`
	StoredPath   string `gorm:"not null"`
	MimeType     string
	SizeBytes    int64
	RefType      string `gorm:"not null"`
	RefID        string `gorm:"type:uuid;not null;index"`
	CreatedAt    time.Time
}

func (FileAttachmentModel) TableName() string { return "file_attachments" }

type AnnouncementModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	CourseID  string `gorm:"type:uuid;not null;index"`
	OrgID     string `gorm:"type:uuid;not null;index"`
	AuthorID  string `gorm:"type:uuid;not null"`
	Title     string `gorm:"not null"`
	Content   string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AnnouncementModel) TableName() string { return "announcements" }

type GroupModel struct {
	ID        string  `gorm:"type:uuid;primaryKey"`
	CourseID  *string `gorm:"type:uuid;index"`
	OrgID     string  `gorm:"type:uuid;not null;index"`
	TeacherID *string `gorm:"type:uuid"`
	Name      string  `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (GroupModel) TableName() string { return "groups" }

type GroupScheduleModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	GroupID   string `gorm:"type:uuid;not null;index"`
	DayOfWeek int    `gorm:"not null;default:0"`
	StartTime string `gorm:"not null"`
	EndTime   string `gorm:"not null"`
	Location  string
	CreatedAt time.Time
}

func (GroupScheduleModel) TableName() string { return "group_schedules" }

type GroupMemberModel struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	GroupID   string    `gorm:"type:uuid;not null;index"`
	StudentID string    `gorm:"type:uuid;not null;index"`
	OrgID     string    `gorm:"type:uuid;not null;index"`
	JoinedAt  time.Time `gorm:"not null"`
}

func (GroupMemberModel) TableName() string { return "group_members" }

type ExamModel struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	CourseID        string `gorm:"type:uuid;index;not null"`
	OrgID           string `gorm:"type:uuid;index;not null"`
	AssessmentType  string `gorm:"not null;default:other;index"`
	Title           string `gorm:"not null"`
	Description     string `gorm:"type:text"`
	DurationMinutes int    `gorm:"not null;default:60"`
	MaxAttempts     int    `gorm:"not null;default:1"`
	DueDate         *time.Time
	MCQEnabled      bool                `gorm:"not null;default:false"`
	MCQPoints       int                 `gorm:"not null;default:0"`
	FileEnabled     bool                `gorm:"not null;default:false"`
	FilePoints      int                 `gorm:"not null;default:0"`
	Questions       []ExamQuestionModel `gorm:"foreignKey:ExamID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (ExamModel) TableName() string { return "exams" }

type ExamQuestionModel struct {
	ID        string            `gorm:"type:uuid;primaryKey"`
	ExamID    string            `gorm:"type:uuid;index;not null"`
	Question  string            `gorm:"not null"`
	Position  int               `gorm:"not null;default:0"`
	Answers   []ExamAnswerModel `gorm:"foreignKey:QuestionID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ExamQuestionModel) TableName() string { return "exam_questions" }

type ExamAnswerModel struct {
	ID         string `gorm:"type:uuid;primaryKey"`
	QuestionID string `gorm:"type:uuid;index;not null"`
	Answer     string `gorm:"not null"`
	IsCorrect  bool   `gorm:"not null;default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (ExamAnswerModel) TableName() string { return "exam_answers" }

type ExamAttemptModel struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	ExamID       string `gorm:"type:uuid;not null;index"`
	StudentID    string `gorm:"type:uuid;not null;index"`
	OrgID        string `gorm:"type:uuid;not null;index"`
	Status       string `gorm:"not null;default:in_progress"`
	StartedAt    time.Time
	ExpiresAt    time.Time
	SubmittedAt  *time.Time
	MCQAnswers   string `gorm:"type:jsonb"`
	MCQScore     *int
	MCQMaxScore  int
	FilePath     string
	FileFeedback string
	FileScore    *int
	FilePoints   int
	TotalScore   *int
	GradedBy     *string `gorm:"type:uuid"`
	GradedAt     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (ExamAttemptModel) TableName() string { return "exam_attempts" }

type ExtraAttemptGrantModel struct {
	ID         string `gorm:"type:uuid;primaryKey"`
	ExamID     string `gorm:"type:uuid;not null;index"`
	StudentID  string `gorm:"type:uuid;not null;index"`
	OrgID      string `gorm:"type:uuid;not null;index"`
	GrantedBy  string `gorm:"type:uuid;not null"`
	ExtraCount int    `gorm:"not null;default:1"`
	CreatedAt  time.Time
}

func (ExtraAttemptGrantModel) TableName() string { return "extra_attempt_grants" }

type CourseTeacherModel struct {
	ID         string    `gorm:"type:uuid;primaryKey"`
	CourseID   string    `gorm:"type:uuid;not null;uniqueIndex:idx_course_teacher_unique"`
	TeacherID  string    `gorm:"type:uuid;not null;uniqueIndex:idx_course_teacher_unique"`
	OrgID      string    `gorm:"type:uuid;not null;uniqueIndex:idx_course_teacher_unique"`
	AssignedAt time.Time `gorm:"not null"`
}

func (CourseTeacherModel) TableName() string { return "course_teachers" }

type FlashcardDeckModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid;not null;index"`
	OrgID     string `gorm:"type:uuid;not null;index"`
	LessonID  string `gorm:"type:uuid;not null;index"`
	Title     string `gorm:"not null"`
	Cards     string `gorm:"type:jsonb;not null"`
	Progress  string `gorm:"type:jsonb;not null"`
	Activity  string `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (FlashcardDeckModel) TableName() string { return "flashcard_decks" }

type CertificateModel struct {
	ID               string    `gorm:"type:uuid;primaryKey"`
	UserID           string    `gorm:"type:uuid;not null;uniqueIndex:idx_certificate_user_course"`
	CourseID         string    `gorm:"type:uuid;not null;uniqueIndex:idx_certificate_user_course"`
	OrgID            string    `gorm:"type:uuid;not null;index"`
	VerificationCode string    `gorm:"uniqueIndex;not null"`
	IssuedAt         time.Time `gorm:"not null"`
}

func (CertificateModel) TableName() string { return "certificates" }

func RunMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(
		&UserModel{},
		&OrganizationModel{},
		&MembershipModel{},
		&CourseModel{},
		&ModuleModel{},
		&LessonModel{},
		&QuizModel{},
		&QuizQuestionModel{},
		&QuizAnswerModel{},
		&LessonProgressModel{},
		&AssignmentModel{},
		&AssignmentSubmissionModel{},
		&QuizAttemptModel{},
		&FileAttachmentModel{},
		&AnnouncementModel{},
		&GroupModel{},
		&GroupScheduleModel{},
		&GroupMemberModel{},
		&ExamModel{},
		&ExamQuestionModel{},
		&ExamAnswerModel{},
		&ExamAttemptModel{},
		&ExtraAttemptGrantModel{},
		&CourseTeacherModel{},
		&FlashcardDeckModel{},
		&CertificateModel{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
