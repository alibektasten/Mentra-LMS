'use client'

import { useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { Check, ChevronLeft, ChevronRight, RotateCcw, Sparkles } from 'lucide-react'
import { FlashcardDTO } from '@/lib/api/ai'
import { reviewFlashcardDeck } from '@/lib/api/flashcard-decks'
import { ReviewRating, useFlashcardDeckStore } from '@/lib/stores/flashcard-decks.store'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils/cn'

interface FlashcardDeckProps {
  cards: FlashcardDTO[]
  deckId?: string
  initialIndex?: number
}

const cardSlide = {
  enter: (moveDirection: number) => ({ opacity: 0, x: moveDirection * 24, scale: 0.985 }),
  center: { opacity: 1, x: 0, scale: 1, transition: { duration: 0.22, ease: [0.22, 1, 0.36, 1] } },
  exit: (moveDirection: number) => ({ opacity: 0, x: moveDirection * -20, scale: 0.99, transition: { duration: 0.18 } }),
}

export function FlashcardDeck({ cards, deckId, initialIndex = 0 }: FlashcardDeckProps) {
  const [index, setIndex] = useState(initialIndex)
  const [flipped, setFlipped] = useState(false)
  const [direction, setDirection] = useState(1)
  const [mastered, setMastered] = useState<number[]>([])
  const recordReview = useFlashcardDeckStore((state) => state.recordReview)
  const saveSyncedDeck = useFlashcardDeckStore((state) => state.saveSyncedDeck)
  const savedDeck = useFlashcardDeckStore((state) => deckId ? state.decks.find((deck) => deck.id === deckId) : undefined)

  if (cards.length === 0) return null

  const card = cards[index]
  const total = cards.length
  const progress = ((index + 1) / total) * 100
  const isMastered = mastered.includes(index)
  const reviewedIndexes = new Set(Object.keys(savedDeck?.progress ?? {}).map(Number))

  const selectCard = (nextIndex: number, nextDirection = 1) => {
    setDirection(nextDirection)
    setFlipped(false)
    setIndex(nextIndex)
  }

  const goNext = () => selectCard((index + 1) % total, 1)
  const goPrev = () => selectCard((index - 1 + total) % total, -1)

  const reset = () => {
    setIndex(0)
    setFlipped(false)
    setDirection(-1)
    setMastered([])
  }

  const markMastered = () => {
    setMastered((current) => current.includes(index) ? current.filter((item) => item !== index) : [...current, index])
  }

  const rateReview = (rating: ReviewRating) => {
    if (!deckId) return
    recordReview(deckId, index, rating)
    if (savedDeck?.synced) {
      void reviewFlashcardDeck(deckId, index, rating)
        .then((deck) => saveSyncedDeck(savedDeck.ownerId, deck))
        .catch(() => undefined)
    }
    setMastered((current) => {
      if (rating === 'again') return current.filter((item) => item !== index)
      return current.includes(index) ? current : [...current, index]
    })
    goNext()
  }

  return (
    <section className="overflow-hidden rounded-2xl border border-[#dcebe3] bg-[radial-gradient(circle_at_top_right,_rgba(167,243,208,0.55),_transparent_34%),linear-gradient(145deg,_#fbfffc,_#f5faf7)] p-4 sm:p-6">
      <div className="mb-5 flex items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-emerald-600 text-white shadow-sm">
              <Sparkles className="h-3.5 w-3.5" />
            </span>
            <p className="text-sm font-semibold tracking-[-0.02em] text-[#1b2a22]">Study deck</p>
          </div>
          <p className="mt-1 text-xs text-[#718078]">
            Card {index + 1} of {total} · {deckId ? `${reviewedIndexes.size} reviewed` : `${mastered.length} mastered this session`}
          </p>
        </div>
        <button onClick={reset} className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-xs font-semibold text-[#6c766f] transition-colors hover:bg-white hover:text-[#1b2a22]" type="button">
          <RotateCcw className="h-3.5 w-3.5" /> Reset
        </button>
      </div>

      <div className="mb-6 h-1.5 overflow-hidden rounded-full bg-[#dcebe3]" aria-label={`${Math.round(progress)}% through deck`}>
        <motion.div
          className="h-full rounded-full bg-gradient-to-r from-emerald-500 to-teal-400"
          animate={{ width: `${progress}%` }}
          transition={{ type: 'spring', stiffness: 180, damping: 24 }}
        />
      </div>

      <div className="relative px-1 pb-4 pt-1 sm:px-4">
        <div className="pointer-events-none absolute inset-x-5 bottom-0 top-5 rotate-[-1.5deg] rounded-2xl border border-emerald-100 bg-white/60 shadow-sm sm:inset-x-8" />
        <div className="pointer-events-none absolute inset-x-5 bottom-1 top-4 rotate-[1.5deg] rounded-2xl border border-emerald-100/80 bg-white/75 sm:inset-x-8" />

        <AnimatePresence mode="wait" custom={direction}>
          <motion.div
            key={index}
            custom={direction}
            variants={cardSlide}
            initial="enter"
            animate="center"
            exit="exit"
            className="relative"
          >
            <button
              type="button"
              onClick={() => setFlipped((value) => !value)}
              onKeyDown={(event) => {
                if (event.key === 'ArrowRight') goNext()
                if (event.key === 'ArrowLeft') goPrev()
                if (event.key === ' ') {
                  event.preventDefault()
                  setFlipped((value) => !value)
                }
              }}
              className="group block w-full rounded-2xl text-left outline-none [perspective:1200px] focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-4"
              aria-label={flipped ? `Definition: ${card.definition}. Press space to show term.` : `Term: ${card.term}. Press space to reveal definition.`}
            >
              <motion.div
                animate={{ rotateY: flipped ? 180 : 0 }}
                transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
                className="relative min-h-[260px] [transform-style:preserve-3d] sm:min-h-[290px]"
              >
                <div className="absolute inset-0 flex flex-col justify-between overflow-hidden rounded-2xl border border-[#d7e8de] bg-white p-5 shadow-[0_18px_40px_rgba(17,68,44,0.12)] [backface-visibility:hidden] sm:p-7">
                  <div className="flex items-start justify-between gap-4">
                    <span className="rounded-full bg-emerald-50 px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.14em] text-emerald-700">Term</span>
                    <span className="text-[10px] font-medium text-[#a1ada5]">Tap to reveal</span>
                  </div>
                  <p className="mx-auto max-w-xl text-center text-xl font-semibold leading-snug tracking-[-0.03em] text-[#1b2a22] sm:text-2xl">{card.term}</p>
                  <p className="text-center text-xs text-[#869189]">Space to flip · ← → to move</p>
                </div>

                <div className="absolute inset-0 flex flex-col justify-between overflow-hidden rounded-2xl border border-emerald-600 bg-gradient-to-br from-emerald-600 via-emerald-600 to-teal-600 p-5 text-white shadow-[0_18px_40px_rgba(5,120,93,0.24)] [backface-visibility:hidden] [transform:rotateY(180deg)] sm:p-7">
                  <div className="flex items-start justify-between gap-4">
                    <span className="rounded-full bg-white/15 px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.14em] text-emerald-50">Definition</span>
                    <span className="text-[10px] font-medium text-emerald-100">Tap to flip back</span>
                  </div>
                  <p className="mx-auto max-w-xl text-center text-base font-medium leading-relaxed text-white sm:text-lg">{card.definition}</p>
                  <p className="text-center text-xs text-emerald-100">Keep going—you’re building recall.</p>
                </div>
              </motion.div>
            </button>
          </motion.div>
        </AnimatePresence>
      </div>

      <div className="mt-2 flex flex-col-reverse items-center justify-between gap-3 sm:flex-row">
        <div className="flex items-center gap-2">
          <Button variant="outline" size="icon" onClick={goPrev} className="h-9 w-9 rounded-xl bg-white" aria-label="Previous card">
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <div className="flex max-w-32 gap-1 overflow-hidden" aria-label="Card progress">
            {cards.map((_, itemIndex) => (
              <button
                key={itemIndex}
                type="button"
                onClick={() => selectCard(itemIndex, itemIndex > index ? 1 : -1)}
                aria-label={`Go to card ${itemIndex + 1}`}
                className={cn(
                  'h-1.5 shrink-0 rounded-full transition-all duration-300',
                  itemIndex === index ? 'w-5 bg-emerald-600' : (deckId ? reviewedIndexes.has(itemIndex) : mastered.includes(itemIndex)) ? 'w-1.5 bg-emerald-300' : 'w-1.5 bg-[#cbd8d0] hover:bg-emerald-300'
                )}
              />
            ))}
          </div>
          <Button variant="outline" size="icon" onClick={goNext} className="h-9 w-9 rounded-xl bg-white" aria-label="Next card">
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>

        {deckId ? (
          <div className="flex items-center gap-1 rounded-xl border border-[#dcebe3] bg-white p-1 shadow-sm">
            <span className="hidden px-1.5 text-[10px] font-semibold uppercase tracking-[0.1em] text-[#839087] sm:block">Recall</span>
            <button type="button" onClick={() => rateReview('again')} className="rounded-lg px-2.5 py-1.5 text-xs font-semibold text-[#a24c3a] transition-colors hover:bg-[#fff0ec]">Again</button>
            <button type="button" onClick={() => rateReview('good')} className="rounded-lg px-2.5 py-1.5 text-xs font-semibold text-[#147657] transition-colors hover:bg-emerald-50">Good</button>
            <button type="button" onClick={() => rateReview('easy')} className="rounded-lg bg-emerald-600 px-2.5 py-1.5 text-xs font-semibold text-white shadow-sm transition-colors hover:bg-emerald-700">Easy</button>
          </div>
        ) : (
          <Button
            variant={isMastered ? 'secondary' : 'default'}
            size="sm"
            onClick={markMastered}
            className={cn('min-w-32 rounded-xl', isMastered && 'bg-emerald-100 text-emerald-800 hover:bg-emerald-200')}
          >
            <Check className="h-3.5 w-3.5" />
            {isMastered ? 'Mastered' : 'Mark mastered'}
          </Button>
        )}
      </div>
    </section>
  )
}
