'use client'

import { motion } from 'framer-motion'
import { ArrowRight, CalendarCheck2, Sparkles } from 'lucide-react'
import { SavedFlashcardDeck, useFlashcardDeckStore } from '@/lib/stores/flashcard-decks.store'

interface FlashcardReviewQueueProps {
  ownerId?: string
  onStartReview: (deck: SavedFlashcardDeck, cardIndex: number) => void
}

function getDueIndexes(deck: SavedFlashcardDeck) {
  const now = Date.now()
  return deck.cards.flatMap((_, index) => {
    const progress = deck.progress[index]
    return !progress || new Date(progress.nextReviewAt).getTime() <= now ? [index] : []
  })
}

export function FlashcardReviewQueue({ ownerId, onStartReview }: FlashcardReviewQueueProps) {
  const decks = useFlashcardDeckStore((state) => state.decks)
  const queue = decks
    .filter((deck) => deck.ownerId === ownerId)
    .map((deck) => ({ deck, dueIndexes: getDueIndexes(deck) }))
    .filter(({ dueIndexes }) => dueIndexes.length > 0)
    .sort((a, b) => b.dueIndexes.length - a.dueIndexes.length)

  const totalDue = queue.reduce((sum, item) => sum + item.dueIndexes.length, 0)
  if (totalDue === 0) return null

  const next = queue[0]

  return (
    <motion.section
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className="relative overflow-hidden rounded-2xl border border-emerald-200 bg-[radial-gradient(circle_at_top_right,_rgba(167,243,208,0.8),_transparent_38%),linear-gradient(135deg,_#ecfdf5,_#f8fffb)] p-4 sm:p-5"
    >
      <div className="absolute -right-4 -top-4 text-emerald-200/70"><Sparkles className="h-20 w-20" /></div>
      <div className="relative flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div className="flex items-start gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-600 text-white shadow-[0_8px_18px_rgba(5,120,93,0.2)]"><CalendarCheck2 className="h-5 w-5" /></span>
          <div>
            <p className="text-sm font-semibold tracking-[-0.02em] text-emerald-950">Your review queue is ready</p>
            <p className="mt-0.5 text-xs leading-relaxed text-emerald-800/70">{totalDue} card{totalDue === 1 ? '' : 's'} due across {queue.length} deck{queue.length === 1 ? '' : 's'}.</p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => onStartReview(next.deck, next.dueIndexes[0])}
          className="inline-flex items-center justify-center gap-1.5 rounded-xl bg-emerald-600 px-3.5 py-2.5 text-xs font-semibold text-white shadow-[0_6px_16px_rgba(5,120,93,0.2)] transition-all hover:-translate-y-px hover:bg-emerald-700"
        >
          Review {next.deck.title}
          <ArrowRight className="h-3.5 w-3.5" />
        </button>
      </div>
    </motion.section>
  )
}
