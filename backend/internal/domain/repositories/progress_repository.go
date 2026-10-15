package repositories

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

type LessonProgressRepository interface {
	Create(ctx context.Context, progress *entities.LessonProgress) error
	FindByUserAndLesson(ctx context.Context, userID, lessonID, orgID uuid.UUID) (*entities.LessonProgress, error)
	FindByUser(ctx context.Context, userID, orgID uuid.UUID) ([]entities.LessonProgress, error)
	FindByUserAndCourse(ctx context.Context, userID, courseID, orgID uuid.UUID) ([]entities.LessonProgress, error)
	Update(ctx context.Context, progress *entities.LessonProgress) error
}
