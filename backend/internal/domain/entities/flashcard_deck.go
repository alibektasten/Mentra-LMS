package entities

import (
	"time"

	"github.com/google/uuid"
)

// FlashcardDeck is a learner-owned AI study deck. Cards and review state are
// kept together because decks are small and are always read as a whole.
type FlashcardDeck struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	OrgID     uuid.UUID
	LessonID  uuid.UUID
	Title     string
	Cards     []Flashcard
	Progress  map[int]FlashcardReview
	Activity  map[string]int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FlashcardReview struct {
	Reviews        int       `json:"reviews"`
	IntervalDays   int       `json:"interval_days"`
	LastReviewedAt time.Time `json:"last_reviewed_at"`
	NextReviewAt   time.Time `json:"next_review_at"`
}
