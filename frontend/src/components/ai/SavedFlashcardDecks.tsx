'use client'

import { motion } from 'framer-motion'
import { ArrowUpRight, CalendarClock, Layers3, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { deleteFlashcardDeck } from '@/lib/api/flashcard-decks'
import { SavedFlashcardDeck, useFlashcardDeckStore } from '@/lib/stores/flashcard-decks.store'

interface SavedFlashcardDecksProps {
  ownerId?: string
  onOpenDeck: (deck: SavedFlashcardDeck) => void
}

function getDueCount(deck: SavedFlashcardDeck) {
  const now = Date.now()
  return deck.cards.filter((_, index) => {
    const progress = deck.progress[index]
    return !progress || new Date(progress.nextReviewAt).getTime() <= now
  }).length
}

export function SavedFlashcardDecks({ ownerId, onOpenDeck }: SavedFlashcardDecksProps) {
  const decks = useFlashcardDeckStore((state) => state.decks)
  const removeDeck = useFlashcardDeckStore((state) => state.removeDeck)
  const savedDecks = decks
    .filter((deck) => deck.ownerId === ownerId)
    .sort((a, b) => getDueCount(b) - getDueCount(a) || new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())

  if (savedDecks.length === 0) return null

  const handleDelete = async (deck: SavedFlashcardDeck) => {
    if (deck.synced) {
      try {
        await deleteFlashcardDeck(deck.id)
      } catch {
        toast.error('Could not delete this deck from your account. Please try again.')
        return
      }
    }
    removeDeck(deck.id)
  }

  return (
    <section className="rounded-2xl border border-[#e7e8e2] bg-white p-4 shadow-[0_8px_24px_rgba(26,40,31,0.035)] sm:p-5">
      <div className="mb-4 flex items-start justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-[#ecf8f0] text-emerald-700"><Layers3 className="h-4 w-4" /></span>
          <div>
            <p className="text-sm font-semibold tracking-[-0.02em] text-[#28332c]">Your saved decks</p>
            <p className="text-xs text-[#7b857e]">Progress is saved on this device.</p>
          </div>
        </div>
        <span className="rounded-full bg-[#f5f5f1] px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.1em] text-[#7b7b74]">{savedDecks.length} {savedDecks.length === 1 ? 'deck' : 'decks'}</span>
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        {savedDecks.map((deck, index) => {
          const dueCount = getDueCount(deck)
          const reviewedCount = Object.keys(deck.progress).length
          return (
            <motion.div
              key={deck.id}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.04, duration: 0.25 }}
              className="group flex min-w-0 items-center gap-3 rounded-xl border border-[#ecece7] bg-[#fcfcfa] p-3 transition-all hover:border-[#cfe4d7] hover:bg-emerald-50/30 hover:shadow-sm"
            >
              <button type="button" onClick={() => onOpenDeck(deck)} className="flex min-w-0 flex-1 items-center gap-3 text-left">
                <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-white text-emerald-700 shadow-sm"><CalendarClock className="h-4 w-4" /></span>
                <span className="min-w-0">
                  <span className="block truncate text-xs font-semibold text-[#343c36]">{deck.title}</span>
                  <span className="mt-0.5 flex items-center gap-1 text-[11px] text-[#7b857e]">
                    <span className={dueCount ? 'font-semibold text-emerald-700' : ''}>{dueCount ? `${dueCount} due` : 'All caught up'}</span>
                    <span>·</span>
                    <span>{reviewedCount}/{deck.cards.length} reviewed</span>
                  </span>
                </span>
                <ArrowUpRight className="h-4 w-4 shrink-0 text-[#a4aba5] transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5 group-hover:text-emerald-700" />
              </button>
              <button
                type="button"
                onClick={() => void handleDelete(deck)}
                aria-label={`Delete ${deck.title}`}
                className="rounded-lg p-1.5 text-[#b1b5af] opacity-0 transition-all hover:bg-red-50 hover:text-red-600 focus:opacity-100 group-hover:opacity-100"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </motion.div>
          )
        })}
      </div>
    </section>
  )
}
