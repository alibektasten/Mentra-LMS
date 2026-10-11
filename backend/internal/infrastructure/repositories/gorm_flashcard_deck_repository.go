package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/database"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMFlashcardDeckRepository struct {
	db *gorm.DB
}

func NewGORMFlashcardDeckRepository(db *gorm.DB) *GORMFlashcardDeckRepository {
	return &GORMFlashcardDeckRepository{db: db}
}

func (r *GORMFlashcardDeckRepository) Create(ctx context.Context, deck *entities.FlashcardDeck) error {
	model, err := toFlashcardDeckModel(deck)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return apperrors.InternalError(err.Error())
	}
	return nil
}

func (r *GORMFlashcardDeckRepository) FindByUser(ctx context.Context, userID, orgID uuid.UUID) ([]entities.FlashcardDeck, error) {
	var models []database.FlashcardDeckModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND org_id = ?", userID.String(), orgID.String()).
		Order("updated_at DESC").
		Find(&models).Error; err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	decks := make([]entities.FlashcardDeck, len(models))
	for i := range models {
		deck, err := toFlashcardDeckEntity(&models[i])
		if err != nil {
			return nil, err
		}
		decks[i] = *deck
	}
	return decks, nil
}

func (r *GORMFlashcardDeckRepository) FindByID(ctx context.Context, id, userID, orgID uuid.UUID) (*entities.FlashcardDeck, error) {
	var model database.FlashcardDeckModel
	err := r.db.WithContext(ctx).First(&model, "id = ? AND user_id = ? AND org_id = ?", id.String(), userID.String(), orgID.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("flashcard deck", id.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toFlashcardDeckEntity(&model)
}

func (r *GORMFlashcardDeckRepository) Update(ctx context.Context, deck *entities.FlashcardDeck) error {
	model, err := toFlashcardDeckModel(deck)
	if err != nil {
		return err
	}
	model.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND org_id = ?", deck.ID.String(), deck.UserID.String(), deck.OrgID.String()).
		Updates(map[string]interface{}{
			"cards":      model.Cards,
			"progress":   model.Progress,
			"activity":   model.Activity,
			"updated_at": model.UpdatedAt,
		})
	if result.Error != nil {
		return apperrors.InternalError(result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFoundError("flashcard deck", deck.ID.String())
	}
	deck.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *GORMFlashcardDeckRepository) Delete(ctx context.Context, id, userID, orgID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ? AND org_id = ?", id.String(), userID.String(), orgID.String()).Delete(&database.FlashcardDeckModel{})
	if result.Error != nil {
		return apperrors.InternalError(result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFoundError("flashcard deck", id.String())
	}
	return nil
}

func toFlashcardDeckModel(deck *entities.FlashcardDeck) (*database.FlashcardDeckModel, error) {
	cards, err := json.Marshal(deck.Cards)
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	progress, err := json.Marshal(deck.Progress)
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	activity, err := json.Marshal(deck.Activity)
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return &database.FlashcardDeckModel{
		ID:        deck.ID.String(),
		UserID:    deck.UserID.String(),
		OrgID:     deck.OrgID.String(),
		LessonID:  deck.LessonID.String(),
		Title:     deck.Title,
		Cards:     string(cards),
		Progress:  string(progress),
		Activity:  string(activity),
		CreatedAt: deck.CreatedAt,
		UpdatedAt: deck.UpdatedAt,
	}, nil
}

func toFlashcardDeckEntity(model *database.FlashcardDeckModel) (*entities.FlashcardDeck, error) {
	var cards []entities.Flashcard
	if err := json.Unmarshal([]byte(model.Cards), &cards); err != nil {
		return nil, apperrors.InternalError("stored flashcard deck has invalid cards")
	}
	progress := make(map[int]entities.FlashcardReview)
	if model.Progress != "" {
		if err := json.Unmarshal([]byte(model.Progress), &progress); err != nil {
			return nil, apperrors.InternalError("stored flashcard deck has invalid review progress")
		}
	}
	activity := make(map[string]int)
	if model.Activity != "" {
		if err := json.Unmarshal([]byte(model.Activity), &activity); err != nil {
			return nil, apperrors.InternalError("stored flashcard deck has invalid review activity")
		}
	}
	id, _ := uuid.Parse(model.ID)
	userID, _ := uuid.Parse(model.UserID)
	orgID, _ := uuid.Parse(model.OrgID)
	lessonID, _ := uuid.Parse(model.LessonID)
	return &entities.FlashcardDeck{
		ID:        id,
		UserID:    userID,
		OrgID:     orgID,
		LessonID:  lessonID,
		Title:     model.Title,
		Cards:     cards,
		Progress:  progress,
		Activity:  activity,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}
