package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

// ErrExamConfigurationChanged signals that an operation observed an older
// exam configuration than the one protected by the shared database lock.
var ErrExamConfigurationChanged = errors.New("exam configuration changed")

type ExamRepository interface {
	Create(ctx context.Context, exam *entities.Exam) error
	FindByID(ctx context.Context, id, orgID uuid.UUID) (*entities.Exam, error)
	FindByCourse(ctx context.Context, courseID, orgID uuid.UUID) ([]*entities.Exam, error)
	Update(ctx context.Context, exam *entities.Exam) error
	UpdateWithAttemptGuard(ctx context.Context, exam *entities.Exam, rejectIfAttempts bool) (bool, error)
	Delete(ctx context.Context, id, orgID uuid.UUID) error
}

type ExamAttemptRepository interface {
	Create(ctx context.Context, attempt *entities.ExamAttempt) error
	StartOrResume(ctx context.Context, candidate *entities.ExamAttempt, allowedAttempts int, allowNewAttempt bool, now time.Time, expectedExamUpdatedAt *time.Time) (*entities.ExamAttempt, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.ExamAttempt, error)
	FindByExamAndStudent(ctx context.Context, examID, studentID uuid.UUID) ([]*entities.ExamAttempt, error)
	FindByExam(ctx context.Context, examID uuid.UUID) ([]*entities.ExamAttempt, error)
	FindActiveAttempt(ctx context.Context, examID, studentID uuid.UUID) (*entities.ExamAttempt, error)
	HasAttempts(ctx context.Context, examID uuid.UUID) (bool, error)
	SaveDraftIfInProgress(ctx context.Context, attemptID, studentID uuid.UUID, answers []entities.ExamMCQAnswer, updatedAt time.Time) (bool, error)
	SubmitIfInProgress(ctx context.Context, attempt *entities.ExamAttempt) (bool, error)
	ExpireIfInProgress(ctx context.Context, attemptID, studentID uuid.UUID, updatedAt time.Time) (bool, error)
	Update(ctx context.Context, attempt *entities.ExamAttempt) error
	CountByExamAndStudent(ctx context.Context, examID, studentID uuid.UUID) (int, error)
}

type ExtraAttemptGrantRepository interface {
	Create(ctx context.Context, grant *entities.ExtraAttemptGrant) error
	SumByExamAndStudent(ctx context.Context, examID, studentID uuid.UUID) (int, error)
}
