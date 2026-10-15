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

// AIDailyUsageModel is a database-backed daily quota counter. The composite
// unique key is the conflict target for atomic increments across app restarts
// and horizontally scaled API instances.
type AIDailyUsageModel struct {
	OrgID     string    `gorm:"type:uuid;not null;uniqueIndex:idx_ai_daily_usage_org_user_day"`
	UserID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_ai_daily_usage_org_user_day"`
	UsageDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_ai_daily_usage_org_user_day"`
	Count     int       `gorm:"not null;default:0"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (AIDailyUsageModel) TableName() string { return "ai_daily_usage" }

// CourseGradeModel contains the teacher-managed publication and adjustment
// state for a student's course result. The calculated component scores remain
// in submissions and attempts.
type CourseGradeModel struct {
	ID                   string `gorm:"type:uuid;primaryKey"`
	CourseID             string `gorm:"type:uuid;not null;uniqueIndex:idx_course_grade_student"`
	StudentID            string `gorm:"type:uuid;not null;uniqueIndex:idx_course_grade_student"`
	OrgID                string `gorm:"type:uuid;not null;index"`
	OverrideGrade        *float64
	OverrideNote         string `gorm:"type:text"`
	PublishedAt          *time.Time
	PublishedBy          *string `gorm:"type:uuid"`
	PublicationChangedAt *time.Time
	PublicationChangedBy *string `gorm:"type:uuid"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (CourseGradeModel) TableName() string { return "course_grades" }

// CourseGradeChangeModel is deliberately separate from CourseGradeModel so an
// override cannot erase its own history during a later correction or removal.
type CourseGradeChangeModel struct {
	ID               string `gorm:"type:uuid;primaryKey"`
	CourseGradeID    string `gorm:"type:uuid;not null;index"`
	CourseID         string `gorm:"type:uuid;not null;index"`
	StudentID        string `gorm:"type:uuid;not null;index"`
	OrgID            string `gorm:"type:uuid;not null;index"`
	ChangedBy        string `gorm:"type:uuid;not null"`
	Action           string `gorm:"not null"`
	PreviousOverride *float64
	NewOverride      *float64
	Reason           string `gorm:"type:text;not null"`
	CreatedAt        time.Time
}

func (CourseGradeChangeModel) TableName() string { return "course_grade_changes" }

type AttendanceRecordModel struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	GroupID   string    `gorm:"type:uuid;not null;uniqueIndex:idx_attendance_group_student_day"`
	StudentID string    `gorm:"type:uuid;not null;uniqueIndex:idx_attendance_group_student_day"`
	OrgID     string    `gorm:"type:uuid;not null;index"`
	Date      time.Time `gorm:"type:date;not null;uniqueIndex:idx_attendance_group_student_day"`
	Status    string    `gorm:"not null"`
	MarkedBy  string    `gorm:"type:uuid;not null"`
	MarkedAt  time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AttendanceRecordModel) TableName() string { return "attendance_records" }

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
		&AIDailyUsageModel{},
		&CourseGradeModel{},
		&CourseGradeChangeModel{},
		&AttendanceRecordModel{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	if err := ensureExamAttemptIntegrity(db); err != nil {
		return fmt.Errorf("exam attempt integrity migration failed: %w", err)
	}
	return nil
}

// ensureExamAttemptIntegrity preserves historical rows while making the
// current-state invariant enforceable by PostgreSQL: a student can have only
// one in-progress attempt for an exam. Attempts that are past their recorded
// server deadline can be deterministically expired. If multiple still-valid
// attempts exist, migration stops instead of choosing a winner: an operator
// must explicitly resolve that academic conflict before re-running it.
func ensureExamAttemptIntegrity(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}

	graceCutoff := time.Now().UTC().Add(-time.Minute)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			UPDATE exam_attempts
			SET status = 'expired', updated_at = NOW()
			WHERE status = 'in_progress' AND expires_at < ?
		`, graceCutoff).Error; err != nil {
			return err
		}

		var unresolvedDuplicateGroups int64
		if err := tx.Raw(`
			SELECT COUNT(*)
			FROM (
				SELECT exam_id, student_id
				FROM exam_attempts
				WHERE status = 'in_progress'
				GROUP BY exam_id, student_id
				HAVING COUNT(*) > 1
			) AS duplicate_attempts
		`).Scan(&unresolvedDuplicateGroups).Error; err != nil {
			return err
		}
		if unresolvedDuplicateGroups > 0 {
			return fmt.Errorf(
				"cannot add idx_exam_attempt_one_in_progress: %d active exam/student duplicate group(s) require explicit academic resolution; retain every row, set the non-continuing attempt status through the approved workflow, then rerun migrations",
				unresolvedDuplicateGroups,
			)
		}

		return tx.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_attempt_one_in_progress
			ON exam_attempts (exam_id, student_id)
			WHERE status = 'in_progress'
		`).Error
	})
}
