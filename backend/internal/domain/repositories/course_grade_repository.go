package repositories

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

type CourseGradeRepository interface {
	Find(ctx context.Context, courseID, studentID, orgID uuid.UUID) (*entities.CourseGrade, error)
	UpdateOverride(ctx context.Context, courseID, studentID, orgID, actorID uuid.UUID, overrideGrade *float64, overrideNote *string, clearOverride bool, updatedAt time.Time) (*entities.CourseGrade, error)
	SetPublication(ctx context.Context, grades []*entities.CourseGrade) error
}
