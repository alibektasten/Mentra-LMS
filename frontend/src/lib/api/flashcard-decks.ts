import { FlashcardDTO } from './ai'
import { apiClient } from './client'

export type ReviewRating = 'again' | 'good' | 'easy'

export interface FlashcardReviewDTO {
  reviews: number
  interval_days: number
  last_reviewed_at: string
  next_review_at: string
}

export interface FlashcardDeckDTO {
  id: string
  lesson_id: string
  title: string
  cards: FlashcardDTO[]
  progress: Record<number, FlashcardReviewDTO>
  activity: Record<string, number>
  created_at: string
  updated_at: string
}

export async function listFlashcardDecks(): Promise<FlashcardDeckDTO[]> {
  const res = await apiClient.get<FlashcardDeckDTO[]>('/flashcard-decks')
  return res.data
}

export async function createFlashcardDeck(input: {
  lesson_id: string
  title: string
  cards: FlashcardDTO[]
}): Promise<FlashcardDeckDTO> {
  const res = await apiClient.post<FlashcardDeckDTO>('/flashcard-decks', input)
  return res.data
}

export async function reviewFlashcardDeck(deckId: string, cardIndex: number, rating: ReviewRating): Promise<FlashcardDeckDTO> {
  const res = await apiClient.post<FlashcardDeckDTO>(`/flashcard-decks/${deckId}/reviews`, {
    card_index: cardIndex,
    rating,
  })
  return res.data
}

export async function deleteFlashcardDeck(deckId: string): Promise<void> {
  await apiClient.delete(`/flashcard-decks/${deckId}`)
}
