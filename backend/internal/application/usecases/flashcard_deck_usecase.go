package usecases

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type FlashcardDeckUseCase struct {
	deckRepo   repositories.FlashcardDeckRepository
	lessonRepo repositories.LessonRepository
}

func NewFlashcardDeckUseCase(deckRepo repositories.FlashcardDeckRepository, lessonRepo repositories.LessonRepository) *FlashcardDeckUseCase {
	return &FlashcardDeckUseCase{deckRepo: deckRepo, lessonRepo: lessonRepo}
}

func (uc *FlashcardDeckUseCase) Create(ctx context.Context, userID, orgID uuid.UUID, req dto.CreateFlashcardDeckRequest) (*dto.FlashcardDeckDTO, error) {
	lessonID, err := uuid.Parse(req.LessonID)
	if err != nil {
		return nil, apperrors.ValidationError("invalid lesson id")
	}
	if _, err := uc.lessonRepo.FindByID(ctx, lessonID, orgID); err != nil {
		return nil, err
	}

	cards := make([]entities.Flashcard, len(req.Cards))
	for i, card := range req.Cards {
		cards[i] = entities.Flashcard{Term: card.Term, Definition: card.Definition}
	}

	now := time.Now().UTC()
	deck := &entities.FlashcardDeck{
		ID:        uuid.New(),
		UserID:    userID,
		OrgID:     orgID,
		LessonID:  lessonID,
		Title:     req.Title,
		Cards:     cards,
		Progress:  make(map[int]entities.FlashcardReview),
		Activity:  make(map[string]int),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.deckRepo.Create(ctx, deck); err != nil {
		return nil, err
	}
	return toFlashcardDeckDTO(deck), nil
}

func (uc *FlashcardDeckUseCase) List(ctx context.Context, userID, orgID uuid.UUID) ([]dto.FlashcardDeckDTO, error) {
	decks, err := uc.deckRepo.FindByUser(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.FlashcardDeckDTO, len(decks))
	for i := range decks {
		result[i] = *toFlashcardDeckDTO(&decks[i])
	}
	return result, nil
}

func (uc *FlashcardDeckUseCase) Review(ctx context.Context, userID, orgID, deckID uuid.UUID, req dto.ReviewFlashcardRequest) (*dto.FlashcardDeckDTO, error) {
	deck, err := uc.deckRepo.FindByID(ctx, deckID, userID, orgID)
	if err != nil {
		return nil, err
	}
	if req.CardIndex >= len(deck.Cards) {
		return nil, apperrors.ValidationError("card index is outside this deck")
	}

	if deck.Progress == nil {
		deck.Progress = make(map[int]entities.FlashcardReview)
	}
	if deck.Activity == nil {
		deck.Activity = make(map[string]int)
	}
	now := time.Now().UTC()
	previous := deck.Progress[req.CardIndex]
	intervalDays, nextReviewAt := nextFlashcardReview(previous, req.Rating, now)
	deck.Progress[req.CardIndex] = entities.FlashcardReview{
		Reviews:        previous.Reviews + 1,
		IntervalDays:   intervalDays,
		LastReviewedAt: now,
		NextReviewAt:   nextReviewAt,
	}
	deck.Activity[now.Format("2006-01-02")]++
	deck.UpdatedAt = now
	if err := uc.deckRepo.Update(ctx, deck); err != nil {
		return nil, err
	}
	return toFlashcardDeckDTO(deck), nil
}

func (uc *FlashcardDeckUseCase) Delete(ctx context.Context, userID, orgID, deckID uuid.UUID) error {
	return uc.deckRepo.Delete(ctx, deckID, userID, orgID)
}

func nextFlashcardReview(previous entities.FlashcardReview, rating string, now time.Time) (int, time.Time) {
	if rating == "again" {
		return 0, now.Add(10 * time.Minute)
	}

	base := previous.IntervalDays
	if base < 1 {
		base = 1
	}
	interval := base * 2
	if rating == "easy" {
		interval = base * 7 / 2
		if interval < 3 {
			interval = 3
		}
	}
	maxInterval := 90
	if rating == "easy" {
		maxInterval = 180
	}
	if interval > maxInterval {
		interval = maxInterval
	}
	return interval, now.AddDate(0, 0, interval)
}

func toFlashcardDeckDTO(deck *entities.FlashcardDeck) *dto.FlashcardDeckDTO {
	cards := make([]dto.DeckFlashcardDTO, len(deck.Cards))
	for i, card := range deck.Cards {
		cards[i] = dto.DeckFlashcardDTO{Term: card.Term, Definition: card.Definition}
	}
	progress := make(map[int]dto.FlashcardReviewDTO, len(deck.Progress))
	for index, review := range deck.Progress {
		progress[index] = dto.FlashcardReviewDTO{
			Reviews:        review.Reviews,
			IntervalDays:   review.IntervalDays,
			LastReviewedAt: review.LastReviewedAt,
			NextReviewAt:   review.NextReviewAt,
		}
	}
	return &dto.FlashcardDeckDTO{
		ID:        deck.ID.String(),
		LessonID:  deck.LessonID.String(),
		Title:     deck.Title,
		Cards:     cards,
		Progress:  progress,
		Activity:  deck.Activity,
		CreatedAt: deck.CreatedAt,
		UpdatedAt: deck.UpdatedAt,
	}
}
