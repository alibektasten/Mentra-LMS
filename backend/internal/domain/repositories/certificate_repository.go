package repositories

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

type CertificateRepository interface {
	Create(ctx context.Context, certificate *entities.Certificate) error
	FindByUserAndCourse(ctx context.Context, userID, courseID, orgID uuid.UUID) (*entities.Certificate, error)
	ListByUser(ctx context.Context, userID, orgID uuid.UUID) ([]entities.Certificate, error)
	FindByVerificationCode(ctx context.Context, code string) (*entities.Certificate, error)
}
