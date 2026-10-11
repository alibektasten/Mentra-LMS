'use client'

import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { FlashcardDTO } from '@/lib/api/ai'
import type { FlashcardDeckDTO } from '@/lib/api/flashcard-decks'

export type ReviewRating = 'again' | 'good' | 'easy'

export interface FlashcardProgress {
  reviews: number
  intervalDays: number
  lastReviewedAt: string
  nextReviewAt: string
}

export interface SavedFlashcardDeck {
  id: string
  ownerId: string
  title: string
  lessonId: string
  cards: FlashcardDTO[]
  createdAt: string
  updatedAt: string
  progress: Record<number, FlashcardProgress>
  dailyActivity: Record<string, number>
  synced: boolean
}

interface FlashcardDeckState {
  decks: SavedFlashcardDeck[]
  createDeck: (input: Pick<SavedFlashcardDeck, 'ownerId' | 'title' | 'lessonId' | 'cards'>) => SavedFlashcardDeck
  saveSyncedDeck: (ownerId: string, deck: FlashcardDeckDTO) => SavedFlashcardDeck
  syncRemoteDecks: (ownerId: string, decks: FlashcardDeckDTO[]) => void
  recordReview: (deckId: string, cardIndex: number, rating: ReviewRating) => void
  removeDeck: (deckId: string) => void
}

function getNextReview(previous: FlashcardProgress | undefined, rating: ReviewRating, now: Date) {
  if (rating === 'again') {
    return { intervalDays: 0, nextReviewAt: new Date(now.getTime() + 10 * 60 * 1000).toISOString() }
  }

  const baseInterval = previous?.intervalDays || 1
  const intervalDays = rating === 'easy'
    ? Math.min(Math.max(3, Math.floor(baseInterval * 3.5)), 180)
    : Math.min(Math.max(1, Math.round(baseInterval * 2)), 90)

  return {
    intervalDays,
    nextReviewAt: new Date(now.getTime() + intervalDays * 24 * 60 * 60 * 1000).toISOString(),
  }
}

export const useFlashcardDeckStore = create<FlashcardDeckState>()(
  persist(
    (set) => ({
      decks: [],
      createDeck: (input) => {
        const now = new Date().toISOString()
        const deck: SavedFlashcardDeck = {
          ...input,
          id: `local-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
          createdAt: now,
          updatedAt: now,
          progress: {},
          dailyActivity: {},
          synced: false,
        }
        set((state) => ({ decks: [deck, ...state.decks] }))
        return deck
      },
      saveSyncedDeck: (ownerId, deck) => {
        const savedDeck = toSavedDeck(ownerId, deck)
        set((state) => ({
          decks: [savedDeck, ...state.decks.filter((existing) => existing.id !== savedDeck.id)],
        }))
        return savedDeck
      },
      syncRemoteDecks: (ownerId, decks) => {
        const syncedDecks = decks.map((deck) => toSavedDeck(ownerId, deck))
        set((state) => ({
          decks: [
            ...syncedDecks,
            ...state.decks.filter((deck) => deck.ownerId !== ownerId || !deck.synced),
          ],
        }))
      },
      recordReview: (deckId, cardIndex, rating) => {
        const now = new Date()
        set((state) => ({
          decks: state.decks.map((deck) => {
            if (deck.id !== deckId) return deck
            const previous = deck.progress[cardIndex]
            const schedule = getNextReview(previous, rating, now)
            const today = now.toISOString().slice(0, 10)
            const dailyActivity = deck.dailyActivity ?? {}
            return {
              ...deck,
              updatedAt: now.toISOString(),
              progress: {
                ...deck.progress,
                [cardIndex]: {
                  reviews: (previous?.reviews ?? 0) + 1,
                  lastReviewedAt: now.toISOString(),
                  ...schedule,
                },
              },
              dailyActivity: {
                ...dailyActivity,
                [today]: (dailyActivity[today] ?? 0) + 1,
              },
            }
          }),
        }))
      },
      removeDeck: (deckId) => set((state) => ({ decks: state.decks.filter((deck) => deck.id !== deckId) })),
    }),
    { name: 'mentra-flashcard-decks' }
  )
)

function toSavedDeck(ownerId: string, deck: FlashcardDeckDTO): SavedFlashcardDeck {
  const progress = Object.entries(deck.progress ?? {}).reduce<Record<number, FlashcardProgress>>((result, [index, review]) => {
    result[Number(index)] = {
      reviews: review.reviews,
      intervalDays: review.interval_days,
      lastReviewedAt: review.last_reviewed_at,
      nextReviewAt: review.next_review_at,
    }
    return result
  }, {})

  return {
    id: deck.id,
    ownerId,
    lessonId: deck.lesson_id,
    title: deck.title,
    cards: deck.cards,
    progress,
    dailyActivity: deck.activity ?? {},
    createdAt: deck.created_at,
    updatedAt: deck.updated_at,
    synced: true,
  }
}
