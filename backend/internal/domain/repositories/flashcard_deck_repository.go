package repositories

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

type FlashcardDeckRepository interface {
	Create(ctx context.Context, deck *entities.FlashcardDeck) error
	FindByUser(ctx context.Context, userID, orgID uuid.UUID) ([]entities.FlashcardDeck, error)
	FindByID(ctx context.Context, id, userID, orgID uuid.UUID) (*entities.FlashcardDeck, error)
	Update(ctx context.Context, deck *entities.FlashcardDeck) error
	Delete(ctx context.Context, id, userID, orgID uuid.UUID) error
}
