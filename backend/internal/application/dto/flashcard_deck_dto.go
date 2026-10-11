package dto

import "time"

type CreateFlashcardDeckRequest struct {
	LessonID string             `json:"lesson_id" binding:"required,uuid4"`
	Title    string             `json:"title" binding:"required,min=1,max=160"`
	Cards    []DeckFlashcardDTO `json:"cards" binding:"required,min=1,max=50,dive"`
}

type DeckFlashcardDTO struct {
	Term       string `json:"term" binding:"required,max=500"`
	Definition string `json:"definition" binding:"required,max=2000"`
}

type ReviewFlashcardRequest struct {
	CardIndex int    `json:"card_index" binding:"gte=0"`
	Rating    string `json:"rating" binding:"required,oneof=again good easy"`
}

type FlashcardReviewDTO struct {
	Reviews        int       `json:"reviews"`
	IntervalDays   int       `json:"interval_days"`
	LastReviewedAt time.Time `json:"last_reviewed_at"`
	NextReviewAt   time.Time `json:"next_review_at"`
}

type FlashcardDeckDTO struct {
	ID        string                     `json:"id"`
	LessonID  string                     `json:"lesson_id"`
	Title     string                     `json:"title"`
	Cards     []DeckFlashcardDTO         `json:"cards"`
	Progress  map[int]FlashcardReviewDTO `json:"progress"`
	Activity  map[string]int             `json:"activity"`
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
}
